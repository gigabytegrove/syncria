# Syncria v0.6.10 — FreeFileSync migration safety

FreeFileSync (`sync.ffs_db`) is not part of the user's data collection and should not be synchronized by Syncria. Existing copies are left untouched while FreeFileSync is still in use.

## Fixed
- Ignore the exact filename `sync.ffs_db`, case-insensitively, at any depth in Google Drive, Dropbox, NAS, and local folder scans. Other filenames remain eligible.
- Do not convert an old FreeFileSync database entry in Syncria's saved sync index into a deletion instruction. The index entry is quietly dropped on the next real sync; the source file stays where it is.
- The dry-run planner omits old and current FreeFileSync metadata entries, even if they exist in legacy checkpoints.
- Label previously saved errors as **Last sync failed** with their run date. These historical failures are hidden while another sync is running and no longer appear as if the current scan had just failed.
- Add regression tests for nested and case-varied FreeFileSync database names, local inventory, historical checkpoint safety, and the error display.

## Important compatibility note
The exact message `Primary scan: duplicate/case-colliding remote name in : sync.ffs_db` comes from the old v0.6.3 scanner, not the current v0.6.9 source. The previous error may remain stored as the last unsuccessful run. Syncria v0.6.9 already includes the Google Drive and NAS sidecar exclusion; this release extends the same rule to the remaining paths and avoids stale failures being mistaken for live errors.

No FreeFileSync databases or other user files are deleted or renamed by this change. The active sync engine, dry-run preview, timestamp selection, recovery controls and in-app updater remain unchanged.
