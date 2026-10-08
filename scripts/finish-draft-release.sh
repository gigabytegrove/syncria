#!/usr/bin/env bash
set -Eeuo pipefail
cd "$(dirname "$0")/.."
version="${1:-$(tr -d '\r\n ' < VERSION)}"
case "$version" in v[0-9]*.[0-9]*.[0-9]*) ;; *) echo "Invalid version" >&2; exit 2;; esac
: "${GH_TOKEN:?GH_TOKEN must be supplied in the trusted publisher environment}"
temp="$(mktemp -d)"; trap 'rm -rf "$temp"' EXIT
found=0
for page in 1 2 3 4 5; do
 status="$(curl -sS -o "$temp/releases.json" -w '%{http_code}' -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" "https://api.github.com/repos/gigabytegrove/syncria/releases?per_page=100&page=$page")"
 test "$status" = 200 || { cat "$temp/releases.json" >&2; exit 1; }
 if python3 - "$temp/releases.json" "$version" "$temp/release.json" <<'PY'
import json,sys
items=json.load(open(sys.argv[1]))
release=next((r for r in items if r.get("tag_name")==sys.argv[2]),None)
if release is None:sys.exit(1)
with open(sys.argv[3],"w") as f:json.dump(release,f)
PY
 then found=1; break; fi
 count="$(python3 -c 'import json,sys;print(len(json.load(open(sys.argv[1]))))' "$temp/releases.json")"
 test "$count" = 100 || break
done
test "$found" = 1 || { echo "No existing release found for $version" >&2; exit 10; }
python3 - "$temp/release.json" <<'PY'
import json,sys
r=json.load(open(sys.argv[1]))
if not r.get("draft"):raise SystemExit("Release already published, no action required")
required={"syncria-linux-amd64","syncria-linux-arm64","syncria-windows-amd64.exe","syncria-darwin-amd64","SHA256SUMS"}
assets={a["name"] for a in r.get("assets",[]) if a.get("state")=="uploaded" and a.get("size",0)>0}
missing=required-assets
if missing:raise SystemExit("Release assets not complete: "+", ".join(sorted(missing)))
print("All five release assets uploaded")
PY
release_id="$(python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["id"])' "$temp/release.json")"
status="$(curl -sS -o "$temp/updated.json" -w '%{http_code}' -X PATCH -H "Authorization: Bearer $GH_TOKEN" -H "Accept: application/vnd.github+json" -H "Content-Type: application/json" -d '{"draft":false}' "https://api.github.com/repos/gigabytegrove/syncria/releases/$release_id")"
test "$status" = 200 || { cat "$temp/updated.json" >&2; exit 1; }
echo "Published Syncria $version after verifying assets"
