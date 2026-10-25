#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
if [[ ! -f go.mod || ! -f Dockerfile || ! -f main.go ]]; then
  echo 'ERROR: full Syncria source checkout is required, not a source subset' >&2
  exit 2
fi
if ! command -v go >/dev/null 2>&1; then echo 'ERROR: Go 1.25+ is required' >&2; exit 2; fi
version="$(go env GOVERSION)"
if [[ ! "$version" =~ ^go1\.([0-9]+)\.([0-9]+) ]] || (( ${BASH_REMATCH[1]} < 25 )); then
  echo "ERROR: Go 1.25 or newer required; found $version" >&2; exit 2
fi
if ! command -v git >/dev/null 2>&1; then echo 'ERROR: git required' >&2;exit 2; fi
if [[ -n "$(git status --porcelain)" ]]; then
  echo 'ERROR: refuse to certify a dirty working tree' >&2;exit 2
fi
mkdir -p verification
{
  printf 'revision=%s\n' "$(git rev-parse HEAD)"
  printf 'go_version=%s\n' "$version"
  printf 'started=%s\n' "$(date -u +%FT%TZ)"
} > verification/baseline.txt
printf '\n== Full package tests and static analysis ==\n'
GOTOOLCHAIN=local go test -count=1 ./... | tee verification/go-test.log
GOTOOLCHAIN=local go vet ./... | tee verification/go-vet.log
printf '\n== Race regressions across controller + agent packages ==\n'
GOTOOLCHAIN=local go test -race -count=2 ./... | tee verification/go-race.log
printf '\n== Cross compile native agents ==\n'
for target in 'linux amd64' 'linux arm64' 'linux arm' 'windows amd64' 'windows arm64' 'darwin amd64' 'darwin arm64';do
  read -r os arch <<< "$target"
  suffix=''; [[ "$os" == windows ]] && suffix='.exe'
  out="verification/syncria-agent-${os}-${arch}${suffix}"
  CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" GOTOOLCHAIN=local go build -trimpath -o "$out" ./cmd/syncria-agent
  echo "PASS $os/$arch"
done
(cd verification && sha256sum syncria-agent-* > SHA256SUMS)
cat > verification/result.txt <<RESULT
source_revision=$(git rev-parse HEAD)
full_test=passed
full_vet=passed
full_race=passed
agent_crossbuilds=7
NOTE=These tests do not validate real NAS/OS/browser upgrades or signed release assets.
RESULT
printf '\nCompleted source-only verification: %s\n' "$(git rev-parse --short HEAD)"
