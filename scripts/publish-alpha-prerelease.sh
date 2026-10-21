#!/usr/bin/env bash
# Build and publish opt-in GitHub alpha assets. Stable Syncria ignores alpha-* tags.
set -Eeuo pipefail
umask 077
repo='gigabytegrove/syncria'
tag='alpha-v1.0.0-alpha.1'
version='v1.0.0-alpha.1'
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
fail(){ printf 'ERROR: %s\n' "$*" >&2; exit 1; }
for cmd in git docker gh sha256sum mktemp cp mv; do
  command -v "$cmd" >/dev/null || fail "$cmd is required"
done
docker info >/dev/null 2>&1 || fail 'Docker Engine unavailable'
docker buildx version >/dev/null 2>&1 || fail 'Docker Buildx required'
gh auth status >/dev/null 2>&1 || fail 'Sign in first: gh auth login'
[[ "$(git rev-parse --show-toplevel)" == "$root" ]] || fail 'Run within Syncria checkout'
[[ "$(git branch --show-current)" == 'feature/v2-release-foundation' ]] || fail 'Expected development branch; do not publish from main'
[[ -z "$(git status --porcelain)" ]] || fail 'Working tree must be clean'
bash scripts/verify-alpha-build-context.sh
gh repo view "$repo" --json name >/dev/null || fail 'GitHub account cannot access repository'
if gh release view "$tag" --repo "$repo" >/dev/null 2>&1; then fail "Release $tag exists"; fi
if git ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then fail "Tag $tag exists"; fi
sha="$(git rev-parse HEAD)"
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/dist"
for arch in amd64 arm64; do
  mkdir -p "$tmp/server-$arch"
  docker buildx build --progress=plain --build-arg APP_VERSION="$version" \
    --build-arg TARGETOS=linux --build-arg TARGETARCH="$arch" \
    --target alpha-server-export --output "type=local,dest=$tmp/server-$arch" .
  test -s "$tmp/server-$arch/syncria" || fail "Missing controller $arch"
  cp "$tmp/server-$arch/syncria" "$tmp/dist/syncria-linux-$arch"
done
mkdir -p "$tmp/agents"
docker buildx build --progress=plain --build-arg APP_VERSION="$version" \
  --target alpha-agent-export --output "type=local,dest=$tmp/agents" .
(cd "$tmp/agents" && sha256sum -c SHA256SUMS)
for target in linux-amd64 linux-arm64 linux-arm windows-amd64 windows-arm64 darwin-amd64 darwin-arm64; do
  extension=''
  [[ "$target" == windows-* ]] && extension='.exe'
  file="syncria-agent-$target$extension"
  test -s "$tmp/agents/$file" || fail "Missing agent $file"
  cp "$tmp/agents/$file" "$tmp/dist/$file"
done
git archive --format=zip --prefix="syncria-$version/" HEAD > "$tmp/dist/syncria-$version-source.zip"
(cd "$tmp/dist" && sha256sum syncria-linux-* syncria-agent-* "syncria-$version-source.zip" > SHA256SUMS)
(cd "$tmp/dist" && sha256sum -c SHA256SUMS)
cat > "$tmp/RELEASE-NOTES.md" <<EOF
# Syncria $version — opt-in alpha
Source revision: $sha

Unverified preproduction build. Do not install on production v0.8.5.
Test with isolated compose.alpha.yaml and disposable NAS folders.
Includes experimental central agent-to-NAS sync and retained NAS history.
Known gaps: real SMB tests, native installers, signed updates and rollback.
The alpha- tag prefix prevents the existing stable updater from selecting this release.
EOF
gh release create "$tag" "$tmp/dist/"* --repo "$repo" --target "$sha" \
  --prerelease --latest=false --title "Syncria $version (opt-in alpha)" \
  --notes-file "$tmp/RELEASE-NOTES.md"
gh release view "$tag" --repo "$repo" --json tagName,isDraft,isPrerelease,url,assets \
  --jq '{tag: .tagName, draft: .isDraft, prerelease: .isPrerelease, url: .url, asset_count: (.assets | length)}'
echo 'Alpha release published; no production instance upgraded.'
