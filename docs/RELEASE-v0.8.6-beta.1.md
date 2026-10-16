# Syncria v0.8.6-beta.1 — Connected Computers and Direct Transfers

**First Beta preview of Syncria's connected-computer features.** This promotes the verified `v0.8.6-alpha.6` functionality without merging experimental code into Stable. Stable remains `v0.8.5`.

## Included

### My Computers: one-command agent installation
- A dedicated, organized computer dashboard with connection health, permitted shared folders, and guided Add Computer.
- **Windows PowerShell, Windows CMD and Linux** setup commands provided by Syncria's own server. No manually downloading platform binaries or copying pairing keys, IP addresses or certificate fingerprints.
- Installer chooses the target CPU, retrieves the exact matching release executable and SHA256SUMS **from the running Syncria hub**, verifies checksum, installs under the local user's profile, creates an initially empty `Syncria Shared` folder and configures startup where supported.
- Short-lived, single-use 20-minute invitation; native agent generates its independent TLS identity and credentials; hub connects back to a validated private/LAN address and verifies the exact TLS pin before saving the device.
- **LAN HTTP or HTTPS for enrollment:** `http://redacted.invalid:9764` or your trusted local hostname works without configuring a reverse proxy or TLS certificates. HTTP enrollments are **unencrypted during initial provisioning** and therefore for trusted private networks only. Public endpoints must use HTTPS.
- Agent-to-agent and Syncria-to-agent file traffic **always** requires pinned HTTPS, even if the hub's dashboard is on HTTP.

### Computer Transfers: separate transfer process and workspace
- Dedicated navigation and dual-pane source/destination explorers with searchable folders and batches of up to 100 selected regular files.
- Each transfer runs in a **separate native OS worker process** (up to three concurrent), without routing file contents through the hub or SMB/NFS.
- Per-file queues, actual received-byte progress, cancel, explicit retry, status filters, and persistent journal/history across hub restarts.
- No-overwrite destination commits and SHA-256-verified copies. A **Move** deletes the source only after the destination copy is verified, and verifies the source has not changed.
- Interrupted transfers are flagged for human review, never automatically replayed and risk overwriting or deleting a potentially completed copy.

### Existing Syncria features retained
Google Drive and Dropbox accounts, NAS connections, bidirectional and unidirectional sync, indexing, recovery policies, authentication, built-in versioned updater, and dashboard workflows remain available.

## Install or upgrade

**Already on alpha.6:** Open **App Updates → Update channel → Beta → Save channel**, check for updates and install `v0.8.6-beta.1`. Beta outranks alpha in the version comparator; no command-line redeployment or Docker replacement is required.

**Running Stable v0.8.5:** Legacy Stable-only installations cannot discover Beta prerelease tags. Back up Syncria's persistent `/data` volume, then use **App Updates → Advanced options → Manual installation** with the platform executable and `SHA256SUMS` from this Beta release. After restart, select the **Beta** update channel.

**Already on Beta:** Syncria discovers the highest published Beta (or subsequent Stable) release while leaving Alpha-only builds out of the Beta channel.

The release tag is **`beta-v0.8.6-beta.1`**, which deliberately begins with `beta-` to prevent legacy v0.8.5 tag scanning from offering this prerelease as Stable.

## Verification

This Beta is gated on Go unit and integration tests, race checks, `go vet`, local Tailwind UI compilation, seven native binaries and their SHA256SUMS, and an end-to-end test involving **two running TLS agents and a separately started worker process** covering copy, verified move and refusal to overwrite a destination. Beta-specific tests confirm:
- `beta-v0.8.6-beta.1` is newer than `alpha-v0.8.6-alpha.6`.
- **Beta** and **Alpha** channels can discover the Beta while **Stable** continues to select v0.8.5.
- The restart verifier compares correctly normalized versions.
- The Beta server delivers matching Beta agent binaries for one-command installation.

## Known limitations

This Beta handles individual files and batches, not recursive directory transfer, resumable partial downloads, WAN relays or unattended native-agent upgrades. Windows startup is per-user at sign-in, **not a machine-wide Windows service**. Linux uses a systemd user service where available. Firewall/routing must allow agent TCP 9765 between the hub and peers.

The checksum manifest and installer downloaded over LAN HTTP are both subject to local network tampering, so prefer HTTPS on untrusted networks. Beta is a testing channel, not the Stable production release.
