#!/usr/bin/env bash
# Fail closed when a Go package is absent from Syncria's Docker build stage.
set -Eeuo pipefail
cd "$(dirname "$0")/.."
[[ -f Dockerfile && -f go.mod && -f compose.alpha.yaml ]] || { echo "Not a Syncria checkout" >&2; exit 2; }
command -v git >/dev/null || { echo "git is required" >&2; exit 2; }
found=0
while IFS= read -r top; do
  [[ -n "$top" ]] || continue
  if ! grep -Fqx "COPY $top/ ./$top/" Dockerfile; then
    printf 'FAIL: missing Go source directory in Dockerfile: %s/\n' "$top" >&2
    found=1
  fi
done < <(git ls-files '*.go' | awk -F/ 'NF>1 {print $1}' | sort -u)
if (( found )); then exit 1; fi
grep -Fq 'go test -count=1 ./... && go vet ./...' Dockerfile || { echo "FAIL: full test and vet build gate missing" >&2; exit 1; }
grep -Fq 'FROM scratch AS alpha-agent-export' Dockerfile || { echo "FAIL: native agent export target missing" >&2; exit 1; }
grep -Fq 'syncria_alpha_isolated_data' compose.alpha.yaml || { echo "FAIL: dedicated alpha volume missing" >&2; exit 1; }
grep -Fq 'syncria-alpha' compose.alpha.yaml || { echo "FAIL: dedicated alpha container missing" >&2; exit 1; }
grep -Fq 'v1.0.0-alpha.1-local' compose.alpha.yaml || { echo "FAIL: distinct alpha build version missing" >&2; exit 1; }
printf 'PASS: all tracked Go source trees are copied into Docker.\n'
printf 'PASS: full Go unit/static build gates and separate alpha data volume configured.\n'
printf 'NOTE: actual Go 1.25 compilation, browser, NAS, and OS runtime validation still required.\n'
