# Syncria Indexed Sync Engine — v0.8.0 development

## First-run index
On first start, Syncria creates `syncria-index.db` in its configured persistent data directory (`/data` in the standard container). It imports prior per-file checkpoints from `state.json` without reading file contents, retains a private `state.json.pre-index-v0.8.0` backup, and moves high-cardinality pair records out of the normal settings JSON. The embedded bbolt database is transactional, pure Go, and restricted to owner permissions.

A first-time sync inventories the chosen cloud/NAS locations and saves metadata for each reconciled file. Metadata includes paths, cloud identifiers, each side's size and modification time, and previously synchronized fingerprints when available. It does not hash file contents merely to compare changed/unchanged status.

## Later runs
The engine loads the DB index at start and compares provider metadata (file sizes/modification timestamps) against committed records, instead of reading old file contents. Unchanged files are silently ignored. The existing planning UI reports only skips, replacements, deletes and conflicts with reasons.

A successful file transfer commits its checkpoint before moving to the next transfer; initial unchanged metadata matches are committed in batches. If index writing fails, the sync stops rather than claiming work completed.

## Google Drive incremental inventory
- At initial inventory, obtain a Drive Changes API cursor **before** listing files to avoid missing concurrent updates, then persist the tree snapshot and cursor together in the local database.
- On normal five-minute checks, ask Google Drive for changes since that cursor. Unchanged Drive folders are not re-listed. File moves, renames and edits can update the snapshot without opening unchanged file bodies.
- The read-only preview uses the same Drive delta feed without persisting its cursor, so running a preview never consumes changes that the live sync must still process.
- Unknown folder structure, ambiguous removal (including possible loss of access), token invalidation, name collision or a periodic 24-hour reconciliation triggers a complete Drive inventory instead of guessing and potentially deleting data.
- A transient Drive API failure stops safely and leaves the last committed cursor untouched. NAS metadata still must be discovered independently.
- Cloud bidirectional conflicts use the newer timestamp as requested. First-run manual review remains an explicit optional preference.

## Important limitations
- Syncria now uses the Google Drive change cursor on live syncs and read-only previews, but must still discover NAS metadata. A local index does not make a recursive SMB scan instantaneous. Existing remote SMB notifications can trigger reconciliations; dependable event-driven NAS inventory skipping is not yet implemented. Periodic reconciliation remains essential to catch missed events.
- The existing `Thumbs.db` exclusions remain in force. Excluded files are not copied or deleted.
- A timestamp-preserving content edit cannot be discovered through timestamp-only comparison.
- Keep an independent backup of critical NAS/Drive data while testing the new index migration.
