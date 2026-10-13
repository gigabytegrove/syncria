#!/usr/bin/env bash
set -Eeuo pipefail
# Free publisher. Runs on dockeradm; no GitHub Actions or billing.
cd "$(dirname "$0")/.."
version="$1"
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo 'Usage: publish-release.sh vMAJOR.MINOR.PATCH' >&2; exit 2;; esac
tmp="$(mktemp -d)"
auth="$(mktemp -d)"
trap 'rm -rf "$tmp" "$auth"; unset GH_TOKEN' EXIT

# Release publishing is an explicit, one-time developer action.
# Never repeatedly retry a stale credential (HTTP 401) as a host service.
if test -z "${GH_TOKEN:-}" && command -v gh >/dev/null 2>&1; then
  GH_TOKEN="$(gh auth token 2>/dev/null || true)"
fi
for attempt in 1 2 3; do
  if test -z "${GH_TOKEN:-}"; then
    test -r /dev/tty || { echo "Interactive GitHub token required." >&2; exit 1; }
    read -rsp "GitHub publishing token (Contents read/write; packages for GHCR): " GH_TOKEN </dev/tty
    echo >/dev/tty
  fi
  http_code="$(curl -sS -o "$tmp/whoami.json" -w '%{http_code}' \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "Accept: application/vnd.github+json" \
    https://api.github.com/user)"
  if test "$http_code" = 200; then break; fi
  echo "GitHub authentication rejected (HTTP $http_code). The previous publisher token is not valid." >&2
  GH_TOKEN=""
done
test -n "${GH_TOKEN:-}" || { echo "No valid publishing token provided" >&2; exit 1; }
export GH_TOKEN
actual_login="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1])).get("login",""))' "$tmp/whoami.json")"
if test -z "${GITHUB_USER:-}"; then GITHUB_USER="$actual_login"; fi
if test "$GITHUB_USER" != "$actual_login"; then
  echo "GITHUB_USER=$GITHUB_USER does not match authenticated token owner $actual_login" >&2
  exit 1
fi
echo "Authenticated GitHub publishing user: $actual_login"
mkdir -p "$tmp/dist"
# Test in a throwaway *writable* source copy so Go can resolve direct
# protocol adapter dependencies without changing the maintainer checkout.
# No GitHub Actions runners, persistent services or runtime updaters are used.
mkdir -p "$tmp/go-cache"
docker run --rm \
  -v "$PWD:/source:ro" -v "$tmp/go-cache:/go/pkg/mod" \
  -w /src golang:1.25-alpine sh -ec \
  'cp -a /source /src; cd /src; go mod tidy; go test ./...; go vet ./...'
for target in linux/amd64 linux/arm64 windows/amd64 darwin/amd64; do
  os="${target%/*}"; arch="${target#*/}"; ext=""
  if test "$os" = windows; then ext=".exe"; fi
  docker run --rm \
    -v "$PWD:/source:ro" -v "$tmp/go-cache:/go/pkg/mod" \
    -v "$tmp/dist:/dist" -w /src \
    -e GOOS="$os" -e GOARCH="$arch" -e CGO_ENABLED=0 \
    -e REL_VERSION="$version" -e EXT="$ext" golang:1.25-alpine \
    sh -ec 'cp -a /source /src; cd /src; go mod tidy; go build -trimpath -ldflags="-s -w -X main.appVersion=$REL_VERSION" -o "/dist/syncria-$GOOS-$GOARCH$EXT" .'
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
# The tag lookup does NOT expose draft releases. Search the authenticated
# release collection first, so an interrupted publish can resume.
found=0
for page in 1 2 3 4 5; do
  code="$(curl -sS -o "$tmp/releases-page.json" -w '%{http_code}' -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" "https://api.github.com/repos/gigabytegrove/syncria/releases?per_page=100&page=$page")"
  if test "$code" != 200; then cat "$tmp/releases-page.json" >&2; exit 1; fi
  if python3 - "$tmp/releases-page.json" "$version" "$tmp/release.json" <<'PY'
import json,sys
items=json.load(open(sys.argv[1]))
match=next((item for item in items if item.get("tag_name")==sys.argv[2]),None)
if match is None:sys.exit(1)
with open(sys.argv[3],"w") as f:json.dump(match,f)
PY
  then found=1; break; fi
  count="$(python3 -c 'import json,sys;print(len(json.load(open(sys.argv[1]))))' "$tmp/releases-page.json")"
  test "$count" = 100 || break
done
if test "$found" = 0; then
  code="$(curl -sS -o "$tmp/release.json" -w '%{http_code}' -X POST -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" -H "Content-Type: application/json" -d "{\"tag_name\":\"$version\",\"name\":\"Syncria $version\",\"draft\":true}" https://api.github.com/repos/gigabytegrove/syncria/releases)"
  if test "$code" != 201; then cat "$tmp/release.json" >&2; exit 1; fi
else
  echo "Reusing existing release record for $version"
fi
if python3 - "$tmp/release.json" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
sys.exit(0 if r.get("draft") else 1)
PY
then
  # Resume drafts; uploading each missing asset is safe after interruptions.
  release_id="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["id"])' "$tmp/release.json")"
  upload_url="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["upload_url"].split("{")[0])' "$tmp/release.json")"
  for file in "$tmp"/dist/*; do
    name="$(basename "$file")"
    if python3 - "$tmp/release.json" "$name" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
sys.exit(0 if any(a["name"]==sys.argv[2] and a.get("state")=="uploaded" for a in r.get("assets",[])) else 1)
PY
    then continue; fi
    code="$(curl -sS -o "$tmp/asset.json" -w '%{http_code}' -X POST -H "Authorization: Bearer $GH_TOKEN" -H "Content-Type: application/octet-stream" --data-binary "@$file" "$upload_url?name=$name")"
    if test "$code" != 201; then cat "$tmp/asset.json" >&2; exit 1; fi
  done
  code="$(curl -sS -o "$tmp/publish.json" -w '%{http_code}' -X PATCH -H "Authorization: Bearer $GH_TOKEN" -H "Content-Type: application/json" -d '{"draft":false}' "https://api.github.com/repos/gigabytegrove/syncria/releases/$release_id")"
  if test "$code" != 200; then
    cat "$tmp/publish.json" >&2
    echo "Unable to make release $version public" >&2
    exit 1
  fi
  echo "Published GitHub release $version"
else
  # A completed GitHub Release must not be retagged or overwritten.
  # Continue to the optional GHCR push, which may have failed previously.
  echo "GitHub release $version already published; verifying and continuing to container publish."
fi
# Verify the newly published release is anonymously visible to all app updaters.
public_status="$(curl -sS -o "$tmp/public-check.json" -w '%{http_code}' \
  -H 'Accept: application/vnd.github+json' \
  "https://api.github.com/repos/gigabytegrove/syncria/releases/tags/$version")"
if test "$public_status" != 200; then
  echo "ERROR: GitHub release is not publicly accessible (HTTP $public_status)." >&2
  exit 1
fi
python3 - "$tmp/public-check.json" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
required={"syncria-linux-amd64","syncria-linux-arm64","syncria-windows-amd64.exe","syncria-darwin-amd64","SHA256SUMS"}
assets={x["name"] for x in r.get("assets",[]) if x.get("state")=="uploaded" and x.get("size",0)>0}
if r.get("draft") or r.get("prerelease") or not required.issubset(assets):
    raise SystemExit("Release not installable: "+", ".join(sorted(required-assets)))
print("Verified: public release and SHA256SUMS are available to Syncria's built-in updater.")
PY
# Optional GHCR publishing must never block direct GitHub binary updates.
if test "${SYNCRIA_PUBLISH_GHCR:-0}" = 1; then
  export DOCKER_CONFIG="$auth"
  trap 'rm -rf "$tmp" "$auth"; unset GH_TOKEN' EXIT
  printf %s "$GH_TOKEN" | docker login ghcr.io -u "$GITHUB_USER" --password-stdin
  builder="syncria-release-${BASHPID}"
  docker buildx create --name "$builder" --driver docker-container >/dev/null
  trap 'docker buildx rm "$builder" >/dev/null 2>&1 || true; rm -rf "$tmp" "$auth"; unset GH_TOKEN' EXIT
  docker buildx build --builder "$builder" --platform linux/amd64,linux/arm64 --build-arg "APP_VERSION=$version" -t "ghcr.io/gigabytegrove/syncria:$version" -t ghcr.io/gigabytegrove/syncria:latest --push .

fi

