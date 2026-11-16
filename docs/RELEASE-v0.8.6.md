# Syncria v0.8.6 — Security maintenance release

This release preserves sync jobs, saved shares, mappings, administration, and existing configuration under `/data`.

- Rebuild all downloadable updater binaries using Go 1.27.4, incorporating the November 10, 2026 Go security patches.
- Update the final Docker runtime from end-of-life Alpine 3.21 to supported Alpine 3.24.4.
- Preserve existing multi-platform release assets (Linux, Windows, macOS) and SHA256SUMS verification.
- Retire the obsolete v0.8.5 automatic publication trigger so its pinned-version verifier does not fail or attempt to republish old artifacts.

Upgrade only after the v0.8.6 release contains all seven platform binaries and checksum file. Existing /data files are retained; upgrade is not a configuration reset.
