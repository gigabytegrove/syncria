# Syncria 1.0 – release-critical work and pre-local verification

This change stays strictly within the approved file-synchronization product: device sync policies, retained verified file versions, and authenticated agent updates. **It does not make Syncria 1.0 or its alpha 90% verified.**

## Integrated into centrally scheduled agent-to-NAS jobs

- Optional `agent_sync_policy` remains absent from existing v0.8.5 and existing development jobs. Unspecified policy preserves current behavior.
- New agent-to-NAS jobs accept `policy` with `kib_per_second` and `windows` (IANA timezone, allowed weekdays, start/end minute of local day).
- The existing central scheduler/job runner remains the only execution engine. The job checks the configured time window before opening a device peer or NAS, and its source reader rechecks during transfer to stop after a window closes.
- Only device-source traffic on the selected job is rate-limited. Other cloud/NAS jobs are unchanged. A transfer canceled by its schedule fails before completing its NAS staging commit.
- Limit is per file stream, not a global bandwidth allocator. Per-site fairness and contention benchmarks remain to be completed.

Example `policy` in the authenticated agent job-create JSON request:

```json
{
  "kib_per_second": 2048,
  "windows": [
    {"timezone":"America/New_York","days":[1,2,3,4,5],"start_minute":540,"end_minute":1020}
  ]
}
```

The Windows in the example permit Monday–Friday 09:00–17:00 local time. Overnight windows and daylight-saving transitions are covered by the isolated policy tests. The dashboard form does not expose these advanced fields yet.

## Approved library features implemented but NOT wired into production operations

- `agentversion` provides private immutable saved versions, persistent manifest, SHA-256 before recovery, bounded quota and per-file retention. Recovery returns a verified stream; it **never overwrites a live file**. The caller still needs a destination-supported publish/restore workflow before this counts as installed versioning in the actual sync engine.
- `agentupdate` verifies Ed25519-signed, time-limited release manifests, platform/architecture match, artifact length and SHA-256. It stages only verified files and refuses replacing existing staged artifacts. This **is not yet an unattended auto-updater**: trusted release signing, OS-specific service installation, atomic executable swap, health probes and rollback remain open.
- These libraries use the standard Go library, have no SaaS dependency, do not change existing storage or credentials, and have no autonomous background services.

## Tests performed locally for this change

Run from the development module with available Go 1.23.2:

```
go test -race -count=3 ./agentcontrol ./agentclient ./agentfiles ./agentpolicy ./agenttransfer ./agentpeer ./agentpublish ./cmd/syncria-agent ./agentversion ./agentupdate ./syncpolicy
go vet ./agentcontrol ./agentclient ./agentfiles ./agentpolicy ./agenttransfer ./agentpeer ./agentpublish ./cmd/syncria-agent ./agentversion ./agentupdate ./syncpolicy
```

**What these do not prove:** a full Syncria v1 Go 1.25 application build, real SMB/cloud compatibility, browser UI tests, performance, native agent service installation or rollback. Do not treat isolated-package passes as an approved platform alpha.

## Gates remaining before requesting one coordinated user test pass

1. Build and test the **complete checked-out GitHub branch** with its pinned Go 1.25 dependencies, without replacing production 0.8.5.
2. Complete actual version-retention integration around destination replacements and explicit restore semantics for each supported storage provider.
3. Complete signed agent update application, services/installers and rollback for supported Windows/Linux/macOS targets.
4. Complete requested read/write endpoint parity, safe conflicts, offline recovery, legitimate topology display and UI policy controls.
5. Exercise staging/restart/retry on real SMB and agent devices and certify safe v0.8.5 migration against copied production data.
6. Verify published checksums, browser workflows, security threat model and disaster rollback in an isolated Docker/VM environment.

Do not move PR #20 to main, change VERSION or publish a formal 1.0 release until release gates pass. The optional Android companion is not required for initial desktop alpha.
