#!/usr/bin/env bash
set -Eeuo pipefail
repo="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo"
command -v go >/dev/null 2>&1 || { echo 'Go required' >&2; exit 1; }
go test -race -count=10 ./nasversion
go vet ./nasversion
out="$(mktemp "${TMPDIR:-/tmp}/syncria-nasversion-coverage.XXXXXX")"
trap 'rm -f "$out"' EXIT
go test -coverprofile="$out" ./nasversion
go tool cover -func="$out" | tail -1
