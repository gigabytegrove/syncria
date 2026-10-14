# Syncria v0.7.1 — Active sync controls and transfer activity

## What changed
- Added a **Stop current run** control for each running sync, available on **My Syncs** and **Activity**. It requests cancellation of the running job without deleting the sync configuration or disabling the schedule. A separate Pause automatic sync action remains available when the user wants to prevent future scheduled runs.
- Added a dedicated **Active syncs** section on **Activity**, showing scans, planned file operations, current file, actual transferred bytes, and completed operations.
- The file plan is generated using the metadata already scanned for the sync, not an additional NAS or cloud scan and not an expensive full-file checksum pass.
- Shows up to 120 pending file actions, with the total pending count and a note when the list is truncated. Long paths remain readable. The current file is always displayed independently of the queue limit.
- During scanning, the discovered-file count is displayed without the misleading `scanned / 0 total` denominator.
- Live updates target the current status fields and queue entries, preserving existing page controls rather than redrawing the page.
- User-requested stop is distinguished from an unexpected failed run in saved results.

## Upgrade notes
The existing v0.7.0 timestamp-first comparisons and streaming timeout fixes are retained. The update does not automatically pause running jobs or reset saved checkpoints. Installing an update may still require the existing running sync to end, so use Stop current run first if necessary.

## Safety and limitations
Canceling is cooperative: a driver that is blocked inside a network read may not stop immediately. Syncria preserves completed destination files, and the SMB transfer layer stages incomplete files in temporary paths. Stop current run does not promise offset-resumable 30 GB partial transfers; resumable transfers remain separate work.

The active file queue is an advisory plan, not an immutable promise: live files may change after scanning, and conflict/deletion protection rules still apply. The queue is limited to 120 rows for performance; the file currently being processed is always displayed. No full-file checksum is performed solely to populate the queue.
