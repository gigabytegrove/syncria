# Syncria 1.0.0 development — approved read-only peer transport

**Development only. No production 1.0.0 release yet.** This is the first functional agent-to-agent file-data path. It does not yet create centrally managed synchronization jobs, write to a NAS/cloud destination, propagate deletions, or offer unattended bidirectional synchronization. Existing v0.8.5 cloud/NAS relationships are unchanged.

## Supported operations

- Device owner approves folder roots and selections with `syncria-agent grants add` (unchanged local-only policy).
- Peer service exposes authenticated, **read-only** inventory and file streams for allowed paths, with SHA-256 ETags and HTTP range support.
- Receiver independently stages authenticated bytes in the existing `agenttransfer` store, committing each SHA-256-verified chunk and verifying the entire object before declaring completion. Interrupted transfers resume from the **committed offset** using `Range` and `If-Match`; changed files invalidate the previous revision.
- No HTTP write/create/delete endpoint exists. No remote caller can create or expand a local folder grant.
- Peer token is separate from controller enrollment. The operator provisions it locally, and the server does not reveal the token after creation.
- The peer handler **fails closed** unless the enrolled controller has acknowledged a heartbeat within the preceding 90 seconds. A controller revocation terminates service at the next scheduled heartbeat. Do not disable the authorization gate.
- Non-loopback listeners require a TLS certificate/private key. HTTPS is always required for remote peer clients. Loopback HTTP is supported for explicit local development only.
- Device filesystem paths are not stored in or exposed to the controller; inventories return grant-relative paths and metadata only.

## Native agent use

Run on the **source device** after enrollment (example, Linux):

```sh
syncria-agent grants add -name Pictures -root /srv/pictures -select albums -read-only
syncria-agent peer init
syncria-agent peer serve -listen 0.0.0.0:9774 -tls-cert /path/to/certificate.pem -tls-key /path/to/private-key.pem
```

The initial token prints once. Keep it out of shell history, logs and process arguments. Use a certificate trusted by the receiving device; TLS verification is never disabled. Normal deployment should use a supervisor so the peer listener and heartbeat loop restart together.

On a **receiver** with its own enrolled agent identity:

```sh
syncria-agent peer inventory -peer https://source.example:9774 -grant <source-grant-id>
syncria-agent peer stage -peer https://source.example:9774 -grant <source-grant-id> -relative albums/example.jpg -source-id <source-device-id>
```

The commands prompt for the peer token via standard input. The receiver records validated bytes under its private `incoming/` directory and prints the staged transfer ID. **No destination file is overwritten.** Promotion requires a future grant-scoped destination adapter and the centralized job conflict/deletion checks.

## Security and operational constraints

- Enforce the device owner's local grant and include/exclude policy on **every** requested source file.
- On Linux, file reads use per-component `openat` with `O_NOFOLLOW`; neither a symlink nor path traversal can redirect a file read. On other Go 1.24+ platforms, `os.OpenRoot` constrains access to the root. Cross-platform filesystem-race and installer validation are still release blockers.
- Inventory errors invalidate the entire result; a partial scan is **never** evidence of deletion. The current inventory does not authorize deletion propagation.
- Peer tokens grant read access to all folders locally approved on that agent. Treat them as sensitive and rotate by stopping the peer service, replacing the local token, and updating receiving peers. More granular per-controller/per-grant credentials are required before GA.
- Current metadata inventory scans and per-transfer SHA-256 checks may read large files twice. Optimize and benchmark NAS-scale data sets before claiming high-performance readiness.
- Do not run the peer service directly on an untrusted public interface. Limit listener access by firewall or private networking, use valid TLS, and keep the controller healthy.

## Verification commands

```sh
go test -race -count=3 ./agentcontrol ./agentclient ./agentpolicy ./agenttransfer ./agentfiles ./agentpeer ./cmd/syncria-agent
go vet ./agentcontrol ./agentclient ./agentpolicy ./agenttransfer ./agentfiles ./agentpeer ./cmd/syncria-agent
```

This test matrix exercises successful downloads and interrupted resume, invalid tokens, missing grants, rejected paths, changes between HEAD and GET, empty/unauthorized data, source folder revocation, controller authorization failure, canceled requests, Unicode filenames, and private peer token storage. The repository's complete `go test ./...` and installed-OS smoke tests are still required; isolated package tests do not substitute for full release evidence.
