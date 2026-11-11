# Syncria 1.0 — Agent security foundation

The package `internal/agentsecurity` provides a reviewed starting boundary for agent enrollment and root-scoped path validation. It is not yet wired to network APIs and is **not production enrollment**.

## Design invariants
- Pairing requires a server-generated random one-time code and a separate administrator approval.
- Codes expire; enrollment credentials are unique to each device, cannot be retrieved from the in-memory registry, and can be revoked.
- Do not enable agent filesystem APIs until authentication, persisted credential storage, an authenticated authorization check for every operation, and handle-relative traversal protections are implemented.
- `AuthorizedPath` is a preflight validation helper, **not a race-safe file-open primitive**. A concurrent symlink replacement can defeat a check-then-open API. Use secure handle-relative opening and platform-native safeguards for actual reads/writes.
- The in-memory registry is deliberately not connected to production jobs: credentials and approvals would be lost across restarts. Implement encrypted or permission-constrained durable storage, recovery, and audit before enabling the feature.
- Never return administrator approval credentials to an unauthenticated device. Rate limit pairing attempts at the eventual network boundary.
- The first release must retain existing cloud/NAS job semantics and startup migrations.

## Current status
- Cryptographic one-time pairing generation: implemented as library code.
- Separate approval, single redemption, credential authentication and revocation: implemented in-memory.
- Root-path preflight checks and symlink rejection: implemented as library code.
- Protocol endpoints, persistence, root grants, race-safe filesystem handles, heartbeat, scheduler integration, user interface and update mechanism: **not yet implemented**.

Tests: `go test ./internal/agentsecurity` and `go test ./...` (run from the repository root with Go 1.25+ and downloaded modules).
