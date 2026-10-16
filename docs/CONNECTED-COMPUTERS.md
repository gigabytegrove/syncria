# Syncria connected computers — direct file transfers

This feature is part of the opt-in **v0.8.6-alpha.3** preview ([release notes](RELEASE-v0.8.6-alpha.3.md)) and under ongoing review in [PR #19](https://github.com/gigabytegrove/syncria/pull/19). **Production Stable remains v0.8.5.** The cross-platform Syncria executable also runs as a native computer agent.

## What the first implementation does

- Enroll Windows and Linux computers in **My Computers**, using a per-agent pairing key plus a pinned SHA-256 TLS certificate fingerprint.
- Browse explicitly shared folders on each computer from the existing Syncria web dashboard.
- Copy and move **individual files** directly between two agent computers without SMB, NFS, Google Drive, Dropbox, or temporary storage on the Syncria hub.
- Show each copy/move operation and its result. A move verifies the received bytes before asking the source agent to remove the original.
- Never overwrite an existing destination file. A copy that lands successfully remains on the destination if final source deletion fails.

Current limitations: **whole-directory copying, automatic agent-to-agent sync relationships, shared agent folder mounts, resumable interrupted downloads, WAN relay and unattended agent upgrades are not included.** Real byte progress and durable per-file transfer history are included in alpha.2. Transfers use machine-to-machine HTTPS; both agents must be able to connect to one another and the Syncria hub must connect to both.

## Agent setup

First, download a Syncria executable from a release containing this feature (not v0.8.5). The native agent is the **same executable** as the server. It does not require SMB sharing, a Windows file-sharing account, a database, or another Docker container.

Windows (PowerShell, in the directory containing the executable):

```powershell
.\syncria-windows-amd64.exe -agent -listen 0.0.0.0:9765 -agent-data "C:\SyncriaAgent" -agent-name "Desktop" -agent-root "Documents=C:\Users\Public\Documents"
```

Linux:

```bash
./syncria-linux-amd64 -agent -listen 0.0.0.0:9765 -agent-data "$HOME/.config/syncria-agent" -agent-name "Laptop" -agent-root "Documents=$HOME/Documents"
```

To share multiple folders, repeat `-agent-root Name=/absolute/path`, using an absolute path and unique label for each root.

The first start generates an independent pairing key, a self-signed TLS certificate and its SHA-256 fingerprint. **Protect the agent identity directory as you would a password.** Do not expose the agent port to the public Internet, and restrict the TCP 9765 firewall rule to your trusted LAN or VPN. For all normal use, run the agent as an unprivileged user that has access only to the intended folders.

In the main Syncria dashboard, open **My Computers → + Add computer**. Follow the three-step onboarding wizard and select Windows or Linux. Enter a friendly name, HTTPS address (for example `https://redacted.invalid:9765`), the pairing key and certificate fingerprint printed by the agent. Syncria tests the connection and rejects invalid keys or fingerprints.

To transfer files, open the separate **Computer Transfers** workspace. Select a source computer and shared folder on the left, choose up to 100 individual files, then select a destination computer and folder on the right. **Copy file** or **Move file** creates independent queued jobs per file. Monitor actual bytes, cancel work, filter history, or retry failures without overwriting files.

## Security and behavior

- Direct transfers use HTTPS 1.2+ with exact certificate fingerprint verification on **both** Syncria→agent and destination agent→source agent connections.
- Each agent only exposes folders explicitly supplied at launch. Paths cannot contain traversal segments or symbolic links; special files are excluded from browsing.
- The hub requests a 30-minute file-scoped transfer ticket from the source agent. The receiving computer uses that ticket to stream the selected file from the source.
- The receiver stages data in a hidden temporary file in the destination directory, checks byte count and SHA-256 streamed by the source, then publishes the destination using no-overwrite hard linking. Successful transfers remove their temporary staging files.
- A **Move** checks the original source's size, modification time and SHA-256 before deleting the source. If the original changed or deletion fails, the destination copy remains and the transfer reports the failure.
- The agent HTTPS address is an origin only: no URL path, query, userinfo or plaintext HTTP allowed. Redirects and environment proxies are disabled for agent links.
- The hub stores agent connection credentials with Syncria's existing `state.json` configuration, which should remain protected by filesystem permissions and regular backups.
- Transfer records persist at `/data/runtime/computer-transfers.json` (mode 0600). An interrupted job is marked as such after a restart; it is not silently replayed. Verify both computers before explicitly retrying.
- Each active transfer uses a **separate OS worker process** spawned by the Syncria hub. Three workers can run concurrently; remaining jobs wait. Worker credentials travel through stdin and are never written to the transfer journal.
- Canceling a transfer kills the worker; on Linux the worker also dies if its parent hub exits unexpectedly. The receiving agent cleans up unfinished staging where possible. A completely verified destination copy may remain after interruption.

## Deployment notes

The central Syncria server still runs as a single Docker container and only needs its existing port. Native agents must be installed **on the computers whose files they access**, because running them inside the hub container would see only the container's filesystem, not each desktop's disks.

If using Docker bridge networking, be sure the container can reach each computer's HTTPS agent port. The receiving computer must also reach the source agent directly. Docker volume mounts and SMB/NFS access are not used for these peer transfers.

Existing cloud accounts, NAS shares, sync relationships, session handling, index data and in-app updater are unchanged. There is no automatic migration or reconfiguration of existing jobs.
