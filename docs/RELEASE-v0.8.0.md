# Syncria v0.8.0 — Indexed Sync (testing release)

## Highlights
- **Fast incremental Google Drive checks:** Syncria builds an initial metadata baseline, stores a per-relationship index in `/data/syncria-index.db`, and uses Google Drive's Changes API to retrieve changes on subsequent runs instead of listing every Drive folder again. Read-only previews also consult the index without consuming or advancing the saved change cursor.
- **Timestamp-first matching:** Sync decisions compare previously committed file sizes and modification times. Files that have not changed are kept out of the file-action lists and are not downloaded or hashed just to establish equality. Existing baseline metadata is imported without rereading whole files.
- **Durable checkpoints:** Verified transfers are indexed transactionally; an incomplete transfer never advances its file checkpoint. Existing `state.json` relationship metadata is imported and saved to a local bbolt DB with owner-only permissions.
- **Better first-run UX:** The progress area explains when a first baseline is being established, which may take longer than later checks.
- **Clear old failure:** My Syncs now has a **Dismiss old failure** button below an existing failure message. It clears the warning (including the "needs attention" banner) permanently, but preserves failed-run counts, file totals, and historical operation records. If the next run fails, a new error appears normally.
- **Cleaner previews:** Unchanged files remain hidden. Skipped items and files planned for replacement are shown with filenames and reasons, including the action filters from v0.7.2.
- **OS metadata excluded:** `Thumbs.db` and related thumbnail databases are excluded and are not deleted as a side effect of changed exclusion rules.
- **Conservative cloud change handling:** Ambiguous Drive removals, folder changes, case collisions, missing/expired tokens, and periodic reconciliation cause a full inventory rather than speculative deletes. Transient API failures leave saved checkpoints untouched.

## First install and backup — read before testing
v0.8.0 migrates file checkpoints from `state.json` to `/data/syncria-index.db`. The previous settings/checkpoint file is preserved in `/data/state.json.pre-index-v0.8.0` (if using the standard container data path). **Before installing, keep your own independent backup of the entire persistent Syncria /data volume and of important cloud/NAS content.** Leave scheduled syncs paused until a read-only preview looks correct.

**Automatic rollback to an older pre-index executable is intentionally blocked** after migration: older versions do not know about the new database and may treat files as untracked. To downgrade, stop syncing and restore the matching pre-migration `state.json` backup *together with* the old executable; do not simply replace the executable while keeping v0.8.0's lightweight settings file. Changes made after migration may not exist in the original backup. Seek help if the index is missing/corrupt; Syncria is designed to refuse syncing rather than rebuild an empty index.

## Performance and reliability limitations
This is a **testing release**. Automated tests, race tests, static analysis, and build verification are run in CI, but full end-to-end performance and interrupted transfers with the user's actual NAS have not yet been verified. Google Drive metadata can use change cursors; direct SMB destinations still require filesystem metadata enumeration or change event triggers and periodic reconciliation. An initial inventory can take considerable time; this is expected. Large transfers can still fail due to network/SMB interruptions, and resumption from an exact byte offset is not yet provided.

## Smoke test
1. Back up the persistent data directory and pause the sync relationship.
2. Install v0.8.0 via **App Updates**, then confirm **v0.8.0** appears in the interface.
3. Verify the existing relationship and its files are still present in **My Syncs**.
4. Use **Test run · Preview changes** and inspect the exact skipped/replaced paths and reasons; unchanged filenames should not appear.
5. Use **Dismiss old failure** to clear the historic Kaleb video deadline warning without resetting failed-run totals.
6. Start a controlled sync, then inspect **Activity**. Follow it with a second run with no deliberate changes: zero unchanged file transfers should occur.
7. Report any repeat reads, unexpected overwrites, or failures with the exact filename and timestamp. Do not enable automatic deletion until data behavior is confirmed.
