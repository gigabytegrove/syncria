# Syncria v0.5.0

This version delivers the cloud-to-NAS and standalone NAS-to-NAS synchronization features requested for Syncria.

- Independent Google Drive and Dropbox accounts, and multiple separately configured NAS devices.
- Cloud ↔ NAS relationships with one or multiple explicit destinations; cross-NAS propagation through the cloud is optional and off by default.
- Standalone NAS ↔ NAS relationships, with one-way or bidirectional synchronization.
- Exactly five NAS-to-NAS conflict-resolution choices: Newer file wins, Primary NAS wins, Secondary NAS wins, Keep both files, and Manual conflict resolution.
- Per-relationship file-deletion propagation, enabled by default, with an independently configurable 10-minute Undo window enabled by default.
- Deleted files are staged only at their NAS/cloud endpoints during the Undo window. Syncria keeps configuration and sync indexes, not user file payloads.
- Interval-based scheduling, plus remote SMB change notifications with periodic reconciliation.
- Saved account, NAS, relationship, and updater state is preserved on upgrade.

**Installation:** The primary update method remains Syncria's own **Updates → Check again → Install latest**, with binary checksum verification and in-app restart. GHCR images, Docker Compose and standalone release binaries remain alternative deployment methods. No additional updater container, Docker socket access, or mount permissions are required for direct SMB connections.

**Compatibility:** New NAS connections use native SMB2/3 access. Older mount-based relationships and their saved configuration remain supported under their existing mount setup; they are not silently converted or removed. Native direct NFS is not implemented in this release.

**Data safety:** A conflict or inaccessible endpoint stops the affected transfer rather than erasing an unverified file. Always maintain independent backups when first enabling synchronization; provider-specific authorization and live connectivity still depend on your own NAS/cloud configuration.
