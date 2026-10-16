# Syncria v0.8.6-alpha.2 — Connected Computers & Dedicated Transfers

**Alpha testing release only.** Stable v0.8.5 remains unchanged. This release builds on alpha.1 without removing existing cloud, NAS, authentication, sync, or updater functionality.

## Real improvements

### My Computers — dedicated device management
- Dedicated device inventory instead of a giant mixed connection/transfer page.
- Live online/offline status, device endpoint, visible shared folders, summary totals, quick transfer shortcut, status recheck, and safe disconnect.
- Three-step onboarding: choose Windows/Linux and name the device; start the native agent; securely pair with address, key, and SHA-256 TLS certificate fingerprint. Includes platform command and copy action.
- Pairing verifies the real reachable agent before saving credentials. No unsecured HTTP or unverified TLS.

### Computer Transfers — a separate full workspace
- Dedicated navigation item and route, not a section embedded under My Computers.
- Two-sided remote file explorers with source and destination computer and shared-root selection, parent navigation, filename filter, selected-file counts and byte totals.
- Queue up to **100 individual files** from one source folder per copy or move batch. Each file gets its own independent transfer record.
- File operations stream **directly from one native agent to another** via mutually pinned HTTPS; no SMB, NFS, temporary hub file storage or cloud relay.
- Up to **3 independent isolated native transfer-worker OS processes** run concurrently. Additional accepted transfers wait in a queue. Worker receives short-lived connection credentials via private stdin, not command-line arguments or durable job files.
- Receiver reports actual transferred bytes; worker records preparation, copy, destination verification and source removal separately.
- A persistent transfer journal in Syncria's existing data volume records pending/active/completed/failed/canceled jobs and actual byte progress. Finished transfer records survive restarts. Jobs interrupted by a hub restart become **Interrupted**, requiring manual inspection and an explicit retry; no blind replay of potentially completed file operations.
- Individual cancel, safe retry of unsuccessful jobs, clear-history-record, queue/active/finished summary counters, and queue filters.
- Active workers stop with the hub's cancellation context. On Linux, a kernel parent-death signal also kills orphaned workers.
- Disconnecting a computer with active jobs is blocked rather than silently orphaning the transfer.

### File integrity and move protection
- No overwriting any existing destination path.
- Source issues a file-scoped short-lived transfer ticket.
- Destination verifies exact byte count and SHA-256 before publishing the staged file.
- A move deletes the source only after a complete verified destination commit, and independently checks that the source still matches.
- Cancellation or an incomplete move may leave a verified destination copy; Syncria does not delete or overwrite this copy automatically. Check both machines before retrying.
- Folder-level recursive transfer, resumable interrupted downloads and automatic agent-to-agent synchronization **remain unsupported**; this alpha handles selected regular files only.

## Installation and update channels

This alpha's GitHub tag is `alpha-v0.8.6-alpha.2`, while the running version is `v0.8.6-alpha.2`. This tag prefix prevents older Stable-only clients from accidentally offering alpha builds.

**Already running alpha.1:** in **App Updates**, select **Alpha**, save, then check for updates and install **v0.8.6-alpha.2**. Syncria verifies SHA256SUMS and activates the new executable inside the existing Docker container.

**Still running Stable v0.8.5:** use **App Updates → Advanced options → Manual installation** with the platform executable and SHA256SUMS from this prerelease; then select Alpha in the updated UI. Back up the entire persistent `/data` volume before crossing channels.

Your Docker image's `:latest` label may remain unchanged because the existing supervised launcher uses `/data/runtime/current` after an in-app update.

## Automated checks

Go tests, race tests, static vet, UI build, cross-platform compilation and SHA-256 verification are release gates. A native two-agent system test runs standalone worker subprocesses to verify a real copy, an authenticated and verified move, rejection of an existing destination and source preservation on failure.

## Troubleshooting

If a computer shows Offline, check that its agent is running, the hub can reach its HTTPS port and the paired TLS fingerprint and key match. The receiving computer must also be able to reach the sending computer directly.

Use **Computer Transfers → Queue and history** for concrete per-file failures; a move failure does not mean both copies were removed. Never remove an interrupted transfer's source or destination until inspecting both systems.
