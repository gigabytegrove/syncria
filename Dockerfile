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
ARG APP_VERSION=v0.6.7
WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY assets/ ./assets/
COPY --from=css /src/assets/syncria.css ./assets/syncria.css
RUN go mod tidy && CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X main.appVersion=$APP_VERSION" -o /out/syncria .

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
