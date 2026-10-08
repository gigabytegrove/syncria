#!/usr/bin/env bash
# Publish a multi-architecture Syncria image from a machine with Docker Buildx.
# Publishing requires registry push credentials on the publisher's machine.
# Users pulling an image made PUBLIC in GHCR require no authentication.
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
command -v docker >/dev/null || { echo "Docker is required on the publishing machine" >&2; exit 1; }
docker buildx version >/dev/null || { echo "Docker Buildx is required" >&2; exit 1; }
IMAGE="ghcr.io/gigabytegrove/syncria"
TAG="${1:-latest}"
[[ "$TAG" =~ ^[a-zA-Z0-9][a-zA-Z0-9_.-]*$ ]] || { echo "Invalid image tag" >&2; exit 1; }
echo "Publishing $IMAGE:$TAG for linux/amd64 and linux/arm64"
docker buildx build --platform linux/amd64,linux/arm64 \
  --tag "$IMAGE:$TAG" --push .
echo "Push completed. Confirm GHCR package visibility is PUBLIC before distributing the Compose file."
