# Syncria v0.8.6-alpha.5 — LAN HTTP Enrollment Support

**Alpha-only corrective release. Stable v0.8.5 remains unchanged.**

## What changed

**One-command enrollment no longer requires the Syncria web dashboard to run over HTTPS on your local network.**

You can open Syncria using an ordinary private LAN origin such as `http://192.168.0.19:9764`, `http://nas1.lan`, `http://syncria.local:9764` or another local/private address. Go to **My Computers → Add computer**, name the computer, and copy the provided Windows PowerShell, Windows CMD or Linux command. The command downloads the agent from **your Syncria server**, checks the release SHA256SUMS, configures it, and securely registers it automatically.

- Server-side validation allows HTTP **only for private LAN IPs, loopback/link-local hosts, local names, and custom DNS names resolving solely to local/private addresses**. HTTP public addresses are rejected; HTTPS remains supported for remote/public hosts.
- Windows and Linux bootstrap scripts use the current origin's HTTP or HTTPS scheme rather than requiring an HTTPS server.
- Agent LAN-interface discovery selects the right default hub port: **80 for HTTP** and **443 for HTTPS**, or the explicit URL port.
- On HTTP LAN installations, a clear **informational warning** explains that the first-time installer download and initial enrollment travel unencrypted over the LAN. The warning is not a blocker.
- Copying command blocks works in **HTTP browser contexts** using a selectable-text fallback, because the modern Clipboard API may require HTTPS.
- The rest of the process remains automatic: **short-lived single-use invitation, independently generated agent credentials, exact TLS certificate fingerprint verification, and encrypted agent-to-agent transfers**.

## Security and operating boundaries

Plain HTTP enrollment on a LAN is a convenience choice, not an encrypted or authenticated provisioning channel. The server-supplied `SHA256SUMS` provides integrity checking against the manifest received, but **cannot protect the installer from a malicious local-network intermediary who can change both the script and checksums**. Prefer HTTPS on untrusted/shared networks or when exposing Syncria beyond your LAN.

Existing direct agent connections and all computer-to-computer file transfers **continue to require encrypted certificate-pinned HTTPS**, even when the Syncria dashboard is accessed over HTTP.

**No automatic full-disk sharing:** the installer makes a dedicated `Syncria Shared` directory. The only files available to the agent are inside explicitly configured roots. Persistent transfer history, isolated worker processes, cancellation, retries and the My Computers dashboard are unchanged.

## Installation

Already on the Alpha channel: **App Updates → Check for updates → v0.8.6-alpha.5**. Back up Syncria's `/data` volume before updating, especially before moving between channels.

The release tag is `alpha-v0.8.6-alpha.5`, protecting Stable clients from experimental builds. The server's current container and mounts are retained by the in-app updater.

## Verification

Release-gated by application tests, race checks, vet, UI build, seven platform binaries and their SHA-256 manifests, and live HTTPS native agent-to-agent copy/move/no-overwrite tests. Added private-LAN HTTP / public-host rejection tests and browser fallback tests.

## Known limitations

Windows agent startup remains per-user (at sign-in), rather than a machine-wide service; Linux uses a user systemd service if available. LAN agent TCP 9765 must be reachable from the Syncria hub and other agents. Automatic WAN relaying, recursively transferring entire directories, resumable partial downloads and unattended native-agent updates are not yet supported.
