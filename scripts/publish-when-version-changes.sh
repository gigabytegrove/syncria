#!/usr/bin/env bash
set -Eeuo pipefail
# Run from a trusted dedicated publisher checkout. This script does not
# publish on arbitrary commits: VERSION must name a new release.
cd "$(dirname "$0")/.."
exec 9>"/etc/syncria-publisher.lock"
flock -n 9 || exit 0
git fetch --quiet origin main
git merge --ff-only origin/main >/dev/null
test -f VERSION || exit 0
version="$(tr -d '\r\n ' < VERSION)"
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo "Invalid VERSION" >&2; exit 2 ;; esac
if test -z "${GH_TOKEN:-}" || test -z "${GITHUB_USER:-}"; then echo 'GH_TOKEN and GITHUB_USER are required' >&2; exit 2; fi
status_file="$(mktemp)"
trap 'rm -f "$status_file"' EXIT
status="$(curl -sS -o "$status_file" -w '%{http_code}' -H "Authorization: Bearer $GH_TOKEN" "https://api.github.com/repos/gigabytegrove/syncria/releases/tags/$version")"
if test "$status" = 200; then
  if python3 - "$status_file" <<'PY'
import json,sys
sys.exit(0 if not json.load(open(sys.argv[1]))['draft'] else 1)
PY
  then
    echo "Syncria $version GitHub binary release already published"; exit 0
  fi
  echo "Resuming incomplete draft release $version"
elif test "$status" != 404; then echo "Cannot determine release status (HTTP $status)" >&2; exit 1
fi
bash scripts/publish-release.sh "$version"
