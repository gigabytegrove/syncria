# Syncria 1.0 alpha development — NAS version restore and retention

This pass modifies only approved **agent-to-NAS synchronization** recovery. Production remains v0.8.5.

## Administrator APIs

The existing administrator session authenticates all routes. Mutations also require same-origin POSTs, and operations share the existing synchronization mutex.

- GET /api/v1/agent-sync/versions?job_id=<job-id> — list archived versions without reading file bodies.
- POST /api/v1/agent-sync/versions/restore with JSON {"job_id":"...","version_id":"...","relative":"folder/file.txt"} — restore into a random, separate recovered filename. Live and historical copies are kept.
- POST /api/v1/agent-sync/versions/retention with {"job_id":"...","keep_per_file":3,"max_bytes":10737418240} — preview the exact proposed historical-file deletions and a digest.
- Repeat retention POST with "confirm":true,"expected_digest":"<preview.digest>" and unchanged job/limits. A mismatch rejects the operation.

Pruning never runs automatically. The last recovery version of each file cannot be deleted. A changed or incomplete inventory fails closed. Multi-file pruning is not transactional: failure stops the operation and reports the count already removed.

## Known limitations

- File bytes stream from NAS archive to NAS staging through bounded controller buffers, never to persistent /data.
- Restore verifies the source stream against a second read of NAS staging using SHA-256.
- Existing archived NAS versions lack capture-time SHA-256 manifests. Therefore this proves transfer integrity, not historical authenticity.
- SMB Rename does not guarantee no-replace on all appliances. Strong random restored names and collision checks mitigate but do not eliminate a malicious concurrent-writer race.
- Retention chronology uses archive directory timestamps, which may be ambiguous or modified outside Syncria; ambiguous snapshots stop pruning.
- Actual SMB appliance, Go 1.25 full-controller build, real OS agent, browser, rollback and migration tests remain required.

This is approved synchronization recovery, not a new remote file-management subsystem.
