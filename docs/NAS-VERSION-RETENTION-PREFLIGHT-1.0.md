# Syncria 1.0 — NAS retention safety preflight

Scope: destination-owned historical versions for approved agent-to-NAS synchronization. Existing cloud/NAS policies are unchanged.

Retention previously verified and then deleted historical versions one at a time. If a later candidate was corrupted, the earlier archives could already have been deleted. ApplyRetention now performs a read-only full preflight of every candidate **before any removal**, checking path, size, timestamp, capture-proof identity and SHA-256 where available. It then checks each candidate again immediately before removing it.

A missing archived payload or proof that cannot be read is now an incomplete inventory error, never a successful empty version. This deliberately fails closed and may require operator inspection after an interrupted NAS operation.

Limitations: Remote SMB cleanup is not multi-file transactional. An outage or third-party NAS modification after preflight can still cause partial deletion; the API reports the completed count. SHA-256 is a corruption check, not a signature. Legacy archives lack capture-time digests. Deletion remains administrator-reviewed and explicitly confirmed. Do not prune important NAS data until real SMB checks, full Go 1.25 application build and upgrade tests pass.

Isolated compatible Go 1.23.2 NAS-package harness passed 38 named tests, ten race repetitions, go vet and 76.0% statement coverage. Six Linux/Windows/macOS architecture test binaries cross-compiled; these were not runtime tested on those platforms. Exact repository-wide controller tests remain pending.
