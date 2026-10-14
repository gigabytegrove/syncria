# Syncria v0.6.4 — preview before syncing

## Fixes

**FreeFileSync metadata no longer blocks Google Drive syncs.** Syncria excludes files named exactly `sync.ffs_db` from replication, on both Google Drive and NAS scans. These are FreeFileSync's comparison databases and can appear multiple times within the same cloud folder. They are not deleted or modified at the source. Other duplicate or case-colliding filenames still stop the scan safely with a clear explanation.

**Preview any existing cloud-to-NAS or NAS-to-NAS sync without changing files.** Open **My Syncs → Test run · Preview changes** to see:
- Individual file/folder names and planned actions.
- Whether a file moves from Google Drive, Dropbox, or a named NAS to another location.
- Which side would win an overwrite based on the configured conflict policy.
- Deletions, skips, missing folders, and conflicts needing review.
- A summary of changes and important safety warnings.

The preview is read-only and never initiates an upload, deletion, or folder creation. It displays up to 500 rows and counts all items. For untracked files already present on both sides, the preview notes that a real content comparison is still needed.

**See what actually happened.** Live sync status continues to show the currently processed item, and each relationship now retains the last 100 real file actions from the most recent run, including the origin and destination when applicable.

**Small layout repair.** The sidebar's status text no longer overflows beneath the logo.

## Existing functionality

Normal sync direction rules, multi-NAS relationships, manual conflicts, 10-minute deletion recovery and storage connections are preserved. The updater remains self-contained and platform-agnostic.
