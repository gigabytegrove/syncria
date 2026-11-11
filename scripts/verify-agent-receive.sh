#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
go test -race -count=3 ./agentcontrol ./agentclient ./agentpolicy ./agenttransfer ./agentfiles ./agentpeer ./agentpublish ./cmd/syncria-agent
go vet ./agentcontrol ./agentclient ./agentpolicy ./agenttransfer ./agentfiles ./agentpeer ./agentpublish ./cmd/syncria-agent
go test -cover ./agentpublish ./cmd/syncria-agent
