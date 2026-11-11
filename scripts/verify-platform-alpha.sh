#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "$0")/.."
req=1.25
have="$(go version | awk '{print $3}' | sed 's/^go//')"
if ! printf '%s\n%s\n' "$req" "$have" | sort -V -C; then
  echo "Go 1.25 or newer is required; found $have" >&2
  exit 1
fi
printf '\n=== SYNCRIA ALPHA: FORMAT ===\n'
if [ -n "$(gofmt -l agent_bridge.go agent_sync_routes.go agent_control_routes.go ui_agent_sync.go agentpeer/stream.go agentpeer/stream_test.go agent_bridge_test.go agent_sync_routes_test.go)" ]; then
  echo 'Go source formatting failed' >&2
  exit 1
fi
printf '\n=== SYNCRIA ALPHA: FULL REPOSITORY RACE TESTS ===\n'
go test -race -count=2 ./...
printf '\n=== SYNCRIA ALPHA: STATIC CHECKS ===\n'
go vet ./...
printf '\n=== SYNCRIA ALPHA: CROSS BUILDS ===\n'
out="${1:-./build/alpha1}"
mkdir -p "$out"
for pair in linux/amd64 linux/arm64 linux/arm windows/amd64 windows/arm64 darwin/amd64 darwin/arm64; do
  goos=${pair%/*}; goarch=${pair#*/}
  extension=""; if [ "$goos" = windows ]; then extension=.exe; fi
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -trimpath -o "$out/syncria-agent-${goos}-${goarch}${extension}" ./cmd/syncria-agent
  printf 'PASS %s\n' "$pair"
done
(cd "$out" && sha256sum syncria-agent-* > SHA256SUMS)
printf '\nLocal builds complete. Cross-compilation is NOT real-platform runtime verification.\n'
