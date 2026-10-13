#!/usr/bin/env bash
set -Eeuo pipefail

# Strictly non-deploying verification of the Syncria v0.5.0 source branch.
# No NAS mounts, no credentials, no Docker socket inside the test container,
# no GitHub release/tag/GHCR publishing, and no user state modifications.
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

for binary in docker git; do
  command -v "$binary" >/dev/null 2>&1 || {
    echo "Missing prerequisite: $binary" >&2
    exit 1
  }
done

echo "== Syncria v0.5.0 isolated source verification =="
git clone --quiet --depth 1 --branch syncria-v0.5.0-work \
  https://github.com/gigabytegrove/syncria.git "$WORK/source"
git -C "$WORK/source" log -1 --format='Source commit: %H %s'
echo "== Go 1.25 dependency resolution, tests, vet and compilation =="
docker run --rm \
  --security-opt no-new-privileges \
  --cap-drop ALL \
  -v "$WORK/source:/source:ro" \
  golang:1.25-alpine \
  sh -eu -c '
    mkdir -p /work
    cp -a /source/. /work/
    cd /work
    go mod tidy
    go test -count=1 ./...
    go vet ./...
    CGO_ENABLED=0 go build -trimpath -o /tmp/syncria-test .
    /tmp/syncria-test -h >/dev/null
  '

echo "PASS: All Go tests, vet, and executable startup checks completed."
echo "No release was published and no NAS/cloud files were accessed."
