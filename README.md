# Syncria

**Pronunciation:** Sink-Re-Ah. **Default local web UI:** `http://127.0.0.1:9764` (configurable with `-listen`). The repository and Go module are named `syncria`. Existing default data directories may retain the historical `gdsync` name to preserve configurations during upgrades. Restart the agent after upgrading; existing instances launched with an explicit `-listen` argument continue using that chosen port.

A lightweight, self-hosted, **experimental** cloud and filesystem synchronization agent written in Go. Single binary, embedded browser UI, no external database, and no JavaScript build tool. Suitable for Windows and Linux (including servers with mounted NAS shares).

> **Alpha warning:** This version has not been live-tested with Google Drive accounts. Make backups before enabling synchronization. Native Google Docs/Sheets/Slides, shortcuts, shared drives, files with duplicate names, and complicated rename/move scenarios are not fully supported. The agent **stops with a visible error** instead of silently processing unsupported Google-native formats. Do not use for your only copy of important data.

## Zero1Local companion — coming soon

Syncria is the planned **Google Drive companion system for [Zero1Local](https://github.com/gigabytegrove/zero1local)**. Future Zero1Local versions will remove their native Google Drive OAuth and synchronization functionality rather than duplicating Syncria's engine.

- Syncria remains a fully standalone application, supporting its existing native Linux/Windows deployment and standalone Docker operation.
- It is **optional**: Zero1Local does not need Syncria for ordinary NAS, local backup, shares or administration, and Syncria does not require Zero1Local.
- A future **optional connector** may display Syncria health and job status in Zero1Local. The current Syncria browser-session endpoints are not a stable connector API; there is **no working Zero1Local integration yet**.
- Syncria can run on a separate host and use only explicitly permitted NAS shares. This is particularly important for low-resource Zero1Local-ix2 devices, which do not have Docker support.
- The companion must not require a specific container management platform, a Docker socket, a secondary updater container, or privileged NAS admin credentials. Syncria retains its in-app update mechanism.

This is a product roadmap, **not** a declaration that Google Drive sync or Zero1Local integration is production-qualified. Track the integration in [Syncria issue #1](https://github.com/gigabytegrove/syncria/issues/1) and see [Zero1Local's companion specification](https://github.com/gigabytegrove/zero1local/blob/main/docs/SYNCRIA.md).

## v0.3.0 dashboard replacement

The original single-page dashboard markup has been deleted. The application now uses a separate `ui.go` template with a responsive sidebar, dedicated Overview, Sync relationships, Storage, Google accounts, Updates and Settings screens, connection-status cards, a modal Drive/local folder navigator, an isolated sync setup wizard, and network-share administration. The existing Go sync engine, saved accounts, credentials, and mappings remain compatible; this update is a **UI replacement**, not a configuration reset.

The UI template includes regression tests in `ui_test.go`. Release builds must pass compilation and tests before binaries are published. The default Docker deployment remains unprivileged; enabling NFS/SMB mounting requires the documented opt-in capabilities. Only published release binaries are offered by the embedded updater.

## In-app updates — no host updater service

Syncria updates **from inside the running application**, including Docker deployments.
Open **Updates → Check again → Install latest**. The authenticated Go application
checks published GitHub Releases, downloads the binary for its own OS/CPU, verifies
it against the release `SHA256SUMS`, stages it in persistent `/data/runtime`,
and activates it using the **existing launcher process in the same container**.
Existing accounts, mappings, job history, storage mounts and runtime backups are
preserved. **Rollback** also runs inside the application. No host systemd unit,
sidecar, Docker socket, Docker management integration, or auxiliary update
container is required.

A **source commit is not an installable release**. The updater deliberately does
not execute arbitrary source code or build compilers inside the running production
application. It will offer a new version only after a trusted maintainer publishes
that version's platform binaries and checksum manifest as a GitHub Release.

Release creation is a **one-time maintainer/build action**, not an installed
updater or scheduled host service. When explicitly publishing a version, a
maintainer may run `scripts/publish-release.sh vMAJOR.MINOR.PATCH` from a
trusted checkout with their own GitHub publishing credentials and local Docker
toolchain. This builds and publishes the release **once**; it installs **nothing
persistent** on the runtime host. Set `SYNCRIA_PUBLISH_GHCR=1` only when a
multiarchitecture container image is also needed. Runtime installations
do not need the maintainer's write credentials.

The earlier optional `syncria-publisher.service`, timer, and
`/etc/syncria-publisher.env` design was removed. On hosts where that optional
legacy service was installed, stop and disable its timer and service, remove
only their dedicated unit files and credentials, and reload systemd. This does
not touch a running Syncria Docker container or its volumes.

**Release requirements:** for in-app installation, the GitHub Release must
contain `syncria-<os>-<arch>` (plus `.exe` on Windows) and `SHA256SUMS`. A
container image is not required for ordinary app updates; rebuilding a container
is still appropriate when changing the base OS or launcher. The updater checks
that synchronization is idle before installing.

## Managed NFS and SMB connections (opt-in)

From the dashboard, use **Storage connections → Connect network share**, choose SMB or NFS, enter the server and share/export path, and connect. Connected shares are stored under `/storage/syncria-shares/<connection ID>` and become available through the normal local destination browser. Syncria stores SMB credentials in a 0600 credentials file in its persistent data volume. Connection definitions are preserved across restarts, and the application attempts to remount them at startup. Disconnecting is blocked while a sync mapping uses that location.

**Security:** Linux mounting requires elevated kernel privileges. The default container deliberately does not have these permissions. For trusted, isolated installations, enable mounting explicitly with `SYNCRIA_ENABLE_MOUNTS=1` plus the container's `SYS_ADMIN` capability and the host's appropriate mount security policy. Container runtimes with restrictive AppArmor/seccomp profiles may still reject the operation. Granting such permissions significantly expands container access and should not be done on a shared or untrusted Docker host. Alternatively, mount shares on the host and expose their parent directory under `/storage` without privilege elevation.

For native Linux agent deployments, install `nfs-utils` or `nfs-common` and `cifs-utils` (depending on distribution), and run the agent under a controlled service with explicit rights to mount network filesystems. Do not expose the web interface directly to the public internet; use TLS and an access-controlled reverse proxy.

## Version v0.4.1 — setup workflow, transfer visibility, updater progress

The **New sync** setup flow is four distinct stages: **1. Source** (Google Drive
or a local/NAS folder), **2. Destination** (an accessible local/NAS folder),
**3. Schedule**, and **4. Review**. Relationships still start **paused** so
the operator explicitly chooses when a new job begins.

Schedules include 5, 10, 15 and 30 minutes, hourly, every six hours, daily,
and **on filesystem changes** on Linux. On-change scheduling requires
`inotifywait`, included in the Docker image starting with v0.4.1. It triggers
on local path events but **does not guarantee detection of changes written by
other hosts to NFS/SMB shares or Google Drive**. Therefore on-change jobs also
run a five-minute reconciliation scan. No new service or updater container is
needed.

Each relationship displays its live current stage, processed entries, files
uploaded/downloaded, and transferred bytes. A determinate entry-processing
bar is displayed when the total is known; otherwise the bar is indeterminate,
rather than fabricating a percentage. Live progress is transient; persisted
per-job totals remain after restarts.

The built-in **Software updates** screen now starts a background operation
inside the Syncria process and polls an authenticated endpoint for stages:
check, download (real received bytes and known Content-Length), SHA-256
verification, installation, restart and successful version confirmation.
Progress is durably recorded under the existing `/data/runtime` directory
and reconciled by Syncria itself on boot. Failures are displayed in the
interface. The UI never automatically navigates away or forces a reload.
The update requires a **published v0.4.1 GitHub release**; a source commit
alone cannot appear in the in-app updater.

**Storage permissions:** existing local disks and host-mounted NAS folders
exposed under `/storage` work without privileged Docker permissions.
Creating new Linux SMB/NFS mounts **inside** an isolated container requires
explicit mount capabilities and is hidden when managed mounting is disabled;
the app does not request a Docker socket. NAS connections not yet exposed
under `/storage` cannot be configured as regular filesystem sync targets
without a mount or a future native network-storage adapter.

**Provider status:** Google Drive and local or mounted filesystem folders are
implemented sources. Dropbox is identified as unsupported and cannot yet be
selected. The UI does not pretend a provider can be synchronized before its
authorization and sync engine exist.

## Version v0.4.0 — verified setup and transfer statistics

The administrator navigation is rendered **only after authentication**. The running
version is presented with one `v` prefix. The Sync relationship workflow is now a
three-stage wizard: choose Google Drive or a storage folder as the source; choose
a destination and frequency; verify both sides and review the mapping. Pressing
Enter on an incomplete form cannot create a relationship. The backend rejects
missing, inaccessible, overlapping, and unmounted managed-storage folders.

**Source providers currently implemented:** Google Drive; filesystem storage
(including host-mounted disks, SMB and NFS shares available within the configured
storage root). Both modes synchronize in both directions. Google Drive mappings
already stored in the configuration remain compatible and do not need migration.
OneDrive, Dropbox, and S3 are **not** implemented in v0.4.0; the provider selector
does not offer nonfunctional options.

Filesystem-to-filesystem mode protects both sides: it never overwrites divergent
files on the first run, blocks simultaneous edits and pauses for manual review on
detected deletions. New sync relationships are always created disabled.

Network connections have an explicit three-stage wizard with discovery,
TCP reachability verification and a final mount check. The account connection
wizard explains and hands off to Google OAuth; existing accounts can verify
their Drive permission. Mounting inside Docker still requires an opt-in and
appropriate kernel permissions. An existing host-managed mount is sufficient.

The Overview and Activity & stats pages expose persisted per-job and aggregate
run counts, successful/failed runs, transferred file counts, payload upload and
download bytes, latest run duration and average payload throughput. Bytes are
measured for successful full-file transfers and **are not** a per-packet or
network-interface bandwidth meter. Historical transfers before this update
cannot be backfilled and start at zero. Statistics persist in `state.json`.

**Upgrade:** back up `state.json` and the data directory. Use the built-in
updater once the versioned v0.4.0 binaries and checksums are published, or
publish from your trusted machine with the existing zero-cost publisher. No
container management platform integration, Docker socket or updater sidecar
is needed.

## Interactive sync location selection

The web dashboard now includes authenticated folder browsers. Choose a connected Google account, click **Browse Drive folders** to select the whole **My Drive** root or navigate its folders, then click **Browse storage folders** to navigate an existing local destination. Each mapping remains paused until explicitly enabled. Folder lists use the linked Google account and the agent's filesystem; no third-party browser service is involved.

**Docker storage boundary:** The storage browser starts at `/storage` and can navigate child folders. Managed NFS/SMB connections are now configurable under **Storage connections**. Network mount management is **disabled by default** because it requires Linux kernel mount privileges and relevant network filesystem support. Enable it only on a trusted single-purpose instance; Docker users must explicitly grant mount permissions and set `SYNCRIA_ENABLE_MOUNTS=1`. Without these permissions, the app reports a mount error rather than silently saving an unusable connection. If a new NAS share is mounted under an exposed host root using the correct bind propagation, its directories become selectable in Syncria. The browser refuses paths outside the configured storage root and excludes symlink escapes.

**Distribution:** The GitHub source has changed, but previously published GHCR image tags are not automatically rebuilt or published. Deploying a new web UI requires a new image build and publisher-authorized push. Do not assume `latest` contains these features until the digest changes.

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

The configured Docker image is `ghcr.io/gigabytegrove/syncria:latest` (linux/amd64 and linux/arm64). GitHub source commits do not automatically rebuild or publish this image. Use `docker compose pull` only after confirming that an image was actually published; ordinary application upgrades use the Updates screen and do not require an image pull.

Create a Compose deployment using the repository's `compose.yaml` with the following environment values:

- `SYNCRIA_BIND_IP`: host-side listener IP (default `127.0.0.1`).
- `SYNCRIA_DATA_DIR`: persistent state, OAuth tokens, recovery data, updater runtime (default `./data`).
- `SYNCRIA_STORAGE_ROOT`: one **existing** host directory exposing locally mounted disks/NAS devices, visible inside Syncria as `/storage` (default `/mnt`).

```sh
docker compose pull
docker compose up -d
```

The runtime deployment does not need a `build:` block for normal application updates. All accounts and relationships are created and changed from the Syncria dashboard. For example, the host path `/mnt/nas1/photos` appears at `/storage/nas1/photos` inside the container.

**Docker isolation limitation:** the app cannot access arbitrary host paths outside the configured storage root. Existing host-mounted SMB/NFS shares are selectable within that root. Syncria also offers opt-in managed mounts on Linux, requiring explicit OS mount capabilities; the default container deliberately lacks those privileges. Never grant Docker socket access for updating Syncria.

The image includes the runtime launcher and uses `/data/runtime/current` for built-in app updates, so future app updates do not require a new container image. Security updates to the base image or launcher still require an image pull/recreate.

## Built-in updater

From the **Software updates** panel, sign in as the administrator and choose **Check for updates**, **Install latest**, or **Rollback**. Releases are fetched directly from GitHub Releases, **not a container registry**. Each release must include the correct OS/CPU binary named `syncria-<os>-<arch>` (with `.exe` for Windows) and a `SHA256SUMS` asset. Downloads are checksum-validated before activation. The previous executable is kept for rollback. State, account credentials, and mappings stay in the configured data directory.

In Docker, the launcher monitors the child agent and activates staged binaries from persistent `/data/runtime/current`. Future container restarts continue using that version, so **rebuilding the Docker image is not necessary for normal app releases**. Rebuilding is still recommended when the container base OS, launcher, or bundled CA certificates change. Linux/macOS standalone installations can restart into the staged binary without modifying their original executable. Windows installs can check for new releases, but automated installation and restart are intentionally disabled in this alpha.

**Update prerequisites:** Publish a GitHub Release containing the platform binaries and `SHA256SUMS`. The updater checks `gigabytegrove/syncria` directly. GitHub Actions are not used; `scripts/publish-release.sh` is the standalone build and publisher. If GitHub still considers the repository private, provide a read-only repository token using the `SYNCRIA_GITHUB_TOKEN` environment variable; do not put this token in a public repository or shared Compose file. Only releases from a repository you trust should be installed. SHA-256 checksums confirm transfer integrity, but do not provide publisher authenticity independent of the release itself. Updater actions are blocked while the agent is actively synchronizing files.

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