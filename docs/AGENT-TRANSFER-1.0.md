# Syncria 1.0.0: locally approved folders and durable transfer staging

These packages are **implemented and covered by automated tests**, but they are not yet attached to remote agent RPC, NAS/cloud adapters or the synchronization planner. They are deliberately isolated to avoid premature remote file access or data loss.

## `agentpolicy` — locally approved sync folders

- Only the device owner using the local agent CLI can add/remove grants. The central controller cannot silently grant itself access to new paths.
- Grant IDs are generated from 128 bits of randomness. Every grant has an absolute, existing, canonical directory, a label, a `read_only` flag (true by default in the CLI), optional selected relative subfolders and explicit exclusion prefixes.
- Exclusions take precedence; all relative path components are validated and traversal sequences rejected. Broad or overlapping root grants are rejected. Grant files use owner-only permissions and atomic replacement; an invalid existing state prevents startup.
- `Allows(grantID, relative, write)` is a **policy check only**. It is not a race-safe file opener. The controller does not receive absolute paths. Future adapters must enforce the active grant at each operation using secure handle-relative opening and must not trust metadata from scans as evidence of safe writes.

Examples (`-state` defaults beside the local identity file):

```sh
syncria-agent grants add -name "Family Photos" -root /home/me/Pictures -select Family -exclude Family/Private
syncria-agent grants add -name "Documents" -root /home/me/Documents -read-only=false
syncria-agent grants list
syncria-agent grants remove -id <id-shown-by-list>
```

## `agenttransfer` — safe local staging and resume

- `Begin` authenticates **nothing by itself**; callers must enforce identity, job authorization and folder scope *before* calling it. The store only accepts validated source revision, relative destination path and a full-file SHA-256.
- Each transfer reserves a 128-bit random session ID and an allocation under a dedicated owner-private staging root. Default documented quota is 32 GiB; callers can provide a smaller quota.
- `Append` accepts a bounded payload (maximum 4 MiB) at an exact offset and validates a SHA-256 chunk digest. Repeated identical complete chunks are idempotent; skipped offsets, overlaps or conflicting replays fail.
- Payload bytes are flushed before advancing the atomic manifest. On restart, any uncommitted trailing data are discarded, and the previous durable offset is resumed. A failed checksum cannot advance the checkpoint.
- `Complete` verifies the entire payload with SHA-256 before marking the session complete. `OpenCompleted` independently verifies the digest again before returning a read-only handle.
- Pending sessions can be expired/aborted and capacity reclaimed. Completed sessions are never automatically deleted. These are staging files on **the endpoint** and are not uploaded to controller `/data`.

## Release blockers and limitations

1. Agent policy and transfer packages have no network-facing API; they do not yet synchronize a computer or NAS.
2. Root access needs platform-specific, race-safe **handle-relative** filesystem operations, and the authorization service must bind every transfer to a live approved job, device and root ID.
3. Full-file checking currently requires one additional read before committing. Immutable source revisions and upload limits will be enforced by the future endpoint adapter.
4. The `Store` protects concurrent goroutines within one process. Multiple independent processes opening the same state directory are not supported; the production service must use one owner or a cross-process lock.
5. Completed staging files need a two-phase promotion operation backed by destination protection/versioning before file replacement is allowed.
6. Full Syncria repository Go 1.25 build, NAS/cloud regression, Windows/macOS runtime smoke tests, service installers and UI integration must still pass before release.

Validation: `go test -race -count=2 ./agentpolicy ./agenttransfer ./cmd/syncria-agent` and `go vet ./agentpolicy ./agenttransfer ./cmd/syncria-agent`.
