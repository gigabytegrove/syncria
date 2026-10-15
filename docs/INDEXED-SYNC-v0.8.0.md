# Syncria Indexed Sync Engine — v0.8.0 development

## First-run index
On first start, Syncria creates `syncria-index.db` in its configured persistent data directory (`/data` in the standard container). It imports prior per-file checkpoints from `state.json` without reading file contents, retains a private `state.json.pre-index-v0.8.0` backup, and moves high-cardinality pair records out of the normal settings JSON. The embedded bbolt database is transactional, pure Go, and restricted to owner permissions.

A first-time sync inventories the chosen cloud/NAS locations and saves metadata for each reconciled file. Metadata includes paths, cloud identifiers, each side's size and modification time, and previously synchronized fingerprints when available. It does not hash file contents merely to compare changed/unchanged status.

## Later runs
The engine loads the DB index at start and compares provider metadata (file sizes/modification timestamps) against committed records, instead of reading old file contents. Unchanged files are silently ignored. The existing planning UI reports only skips, replacements, deletes and conflicts with reasons.

A successful file transfer commits its checkpoint before moving to the next transfer; initial unchanged metadata matches are committed in batches. If index writing fails, the sync stops rather than claiming work completed.

## Important limitations
- Syncria still needs to discover what changed remotely. The index makes *local comparison and persistence* fast but does not make a remote SMB recursive scan or Google Drive full-tree listing instantaneous. Google's changes cursor and reliable SMB notifications are separately tracked in Issue #12.
- The existing `Thumbs.db` exclusions remain in force. Excluded files are not copied or deleted.
- A timestamp-preserving content edit cannot be discovered through timestamp-only comparison.
- Keep an independent backup of critical NAS/Drive data while testing the new index migration.
