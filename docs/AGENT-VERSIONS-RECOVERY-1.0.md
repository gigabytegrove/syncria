# Syncria 1.0 — local approved-folder version recovery

**Development only, not a formal alpha or complete automatic version-retention system.** This work connects the existing endpoint-owned `agentversion` archive to the locally authorized file-access and atomic destination-publication modules. It adds no server-side file cache, controller filesystem write API or new synchronization scheduler.

## Supported, tested now

- An **enrolled** native desktop agent can locally save an immutable version of a file inside an explicitly approved folder (`versions save`). The archive is private to the endpoint, protected by a byte quota, persisted across restarts and SHA-256 verified.
- `versions list` enumerates stored history by job identifier and optional relative pathname.
- `versions restore` reads a saved version and places the verified bytes inside a locally approved **writable** folder. Its destination is a separate sibling filename such as `photos/birthday (restored-4f019a8f7c32).jpg`, not the original file. **No current content is replaced or removed.** A second attempt for the same recovered filename is rejected.
- Existing locally approved include/exclude selections, writable-grant checks and grants revoked during transfer are enforced. On Linux, no-replace publishing is implemented via opened directory handles and atomic links, never path-based `os.Rename` overwrite.
- Invalid snapshots (source changes detected after capture) have their recorded version removed; archive deletion commits the manifest *before* unlinking the blob.

## Local usage

The agent must already be enrolled through the existing Syncria Devices UI. Set the grant locally, then:

```sh
# Create an approved local grant on the device, or select its existing ID.
syncria-agent grants list
# Save one historical file version (existing 32-character lowercase hex sync job ID).
syncria-agent versions save -job <job-id> -grant <approved-folder-id> -relative photos/photo.jpg
# List saved versions and recover one to a new filename.
syncria-agent versions list -job <job-id>
syncria-agent versions restore -version <version-id> -grant <approved-writable-folder-id>
```

Default version storage is in the enrolled device's private configuration directory. `-archive` and `-quota-bytes` can select a different private archive and a bounded size. Never put it on an untrusted shared folder or inside a synchronized destination. These commands run **on the endpoint only**; they do not provide a remote file browser, controller-triggered restore or independent backup service.

## Remaining release blockers

- Automatic version capture at NAS/cloud/device replacement points is **not** integrated. The current CLI captures explicit owner-selected versions; it does not promise historical recoverability for every sync operation.
- Atomic publication to a locally writable folder currently uses the supported **Linux** backend. Windows/macOS must pass equivalent native no-replace and symlink-race tests before enabling local restore there. Cross-compiling an executable is not runtime proof.
- Existing SMB/NAS snapshot/restore and conflict resolution still need real endpoint-level acceptance, storage quota enforcement under load, retention policy UI, and recovery drills.
- Full Go 1.25 repository build, Docker install/upgrade, browser testing, v0.8.5 migration and platform integration remain unverified in this isolated Go 1.23.2 environment.

Do not replace a production deployment or publish 1.0 stable based on these package tests alone.
