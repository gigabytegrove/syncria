# Syncria 1.0.0 — implementation contract and release gates

Status: **approved scope; not implemented or released**. Base audited: `main` at `f4d694d8214cd203a7881ba52563223adbd2132d` (published v0.8.5).

## Invariants

1. Syncria remains a self-hosted **synchronization engine**, not a remote desktop, generic file manager, backup replacement, or unrestricted device administration suite.
2. Preserve existing Google Drive, Dropbox, NAS/cloud, NAS-to-NAS, locally mounted storage, saved relationships, credentials, updater, UI behavior and bbolt index. Existing /data configuration and indexed checkpoints must migrate without loss.
3. Agents are **opt-in first-class endpoints** using existing job and relationship semantics, not a second independent scheduler/index/reconciliation engine.
4. File payloads must not be persisted on the controller. Use bounded streaming, authenticated endpoint connections, atomic destination commits and destination-owned temporary storage.
5. Treat unreachable, permission-denied, incomplete or invalid source scans as **unknown**, never as proof of deletion. Explicitly authorize roots and fail closed on access violations.
6. Preserve selected conflict policies, deletion protections and recovery behavior. Never silently change default policies as part of a migration. Synchronization is not a backup.
7. Avoid a mandatory external SaaS account, cloud relay, database server, Docker socket, sidecar updater or GitHub Actions for normal running operation. LAN-only installation must remain functional.

## Release delivery map

| ID | Scope | Engineering contract | Required proof |
|---|---|---|---|
| A01 | Distributed agents | Windows service, Linux service, supported macOS daemon, dedicated endpoint IDs; explicit approval and revocable credential; bounded heartbeat | enrollment, restart persistence, revocation, expired-credential and offline tests |
| A02 | Device-to-storage synchronization | Agent endpoint adapter implementing scan/open/put/mkdir against approved roots; reuse central policies, job history and index | agent-to-NAS, NAS-to-agent, agent-to-agent e2e tests on real files |
| A03 | Selective folder sync | normalized allow-list roots and include/exclude selectors; reject escapes and overlap; preview | unit, race, symlink, case-sensitive/insensitive path and migration tests |
| A04 | File versioning/recovery | destination-supported version storage with retention and quota; atomic restore; independent backup warning | restore original bytes; crash/retention/permission/full-disk tests |
| A05 | Conflict management | preserve both sides on ambiguous edits; per-relationship explicit options; durable pending decisions | simultaneous writes, clock skew, duplicates and conflict restore |
| A06 | Automatic transfer resumption | durable offset, source revision/size/hash binding, verified chunk boundaries, safe revalidation | terminate during multi-GB file; resume exact bytes; changed-source rejection |
| A07 | Centralized monitoring | accurate heartbeat with last-seen, reachability, disk info, queue, errors, compatibility | offline/online transitions, pagination, stale data, permission tests |
| A08 | Agent self-updates | signed manifest, verified artifacts, compatibility range, rollout, rollback; do not rely on mutable unsigned binaries | tamper rejection, rollback, interrupted update, Windows/Linux/macOS smoke |
| A09 | Optional Android companion | native Kotlin, Android Storage Access Framework, background constraints and approved media folders; web UI stays primary | permission revoke, Wi-Fi/power constraints, process death, device tests |
| A10 | Visual topology | real endpoint/job graph; statuses sourced from server; no synthetic edges or online claims | multi-endpoint graph and accessibility tests |
| A11 | File permissions and audit | resource ACLs, deny-by-default device root access, log enrollment, data operations and admin actions | cross-device isolation, unauthorized requests, replay and immutable audit tests |
| A12 | Bandwidth/schedules | per-job rates, concurrency, sync windows, pause/resume and timezone/DST semantics | measured limits, time window, restart persistence and job fairness |

## Architecture implementation order

**Foundation:** define controller/agent wire contracts, capability negotiation and transport authentication. Keep controller authentication/session routes separate from scoped agent credentials. Use device-specific keys and approval codes with TTL, one-time redemption, rotation and revocation. Do not permit directory browsing until a root is granted. Never accept arbitrary absolute paths from the controller; send root identifiers plus canonical relative paths only.

**Endpoint adapter:** evolve the current `bridgeEndpoint` abstraction to support negotiated endpoint capabilities while preserving cloud/NAS adapters. Do not change existing saved relationship JSON in place without a reversible migration. Keep bbolt as the single checkpoint index; make new device relationships reference the same index and history model.

**Transfer protocol:** requests identify job, endpoint ID, approved root, relative path, immutable source revision and bounded byte range. Use per-transfer unique IDs and idempotency keys. Check revision and path authorization for **every** scan/open/write operation; handle mid-transfer revoke safely. Store partial files on destination, in the same filesystem as final paths; fsync and atomically promote only after final digest verification. Resume offsets are trusted only when the source revision and the stored chunk digests still match.

**Versions and conflicts:** maintain durable metadata for pre-replacement versions with quotas; never call an overwritten file recoverable unless its bytes have actually been saved. A missing or stale scan cannot initiate delete propagation. Surface explicit decisions for simultaneous changes and any provider without safe rollback capability.

**Operations and UX:** existing web UI remains primary; add Devices, Transfers, Topology, Versions, Audit and Policies with real addressable routes and guarded API handlers. Never render inactive controls for unfinished capabilities. Maintain accessible responsive layouts and readable nontechnical terminology.

**Updates:** existing standalone controller updater remains intact. Agent update manifests need release signing, architecture/platform checks, staged install, health confirmation and rollback; no unattended reboot.

**Android:** deliver after desktop agent/central APIs stabilize, as an optional separate native app. Adhere to OS-managed scheduling; do not guarantee immediate background sync.

## Test and verification matrix

- **Unit:** protocol parsing, ACL, relative-path canonicalization, link escape, enrollment TTL/replay, revoked credentials, checkpoint validation, conflict decisions, version retention, DST schedules and quotas.
- **Integration:** baseline Google Drive/Dropbox/SMB jobs remain identical; new agent adapter works with the same job/index; zero migration data loss with a copied v0.8.5 data directory; interrupted transfer resumed or safely restarted with verified final checksum.
- **Adversarial:** unauthorized root, symlink swap, malformed chunks, duplicate messages, stale or modified source, network partition, out-of-space, unexpected reboot, revoked device during stream, UI CSRF and token exposure in logs.
- **Performance:** record scan duration, stable-content skip rate, SMB listings, memory/CPU, network traffic and transfer rates against v0.8.5 using representative NAS trees. Do not publish unmeasured performance claims.
- **Cross-platform:** Windows amd64/arm64, Linux amd64/arm64/arm, macOS amd64/arm64 controller parity where existing release platforms support it; agent service operation per target; native Android tests separately.
- **Release assets:** reproducible source commit, versioned binaries, checksum manifest, signed agent-update manifest, image digest when published, migration guide, rollback guide, release notes and a verified version response after upgrade.

## Blocking release checklist

Every checkbox below is a **release blocker** unless explicitly marked optional. No box is checked simply because code was committed.

- [ ] A01–A08, A10–A12 implemented, integrated, and verified (Android A09 may ship later only if clearly labeled optional/future in release notes).
- [ ] Existing controller and cloud/NAS tests pass without regressions.
- [ ] Agent enrollment, approval, scoped access, expiry, revocation, secrets at rest and audit traces tested.
- [ ] Large interrupted transfers, mismatched revisions, recovery and atomic replacement tested.
- [ ] Conflict and deletion tests prove no unintended loss.
- [ ] v0.8.5 configuration, index, passwords, relationships and updates migrate on copied production-like data.
- [ ] Windows, Linux and macOS agent packages install, run, upgrade and uninstall cleanly on supported OSes.
- [ ] UI pages, APIs, accessibility and authorization tested in a real browser.
- [ ] Documented threat model, vulnerability review, resource budgets, backups and operational recovery drills complete.
- [ ] Built binaries and images verified against source revision, checksums and signed update metadata.
- [ ] Release candidate tested with reproducible evidence and no unresolved data-loss/security blockers.
- [ ] Only then set `VERSION` to `v1.0.0`, tag and publish a production release.

## Baseline code pointers

- `bridge_engine.go`, `bridge_policy.go`: current streaming endpoint and conflict semantics.
- `storage_sync.go`: local storage checks and atomic file staging.
- `file_index.go`, `file_index_startup.go`: existing persistent bbolt checkpoints.
- `bridge_recovery.go`, `bridge_recovery_routes.go`: existing recovery semantics.
- `scheduled_fast_check.go`, `sync_run_history.go`: existing scheduling and history.
- `main.go`, `ui.go`: existing management APIs, session boundaries and UI.
- `updater.go`, `update_progress.go`: existing controller update path.
- `docs/PRODUCT-DIRECTION.md`: already locked synchronization-first direction.

## Current reality

The published production release remains **v0.8.5**. This document defines exactly what must be built and demonstrated for the **v1.0.0** milestone. It does not constitute feature implementation, runtime verification, or release approval.
