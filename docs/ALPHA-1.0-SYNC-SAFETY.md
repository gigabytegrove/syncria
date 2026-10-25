# Syncria 1.0 alpha — non-destructive device synchronization safety

This change stays inside the approved centralized file synchronization scope. It adds no remote file manager, unrestricted endpoint administration, or independent sync engine.

## First-run device-to-NAS conflicts

- An existing NAS file with the same relative path as an approved device file is a **conflict** even if size and modification timestamp match. Metadata is not proof of byte equality.
- Device-to-NAS alpha will not silently overwrite or silently checkpoint an unverified initial collision. Move or rename the NAS file after review, or use an empty NAS destination folder.
- Keep the real `agent_bridge` provider in shared policy execution. NAS-only provider identity must not mask device policy.
- An independently modified NAS file is a conflict; changes made to the device do not authorize overwriting that destination.
- A missing, inaccessible, revoked, or incomplete device inventory must not be interpreted as proof of deletion. Source deletions do not propagate to NAS for alpha.
- Existing Google Drive, Dropbox and NAS-only policy defaults remain unchanged.

## Inventory protocol safety

- Successful inventory requires a **non-null** `entries` JSON array. A true empty array is valid; absent or null fields, truncated JSON, unknown envelope fields, and trailing data fail closed.
- Enforce the entire inventory-response byte limit before JSON decoding.
- Validate locally issued folder grant IDs before network calls, and validate/URL-encode the selected file's relative path. Legal characters such as `?` and `#` must stay filename characters.

## Verification and release gates

Run `go test -race -count=3 ./agentpeer ./agentcontrol ./agentclient ./agentpolicy ./agentfiles ./agenttransfer ./agentpublish ./cmd/syncria-agent` and `go vet` for those packages. Run the controller bridge-conflict tests with the entire Go 1.25 repository test suite.

The isolated Go 1.23 agent-package suite is not a full controller build. The actual SMB endpoint, real-device installations, browser UI, a copy of the production data migration, and rollback still require verification. Do not promote this draft as a platform-wide alpha release or overwrite a working 0.8.5 deployment without those gates.
