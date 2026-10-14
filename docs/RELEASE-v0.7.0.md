# Syncria v0.7.0 — Timestamp-based sync restoration

Syncria now checks file modification timestamps and file sizes when first matching files already present on Google Drive and a NAS. It no longer reads the entire NAS file just to determine whether an existing file needs to be copied. For example, a 30 GB video with matching modification times and sizes can be marked unchanged from metadata alone.

- Newer source or destination timestamps are considered under the relationship's configured direction and conflict policy.
- One-way cloud → NAS never sends a NAS file back to the cloud.
- Matching timestamps and sizes are treated as unchanged; differing sizes with matching timestamps require review.
- Missing modification times are not treated as proof that files match. Explicit one-way cloud relationships still obey the source-wins policy.
- The read-only preview uses the same first-run timestamp checks.

This version changes file comparison policy, not cloud or NAS data locations. As with any synchronization engine, retain an independent backup of important files before enabling destructive sync/delete operations. Timestamp-only comparisons cannot identify changed contents whose timestamps were preserved.

The release pipeline requires Go tests, race tests, vet, cross-platform builds and checksummed binaries to pass before the in-app updater release is published.
