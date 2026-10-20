# Syncria 1.0.0-alpha.1 — integrated platform testing checklist

**Release state: PRE-ALPHA CANDIDATE. Do not publish a GitHub release until the verified gates below pass.**

This candidate extends the existing v0.8.5 Go server, its standard dashboard, background scheduler, durable bbolt checkpoints and NAS bridge with a locally approved device-to-NAS **one-way relationship**. No separate synchronizer is introduced.

## Operator setup and alpha scope

1. **Use disposable folders first and preserve a separate backup**. Keep the existing v0.8.5 installation and `/data` snapshot intact. Do not upgrade production in place for this candidate.
2. Enroll a native desktop agent through **Devices** in Syncria. The agent's identity is durable and approval-gated.
3. On the agent, approve a source folder using `syncria-agent grants add` and identify its grant ID using `grants list`. Create its independent peer credential with `syncria-agent peer init` (secret displayed once). The local device owner starts `syncria-agent peer serve`, with a trusted TLS certificate and key for LAN use.
4. In **Devices → Sync device files to a NAS — alpha**, select the enrolled device, enter its peer HTTPS address, source grant ID, peer token, existing direct-SMB NAS, destination subfolder, and scan interval.
5. Syncria verifies the source inventory and destination NAS folder before creating the central job. The token is saved in a private owner-only file under `/data/agent-peer-keys/`; it does not appear in `state.json`, the list API or dashboard HTML.
6. The new sync starts **paused**. Inspect its destination and existing contents before enabling it from **My Syncs** or selecting **Sync now**. Central run history, progress, cancellation, per-job scheduling and bbolt checkpoints are reused.
7. A missing or inaccessible device returns an error, never an empty file list. Devices expose selected regular files only. A sync **never propagates deletions from the device**. Existing destination files with a conflicting first-run state stop for manual review. Independent edits to existing destination files stop the relationship rather than silently overwriting changes.
8. The other separate feature, **device-to-device peer receive**, remains a user-initiated workflow. Verified publication is Linux-only; Windows and macOS are currently supported as read-only approved sources, but are not yet supported as publication targets.

## Isolated Docker candidate deployment

The project now includes `compose.alpha.yaml` which **never references** the production image tag or existing `/data` volume. On a test Docker host:

```bash
git clone --branch feature/v2-release-foundation --single-branch https://github.com/gigabytegrove/syncria.git syncria-alpha
cd syncria-alpha
docker compose -f compose.alpha.yaml build --no-cache
docker compose -f compose.alpha.yaml up -d
docker compose -f compose.alpha.yaml ps
docker compose -f compose.alpha.yaml logs --tail=80
```

The alpha dashboard listens on port **19764** by default (change with `SYNCRIA_ALPHA_PORT`). Create a new administrator password; do **not** attach existing production data or NAS folders containing important files. The Dockerfile runs all repository package tests as part of the image build and will fail closed if compilation/tests fail. Cross-platform desktop agent binaries are separately available in the development candidate bundle. Do not enter peer secrets over an untrusted plaintext admin connection; use a TLS reverse proxy for remote browser administration.

## Safety boundaries

- Controller-managed agent jobs cannot browse unrestricted device folders, force-write to a device, or mutate grants. The source agent authorizes every new peer request and requires ongoing central heartbeat approval.
- Agent inventories are all-or-error, capped, validated against case collisions, traversal and suspicious entries. Content is streamed with a strong digest and length check. No payloads are intentionally stored in controller `/data`.
- NAS writes use the existing remote staging/commit path; source change or transport failure fails before a staged NAS object is committed. No new alpha source-to-NAS relationship can overlap an existing defined NAS source/destination relationship.
- **Unresolved:** existing SMB `Put` replaces a *previously synchronized, unmodified* destination file after creating a transient backup. Production-grade retained file version history, crash restoration of every replacement failure mode, and all platform-specific network/storage regression tests remain release blockers. Do not test replacement of valuable files.
- The source agent peer token is independent of enrollment; treat it as a sensitive credential. Network communication requires trusted TLS except explicit loopback development. There is no automatic remote agent installation or signed updater yet.

## Build and acceptance commands

After cloning PR #20 into a machine with Go **1.25 or newer**, run:

```bash
bash scripts/verify-platform-alpha.sh
```

This checks package tests with `-race`, vet and seven native desktop-agent cross-build targets without GitHub Actions. It does **not** replace real-device testing. Next perform: clean data directory installation; upgrade against a **copy** of a v0.8.5 data directory; actual SMB server and agent transfers on Windows/Linux/macOS sources; server restart mid-transfer; disconnection, revocation and symlink attacks; TLS certificate errors; UI browser smoke tests; rollback to v0.8.5.

## Gate status at development commit

- [x] Scoped agent files and peer streaming exist.
- [x] Central agent job adapter reuses existing bridge policy, scheduler and run-index checkpoints, with disabled deletion propagation.
- [x] Admin-only relationship API and Devices UI use existing sessions, one-way behavior and private peer credentials.
- [x] Isolated Go 1.23.2 race tests (nine targets) and vet passed; controller adapter exercised against HTTP fake endpoints.
- [ ] Entire Go 1.25 repository `go test -race ./...` passes.
- [ ] Real SMB integration, migration, rollback and browser smoke testing pass.
- [ ] Full Windows/macOS runtime verification and safe destination file publication.
- [ ] Formal signed platform binaries, release checksums, update compatibility, and standalone/Docker deployment verification.
- [ ] Production-grade version retention, conflict and recovery acceptance.

**Do not label this code a platform-wide approved alpha release or publish `v1.0.0-alpha.1` until the unchecked gates pass.**
