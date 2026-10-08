# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
ARG TARGETOS=linux
ARG TARGETARCH=amd64
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/syncria .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata && mkdir -p /data /sync
COPY --from=build /out/syncria /usr/local/bin/syncria
COPY launcher.sh /usr/local/bin/syncria-launcher
RUN chmod 755 /usr/local/bin/syncria-launcher
WORKDIR /data
EXPOSE 9764
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/syncria-launcher"]
CMD ["-data", "/data", "-listen", "0.0.0.0:9764"]
