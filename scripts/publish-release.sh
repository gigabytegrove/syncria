#!/usr/bin/env bash
set -Eeuo pipefail
# Free publisher. Runs on dockeradm; no GitHub Actions or billing.
cd "$(dirname "$0")/.."
version="$1"
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo 'Usage: publish-release.sh vMAJOR.MINOR.PATCH' >&2; exit 2;; esac
: "${GITHUB_USER:?Set GITHUB_USER to your GitHub username}"
if test -z "${GH_TOKEN:-}"; then read -rsp "Classic GitHub token (public_repo, write:packages): " GH_TOKEN; echo; fi
export GH_TOKEN
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"; unset GH_TOKEN' EXIT
mkdir -p "$tmp/dist"
# Run tests once for the native Linux builder. Cross-compiled tests cannot execute on that host.
docker run --rm --network none -v "$PWD:/src:ro" -w /src golang:1.23-alpine sh -ec 'go test ./... && go vet ./...'
for target in linux/amd64 linux/arm64 windows/amd64 darwin/amd64; do
  os="${target%/*}"; arch="${target#*/}"; ext=""
  if test "$os" = windows; then ext=".exe"; fi
  docker run --rm --network none -v "$PWD:/src:ro" -v "$tmp/dist:/dist" -w /src -e GOOS="$os" -e GOARCH="$arch" -e CGO_ENABLED=0 -e REL_VERSION="$version" -e EXT="$ext" golang:1.23-alpine sh -ec 'go build -trimpath -ldflags="-s -w -X main.appVersion=$REL_VERSION" -o "/dist/syncria-$GOOS-$GOARCH$EXT" .'
done
(cd "$tmp/dist" && sha256sum syncria-* > SHA256SUMS)
gitsha="$(git rev-parse HEAD)"
# Repeatable after a failed upload: never retag a different commit.
tag_status="$(curl -sS -o "$tmp/tag.json" -w '%{http_code}' -H "Authorization: Bearer $GH_TOKEN" "https://api.github.com/repos/gigabytegrove/syncria/git/ref/tags/$version")"
if test "$tag_status" = 200; then
  existing_sha="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["object"]["sha"])' "$tmp/tag.json")"
  test "$existing_sha" = "$gitsha" || { echo "Release tag exists at another commit" >&2; exit 1; }
elif test "$tag_status" = 404; then
  code="$(curl -sS -o "$tmp/tag.json" -w '%{http_code}' -X POST -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" -H "Content-Type: application/json" -d "{\"ref\":\"refs/tags/$version\",\"sha\":\"$gitsha\"}" https://api.github.com/repos/gigabytegrove/syncria/git/refs)"
  if test "$code" != 201; then cat "$tmp/tag.json" >&2; exit 1; fi
else
  cat "$tmp/tag.json" >&2; echo "Cannot inspect release tag" >&2; exit 1
fi
release_status="$(curl -sS -o "$tmp/release.json" -w '%{http_code}' -H "Authorization: Bearer $GH_TOKEN" "https://api.github.com/repos/gigabytegrove/syncria/releases/tags/$version")"
if test "$release_status" = 404; then
  code="$(curl -sS -o "$tmp/release.json" -w '%{http_code}' -X POST -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" -H "Content-Type: application/json" -d "{\"tag_name\":\"$version\",\"name\":\"Syncria $version\",\"draft\":true}" https://api.github.com/repos/gigabytegrove/syncria/releases)"
  if test "$code" != 201; then cat "$tmp/release.json" >&2; exit 1; fi
elif test "$release_status" != 200; then
  cat "$tmp/release.json" >&2; echo "Cannot inspect existing release" >&2; exit 1
fi
python3 - "$tmp/release.json" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
if not r["draft"]:
    raise SystemExit("Release already published. Bump VERSION instead of overwriting it.")
PY
release_id="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["id"])' "$tmp/release.json")"
upload_url="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["upload_url"].split("{")[0])' "$tmp/release.json")"
for file in "$tmp"/dist/*; do
  name="$(basename "$file")"
  if python3 - "$tmp/release.json" "$name" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
sys.exit(0 if any(a["name"]==sys.argv[2] for a in r.get("assets",[])) else 1)
PY
  then continue; fi
  code="$(curl -sS -o "$tmp/asset.json" -w '%{http_code}' -X POST -H "Authorization: Bearer $GH_TOKEN" -H "Content-Type: application/octet-stream" --data-binary "@$file" "$upload_url?name=$name")"
  if test "$code" != 201; then cat "$tmp/asset.json" >&2; exit 1; fi
done
auth="$(mktemp -d)"
trap 'rm -rf "$tmp" "$auth"; unset GH_TOKEN' EXIT
printf %s "$GH_TOKEN" | docker --config "$auth" login ghcr.io -u "$GITHUB_USER" --password-stdin
builder="syncria-release-${BASHPID}"
docker buildx create --name "$builder" --driver docker-container >/dev/null
trap 'docker buildx rm "$builder" >/dev/null 2>&1 || true; rm -rf "$tmp" "$auth"; unset GH_TOKEN' EXIT
docker --config "$auth" buildx build --builder "$builder" --platform linux/amd64,linux/arm64 --build-arg "APP_VERSION=$version" -t "ghcr.io/gigabytegrove/syncria:$version" -t ghcr.io/gigabytegrove/syncria:latest --push .
code="$(curl -sS -o "$tmp/publish.json" -w '%{http_code}' -X PATCH -H "Authorization: Bearer $GH_TOKEN" -H "Content-Type: application/json" -d '{"draft":false}' "https://api.github.com/repos/gigabytegrove/syncria/releases/$release_id")"
if test "$code" != 200; then cat "$tmp/publish.json" >&2; exit 1; fi
echo "Published $version"
