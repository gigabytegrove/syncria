# Syncria connected computers — direct file transfers

This feature is part of the opt-in **v0.8.6-beta.1 Beta** ([release notes](RELEASE-v0.8.6-beta.1.md)) and under ongoing review in [PR #19](https://github.com/gigabytegrove/syncria/pull/19). **Production Stable remains v0.8.5.** The cross-platform Syncria executable also runs as a native computer agent.

## What the first implementation does

- Enroll Windows and Linux computers in **My Computers**, using a per-agent pairing key plus a pinned SHA-256 TLS certificate fingerprint.
- Browse explicitly shared folders on each computer from the existing Syncria web dashboard.
- Copy and move **individual files** directly between two agent computers without SMB, NFS, Google Drive, Dropbox, or temporary storage on the Syncria hub.
- Show each copy/move operation and its result. A move verifies the received bytes before asking the source agent to remove the original.
- Never overwrite an existing destination file. A copy that lands successfully remains on the destination if final source deletion fails.

Current limitations: **whole-directory copying, automatic agent-to-agent sync relationships, shared agent folder mounts, resumable interrupted downloads, WAN relay and unattended agent upgrades are not included.** Real byte progress and durable per-file transfer history are included in Beta. Transfers use machine-to-machine HTTPS; both agents must be able to connect to one another and the Syncria hub must connect to both.

## Agent setup — one command

On your Syncria server, open **My Computers → Add computer** and enter a computer name. Syncria generates a **single-use, 20-minute invitation** and three setup commands:

- **Windows PowerShell:** copy the PowerShell command.
- **Windows CMD:** copy the CMD command (it launches the PowerShell installer).
- **Linux terminal:** copy the Linux command.

Run that command on the computer being added. It detects the CPU architecture, downloads the platform-specific executable and checksum manifest **from your own Syncria server**, verifies the download, creates an initially empty `Syncria Shared` folder, installs the agent under the user's account, and configures startup where supported. The native agent automatically discovers local IP addresses and generates a pairing key and TLS certificate. Syncria validates those credentials and the HTTPS certificate fingerprint before adding the computer.

**Both HTTP on a trusted private LAN and HTTPS are supported for the Syncria web dashboard and initial enrollment.** For example, `http://192.168.0.19:9764` works without setting up SSL. Public hosts must use HTTPS. If using HTTP, the installer and initial pairing travel unencrypted on your LAN. The checksum downloaded over that same HTTP connection **does not authenticate the binary against an active network attacker**; use HTTPS on untrusted networks. All actual Syncria-to-agent and agent-to-agent file transfers remain certificate-pinned HTTPS regardless of the dashboard scheme.

The agent listens on TCP 9765. Ensure your Syncria server can reach that port and, for direct transfers, that the destination computer can reach the source. Restrict any firewall rule to your trusted LAN/VPN. Agent credentials are stored in its private identity directory and should be protected. Windows agent startup is per-user at sign-in; Linux uses a systemd user service if available, with a foreground-independent fallback if not.

### Advanced: manual agent setup

For troubleshooting only, the same published native executable supports manual mode with `-agent`, `-agent-data`, `-agent-name`, `-listen` and one or more `-agent-root Name=/absolute/path` options. For normal use, prefer the one-command setup so you do not need to enter IP addresses or paste keys and certificate fingerprints manually.

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
