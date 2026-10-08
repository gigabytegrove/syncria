# Syncria

**Pronunciation:** Sink-Re-Ah. **Default local web UI:** `http://127.0.0.1:9764` (configurable with `-listen`). The repository is temporarily named `gdsync` until its owner renames it to `syncria`. The Go module path and existing default data folder retain `gdsync` for compatibility during the transition. Restart the agent after upgrading; existing instances launched with an explicit `-listen` argument continue using that chosen port.

A lightweight, self-hosted, **experimental** Google Drive ↔ filesystem synchronization agent written in Go. Single binary, embedded browser UI, no external database, and no JavaScript build tool. Suitable for Windows and Linux (including servers with mounted NAS shares).

> **Alpha warning:** This version has not been live-tested with Google Drive accounts. Make backups before enabling synchronization. Native Google Docs/Sheets/Slides, shortcuts, shared drives, files with duplicate names, and complicated rename/move scenarios are not fully supported. The agent **stops with a visible error** instead of silently processing unsupported Google-native formats. Do not use for your only copy of important data.

## Features

- Multiple Google OAuth accounts (individual account tokens, multiple independently configured sync mappings)
- Full **My Drive** root (`root`) or individual Google Drive folder by ID
- Bidirectional regular-file/folder changes, scanning every configured interval
- Preserves state between runs, avoids overwriting initial conflicts, detects concurrent changes
- Local deletions moved to an agent recovery directory; remote deletions move to Google Drive trash
- Dashboard password, signed session cookies, OAuth CSRF state, loopback-only default listener
- No database or external process dependencies; standard library Go 1.23+

## Build and run

```sh
go build -trimpath -ldflags='-s -w' -o syncria .
./syncria -listen 127.0.0.1:9764
```

Open http://localhost:9764 and set a password of at least 12 characters.

For server / NAS use, choose a data directory you back up, mount the NAS shares in the host OS, and use:

```sh
./syncria -data /var/lib/gdsync -listen 127.0.0.1:9764
```

**Never publish the listener directly on the internet**. Put a TLS-enabled, authenticated reverse proxy in front of it if you need remote management. On Windows run `syncria.exe`; create the local directories before adding a mapping. For Windows file shares, prefer an existing UNC path with service-account permissions, rather than relying on a drive letter only available in an interactive user session.

## Offline Docker image deployment (no registry or CI required)

For hosts without access to a container registry or paid CI runners, Syncria can be distributed as a prebuilt Docker image archive. Import the provided image archive onto the Docker host:

```sh
docker load -i syncria-docker-linux-amd64.tar.gz
```

Use an image reference of `syncria:offline` and `pull_policy: never` in Docker Compose. Persist `/data`; expose an existing host storage root under `/storage`. NAS folders and Google accounts are managed in the web interface. No GitHub Actions, registry login, source checkout, or local compilation is required to **deploy** an already-built archive. The archive must be loaded onto each Docker host and is currently built for Linux AMD64 only.

This is an archive-based distribution, **not** an anonymously pullable GHCR image. A public image registry would require a publisher to push the built image, which has not been done. Archive delivery and future releases still require generating and distributing new image archives.

## Docker Compose (no build required)

The published Docker image is `ghcr.io/gigabytegrove/syncria:latest` (linux/amd64 and linux/arm64). The automated GitHub Actions image publishing workflow was removed because runner jobs are blocked by account billing. The GHCR image is not currently published. Until an image is published independently of Actions, use the offline image archive described above instead of `docker compose pull`.

Create a Compose deployment using the repository's `compose.yaml` with the following environment values:

- `SYNCRIA_BIND_IP`: host-side listener IP (default `127.0.0.1`).
- `SYNCRIA_DATA_DIR`: persistent state, OAuth tokens, recovery data, updater runtime (default `./data`).
- `SYNCRIA_STORAGE_ROOT`: one **existing** host directory exposing locally mounted disks/NAS devices, visible inside Syncria as `/storage` (default `/mnt`).

```sh
docker compose pull
docker compose up -d
```

Do not add a `build:` block. All Google account and Google Drive sync mappings are created and changed from the Syncria dashboard. For example, the host path `/mnt/nas1/photos` appears at `/storage/nas1/photos` inside the container.

**Docker isolation limitation:** the app cannot mount arbitrary host paths or unmounted SMB/NFS shares from its web UI. The storage root must already exist on the Docker host and contain the shares. Adding genuinely new SMB/NFS network mounts from the UI requires additional mount-management features not implemented yet. Do not grant Docker socket or privileged access to work around this.

The image includes the runtime launcher and uses `/data/runtime/current` for built-in app updates, so future app updates do not require a new container image. Security updates to the base image or launcher still require an image pull/recreate.

## Built-in updater

From the **Software updates** panel, sign in as the administrator and choose **Check for updates**, **Install latest**, or **Rollback**. Releases are fetched directly from GitHub Releases, **not a container registry**. Each release must include the correct OS/CPU binary named `syncria-<os>-<arch>` (with `.exe` for Windows) and a `SHA256SUMS` asset. Downloads are checksum-validated before activation. The previous executable is kept for rollback. State, account credentials, and mappings stay in the configured data directory.

In Docker, the launcher monitors the child agent and activates staged binaries from persistent `/data/runtime/current`. Future container restarts continue using that version, so **rebuilding the Docker image is not necessary for normal app releases**. Rebuilding is still recommended when the container base OS, launcher, or bundled CA certificates change. Linux/macOS standalone installations can restart into the staged binary without modifying their original executable. Windows installs can check for new releases, but automated installation and restart are intentionally disabled in this alpha.

**Update prerequisites:** Publish at least one GitHub Release with the platform assets; GitHub Actions on version tags builds them. The updater currently uses the existing `gigabytegrove/gdsync` GitHub release endpoint; GitHub's standard repository redirect should preserve it after the planned rename to `gigabytegrove/syncria`. If GitHub still considers the repository private, provide a read-only repository token using the `SYNCRIA_GITHUB_TOKEN` environment variable; do not put this token in a public repository or shared Compose file. Only releases from a repository you trust should be installed. SHA-256 checksums confirm transfer integrity, but do not provide publisher authenticity independent of the release itself. Updater actions are blocked while the agent is actively synchronizing files.

## Google authentication (per-installation setup)

1. In Google Cloud Console, create/select a project and enable the **Google Drive API**.
2. Configure the OAuth consent screen. Add your Google accounts as test users while the app is in testing status, or complete Google's app verification for broader distribution.
3. Create an **OAuth client of type Web application** and register `http://localhost:9764/oauth/callback` (or your HTTPS dashboard domain plus `/oauth/callback`) as an **Authorized redirect URI**.
4. Under OAuth configuration in the Syncria web UI, enter the client ID, client secret, and dashboard URL (scheme + host + optional port, without trailing slash). Save, then click **Connect Google account**.
5. Repeat Connect Google account for each family member and create independent mappings.

Syncria requests the **full Google Drive scope**, which Google classifies as **restricted**. A publicly distributed OAuth app must meet Google's restricted-scope verification requirements. The owner of an installation must currently provide their own OAuth credentials; there is no embedded shared client secret.

## Sync behavior

- `root` represents the entire **My Drive file hierarchy**, not Google Shared Drives or items merely under “Shared with me”.
- New files on one side copy to the other. Identical files are tracked. **Different files at the same path on first sync cause a conflict error** and are left untouched.
- If a tracked file changes on both sides, the job pauses on that item and reports a conflict instead of discarding changes.
- Deletions propagate once a successful prior sync snapshot establishes the file. Locally removed remote files go to Google Drive trash; remotely deleted local files go under `<data>/recovery/<job id>/`.
- A missing or disconnected NAS path causes an error; it does not automatically trigger deletion, because that path is inaccessible.
- Remote Google Docs/Sheets/Slides are not ordinary binary files. This first build stops before syncing them to avoid corruption; exports and native-file round-tripping require separate handling. Note that the full-Drive option is not yet feature-complete for accounts containing such items.
- Some rename operations are currently represented as deletion+addition. In particular, moving a folder containing synced descendants needs additional safeguards before general-purpose release. Do not enable the alpha on a production Drive without a separate backup.

## Security and limitations

- State includes Google refresh and access tokens and the OAuth client secret. `state.json` is written with restrictive permissions on supported filesystems. Protect and back up its parent folder. On Windows, set filesystem ACLs appropriately.
- Google API requests are sequential; full tree rescans and MD5 hashing every local file are not optimized for very large Drives.
- Only one agent instance should use a data directory. Running two instances against the same state can corrupt its contents.
- The scheduler scans about every 30 seconds and makes a best effort to run jobs at their configured intervals. Runs are serialized per job.
- Support for shortcuts, Google-native documents, shared drives, Unicode normalization/case-folding differences, high-volume retries, checkpointed resumable uploads, and cross-device atomicity is planned, not currently implemented.

## Verify

```sh
go test ./...
go vet ./...
```

License: see LICENSE.