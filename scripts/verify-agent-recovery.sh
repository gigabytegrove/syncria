#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
packages=(./agentversion ./agentrestore ./agentpublish ./agentpolicy ./agentfiles ./cmd/syncria-agent)
go test -race -count=3 "${packages[@]}"
go vet "${packages[@]}"
go test -cover "${packages[@]}"
