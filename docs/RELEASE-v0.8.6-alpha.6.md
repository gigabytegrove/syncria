# Syncria v0.8.6-alpha.6 — LAN HTTP Enrollment (Complete UI and Help Correction)

**Alpha preview. Stable remains v0.8.5.** This supersedes the alpha.5 LAN HTTP implementation by also ensuring the My Computers wizard and agent setup documentation consistently explain the supported connection modes.

## Fix

- No more blocking **"One-command enrollment requires Syncria to be opened using trusted HTTPS"** message when the user accesses Syncria via a private LAN `http://` address.
- The Windows PowerShell, Windows CMD and Linux installer commands are generated from the currently opened HTTP **or** HTTPS origin, and run against that Syncria server.
- The server validates that a plain HTTP enrollment URL belongs to a **private LAN IP or local hostname**, allowing private split-horizon DNS names where they resolve exclusively to private addresses. Public HTTP hub addresses are rejected; normal public HTTPS remains available.
- Agent enrollment honors HTTP as its hub transport, including correct default TCP port 80. Agents continue to establish and verify their own certificate-pinned **HTTPS** listener for all remote browsing and direct file transfers.
- **Copy command** works from regular HTTP browser pages using a legacy selection-copy fallback, even when modern `navigator.clipboard` is blocked in an insecure context.
- My Computers wizard wording and the Connected Computers setup guide now describe LAN HTTP support plainly and avoid conflicting HTTPS-only instructions.

## No new setup required

Go to **My Computers → Add Computer**, name the computer and copy the correct one-liner. It downloads a verified agent from your Syncria server, sets up a dedicated `Syncria Shared` directory, registers an appropriate per-user startup mechanism and securely pairs the machine using a one-time 20-minute invitation. IP addresses, agent pairing keys and certificate fingerprints are negotiated automatically.

## Network security tradeoff

HTTP on a trusted private LAN is supported for convenience. **The initial script download, manifest and enrollment credentials are transmitted without encryption and can be intercepted or modified by an attacker on that LAN.** Matching a SHA256SUMS file downloaded over the same insecure channel does not protect against a malicious intermediary. Prefer HTTPS on shared, guest, corporate, remote or untrusted networks.

**Direct computer-to-computer transfers are unaffected**: they still use pinned HTTPS, no-overwrite destination commits and SHA-256 verification. Existing NAS/Google Drive sync, update channels, transfer history and separate isolated transfer workers remain intact.

## Verification

Go tests, `-race`, vet, compiled web UI, seven cross-platform native executables and SHA256SUMS; real paired HTTPS agent copy/move and no-overwrite end-to-end test. New regressions assert real HTTP LAN bootstrap script generation, private/custom LAN hostname acceptance, public HTTP rejection, and unblocked HTTP browser onboarding.

## Update

On the **Alpha** channel, use **App Updates → Check for updates → Install v0.8.6-alpha.6**. The existing Syncria Docker container and persistent data stay in place. Back up `/data` before an experimental update.

Release tag: `alpha-v0.8.6-alpha.6`. Stable `v0.8.5` remains unchanged.
