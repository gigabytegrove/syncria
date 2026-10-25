# Syncria v0.8.6-alpha.4 — Self-Hosted One-Command Computer Agents

**Experimental Alpha. Production Stable remains v0.8.5.** Existing cloud, NAS, indexed synchronization, verified update channels, and the separate computer-transfer manager are retained.

## Automatic agent installation and pairing

**My Computers → Add computer** is now a one-command onboarding flow:

1. Enter a friendly computer name.
2. Syncria issues a private, random 256-bit, **single-use invitation** valid for 20 minutes. It displays copyable **Windows PowerShell**, **Windows CMD**, and **Linux terminal** command blocks.
3. On the computer being added, copy and execute the matching command. No separate manual downloads, address entry, pairing-key copying, or TLS fingerprint entry.
4. The command retrieves the native platform executable and `SHA256SUMS` **from the running Syncria server itself**, not from GitHub or a third-party endpoint on the target machine.
5. Syncria's server obtains its own matching release binary once from the official published GitHub release, verifies the binary's SHA-256, and caches the checksum-verified binary under the persistent `/data/runtime/agent-downloads` directory for subsequent installations. The client installer independently verifies its downloaded binary against the same manifest before activation.
6. Installer detects CPU architecture and installs a regular user-mode agent under the user's profile/home. Creates an **empty** `Syncria Shared` folder; it does **not** expose the whole disk, Documents directory, or arbitrary folders.
7. The agent creates its own random bearer pairing key and persistent TLS identity, detects candidate local interface addresses and announces them to the Syncria hub. The hub requires a valid invite, connects back to the proposed private/LAN address, verifies the exact TLS certificate pin and secret, and stores the verified configuration automatically.
8. The pending invitation file is protected inside the agent's identity directory and deleted after verified pairing. No enrollment secrets are included in permanent startup commands.
9. The UI polls the invitation and completes the wizard automatically when the securely verified computer comes online.

### Agent startup

- **Windows:** installer starts the native agent immediately and creates a per-user Windows Startup shortcut to start at sign-in. It attempts an inbound **local-subnet-only** Windows Firewall rule; making that rule may require an elevated PowerShell window. This is a per-user startup agent, **not a Windows service that runs when signed out**.
- **Linux:** installer registers and starts a `systemd --user` service when available, configured to restart on failure; otherwise it starts the agent using `nohup` and tells the user that startup persistence was unavailable. User services may require lingering or a signed-in session to run without an interactive login.
- Both listen on TCP **9765** for agent-to-agent HTTPS. Your Syncria hub must be reachable from the computer over trusted HTTPS. NAT, firewalls, routing and custom TLS certificates can affect automatic discovery and must be configured correctly.

### Explicit security boundaries

- Creating invitations and viewing their status requires Syncria admin session authorization.
- Bootstrap and asset endpoints require an unexpired invitation. Replayed, expired, or already-consumed credentials are rejected.
- The installer only accepts trusted HTTPS with normal certificate verification. It doesn't bypass TLS validation.
- Only known platform filenames are served; arbitrary paths and script destinations are rejected. Files are never shared until the verified connection is established.
- Pairing addresses are limited to actual private/loopback/link-local IPs presented by the agent, and must respond through a verified pinned TLS client; a public arbitrary URL cannot be enrolled.
- These are executable administrative installation commands. Run only commands generated directly in your own Syncria instance.

### Keeping existing functionality

The separate **Computer Transfers** workspace remains intact, including direct peer-to-peer transfer workers, batch operations, measured progress, persisted history, retries, cancellation and no-overwrite SHA256-verified moves.

### Limitations

Agent installation is a user-mode agent, not an always-on machine-level Windows service. Running in the background while signed out requires a service-aware Windows implementation. The first time Syncria downloads the release binary into its cache it needs outgoing access to GitHub, even though the newly enrolled target computer does not.

Recursive whole-folder transfers, resume within a partially copied file, WAN relay and unattended native-agent updates remain future work.

## Upgrade

With Alpha selected in **App Updates**, install **v0.8.6-alpha.4** using the regular verified updater. Existing paired agents remain as-is; one-command installation applies to newly added computers. The server's HTTPS origin must be accessible and its certificate trusted by the client platform.

The GitHub prerelease tag is `alpha-v0.8.6-alpha.4`. Stable users must manually opt into Alpha using their existing verified manual-install route and back up `/data` first.
