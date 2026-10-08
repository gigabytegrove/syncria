#!/usr/bin/env bash
set -Eeuo pipefail
# One-time self-hosted, no-card publishing service installer (Linux/systemd).
[ "$(id -u)" -eq 0 ] || { echo "Run as root on the dedicated Docker publisher host" >&2; exit 1; }
for cmd in docker git curl systemctl flock python3; do command -v "$cmd" >/dev/null || { echo "Missing $cmd" >&2; exit 1; }; done
repo="$(cd "$(dirname "$0")/.." && pwd)"
test -d "$repo/.git" || { echo "Run from your Syncria git checkout" >&2; exit 1; }
read -rp "GitHub publisher username: " github_user
read -rsp "Classic token (public_repo and write:packages): " gh_token
echo
[ -n "$github_user" ] && [ -n "$gh_token" ] || { echo "Username and token required" >&2; exit 1; }
umask 077
env_file="/etc/syncria-publisher.env"
printf 'GITHUB_USER=%s\nGH_TOKEN=%s\n' "$github_user" "$gh_token" > "$env_file"
chmod 600 "$env_file"
unset gh_token
cat > /etc/systemd/system/syncria-publisher.service <<EOF
[Unit]
Description=Syncria self-hosted release publisher
After=network-online.target docker.service
Wants=network-online.target
Requires=docker.service

[Service]
Type=oneshot
WorkingDirectory=$repo
EnvironmentFile=$env_file
ExecStart=/bin/bash $repo/scripts/publish-when-version-changes.sh
TimeoutStartSec=3600
UMask=0077
EOF
cat > /etc/systemd/system/syncria-publisher.timer <<'EOF'
[Unit]
Description=Check Syncria VERSION for new releases

[Timer]
OnBootSec=2min
OnUnitInactiveSec=15min
Persistent=true
Unit=syncria-publisher.service

[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now syncria-publisher.timer
echo "Publisher installed. Trigger immediately: systemctl start syncria-publisher"
echo "Logs: journalctl -u syncria-publisher -n 100 --no-pager"
