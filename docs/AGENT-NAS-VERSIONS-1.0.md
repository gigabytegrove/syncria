# Syncria 1.0 — NAS-side retained versions for agent-to-NAS jobs

**Development only: this change does not make 1.0.0 release-ready.**

## Integration

- Only `agent_bridge` jobs use retained replacement versions. Existing NAS-to-NAS and cloud-to-NAS relationships retain their present write behavior.
- The source is streamed to a temporary file **on the destination NAS** and verified before publication. On a replacement, the existing file is renamed on that same SMB share to `.syncria-versions/<job-id>/<random-version-id>/<relative-path>`, then the validated staging file is renamed to the original filename.
- If the final rename fails, Syncria attempts to restore the original file. A double failure leaves the original at its version path and produces an explicit error. No old file is silently deleted in the success path.
- The private archive tree is **excluded from ordinary SMB synchronization scans**, so saved versions cannot feed back as new user files.
- Administrators can enumerate retained version paths using `GET /api/v1/agent-sync/versions?job_id=<job-id>` with the existing signed browser session. The API exposes metadata only; no archive payloads or credentials.
- Versions are stored on the destination **NAS**, not under the Syncria controller's `/data`, and do not introduce an extra index or independent scheduler.

## Alpha safety and limitations

- Retained versions consume capacity on the destination NAS and are **not automatically pruned yet**. Quotas and retention policy enforcement remain required before release; monitor available space and use disposable test folders only.
- This first version of the API is **list-only**. Restoring an archived version to a fresh non-conflicting name through the authenticated dashboard remains to be implemented. Existing live destination files must never be silently overwritten.
- SMB servers differ in rename guarantees, durability and error codes. Full SMB integration tests on the user's actual NAS and power-loss recovery simulations remain release gates.
- Source deletion remains disabled for `agent_bridge`. Initial destination collisions and independent destination modifications continue to fail closed under the shared conflict engine.
- The complete Go 1.25 build, v0.8.5 migration regression, UI testing, cross-platform agent installation and rollback verification are not yet complete.

## Isolated unit evidence

`go test -race -count=3 -cover ./nasversion` and `go vet ./nasversion` pass in a Go 1.23.2 isolated module. Unit tests exercise initial creation, version retention, publication failures, rollback failures, cancellation, unsafe paths, malformed archives and bounded metadata listing. The full application and SMB adapter are not covered by this isolated result.
