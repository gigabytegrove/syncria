# Syncria v0.7.2 — Actionable sync details and efficient metadata reconciliation

## Changes
- **Unchanged files are silent.** The engine scans directory/Drive metadata, but does not open the contents of an unchanged, previously tracked file or display its path in previews or live transfer activity.
- **Skipped files and replacements are visible, by filename and reason.** Every result includes the planned action and reason; the preview shows only action items, including skipped/blocked/review entries, rather than using its 500-row budget on thousands of unchanged files.
- **Preview action filters:** all actionable, skipped (with reasons), replacements/overwrites (with reasons), new copies, review/blocked, deletes.
- **Correct progress wording:** scanning displays counts and endpoint names, not every unchanged file path.
- **First-run timestamps:** files with equal sizes and modification times within two seconds are treated as equivalent metadata matches, without opening full file contents. Conflicting sizes still require review.
- **Transient NAS-bound transfer retry:** up to three attempts with backoff if a NAS-bound transfer fails with a transient deadline/network error; existing NAS destination files are preserved until a staged transfer verifies successfully. Blind cloud-upload retries remain disabled to prevent duplicates.
- **Live Activity:** shows up to 2,000 relevant per-file results with reasons; preview can show up to 10,000 actionable results, with a clear truncation warning and full counts.

## Important
Syncria still inventories metadata on both endpoints during a reconciliation; discovering remote changes without any metadata requests would require a reliable change feed or filesystem notifications for every configured NAS. The existing v0.7.1 cancellation controls are retained.

A context deadline during a **genuinely necessary** transfer can still occur because of network outages, remote service timeouts, or SMB behavior. This release reduces unnecessary transfer attempts and retries safe transient NAS-bound errors, but does not yet provide byte-offset continuation for interrupted 30 GB uploads or a guarantee that network I/O can never fail.

No user sync configuration or stored checkpoints are reset by this release.

## Metadata exclusions and baseline clarity
- `Thumbs.db`, `ehthumbs.db`, and `ehthumbs_vista.db` are ignored case-insensitively at every folder depth across supported connectors. Files already present are not removed or interpreted as deleted when an old checkpoint contains them.
- The live UI indicates when a sync is establishing its initial baseline so longer first-time inventories are expected.
- **Index limitation:** v0.7.2 retains the existing `PairStates` checkpoint metadata in `config.json`; it does **not** introduce a new SQLite or key-value index file or incremental Google Drive change-feed tracking. A dedicated persistent index and fast delta scanning are being developed separately because they require safe migration and remote deletion handling.
