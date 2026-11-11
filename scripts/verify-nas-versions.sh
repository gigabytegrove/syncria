#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go version
go test -race -count=3 ./nasversion
go vet ./nasversion
go test -race ./...
go vet ./...
