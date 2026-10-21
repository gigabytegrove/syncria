# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM node:22-alpine AS css
WORKDIR /src
COPY package.json tailwind.config.cjs ./
COPY ui.go ./
COPY assets/syncria.tailwind.css ./assets/syncria.tailwind.css
RUN npm install --no-audit --no-fund && npm run build:css

FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG APP_VERSION=v0.8.5
WORKDIR /src
COPY go.mod ./
COPY *.go ./
# Include the approved device-agent and controller packages. The pre-1.0
# Dockerfile copied only root Go sources and could not build these imports.
COPY agentclient/ ./agentclient/
COPY agentcontrol/ ./agentcontrol/
COPY agentfiles/ ./agentfiles/
COPY agentpeer/ ./agentpeer/
COPY agentpolicy/ ./agentpolicy/
COPY agenttransfer/ ./agenttransfer/
COPY agentpublish/ ./agentpublish/
COPY agentrestore/ ./agentrestore/
COPY agentversion/ ./agentversion/
COPY nasversion/ ./nasversion/
COPY agentupdate/ ./agentupdate/
COPY syncpolicy/ ./syncpolicy/
COPY synctopology/ ./synctopology/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY assets/ ./assets/
COPY --from=css /src/assets/syncria.css ./assets/syncria.css
RUN go mod tidy && go test -count=1 ./... && go vet ./... && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X main.appVersion=$APP_VERSION" -o /out/syncria .

# Optional alpha artifact target: build seven verified native agent programs.
# Use: docker build --target alpha-agent-export --output type=local,dest=./alpha-agent-binaries .
FROM build AS alpha-agent-build
RUN set -eu; mkdir -p /out/agents; \
    for target in linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64 darwin/amd64 darwin/arm64; do \
      target_os="${target%/*}"; target_arch="${target#*/}"; suffix=""; \
      if [ "$target_os" = windows ]; then suffix=".exe"; fi; \
      CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build -trimpath -o "/out/agents/syncria-agent-${target_os}-${target_arch}${suffix}" ./cmd/syncria-agent; \
    done; \
    cd /out/agents; sha256sum syncria-agent-* > SHA256SUMS

FROM scratch AS alpha-agent-export
COPY --from=alpha-agent-build /out/agents/ /

# Export the controller executable with tests and vet as mandatory build gates.
FROM scratch AS alpha-server-export
COPY --from=build /out/syncria /syncria

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata nfs-utils cifs-utils samba-client util-linux inotify-tools && mkdir -p /data
COPY --from=build /out/syncria /usr/local/bin/syncria
COPY launcher.sh /usr/local/bin/syncria-launcher
RUN chmod 755 /usr/local/bin/syncria-launcher
WORKDIR /data
EXPOSE 9764
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/syncria-launcher"]
CMD ["-data", "/data", "-listen", "0.0.0.0:9764"]
