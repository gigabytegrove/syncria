# Syncria v0.8.1 — Fast recurring checks, single-flight scheduler and full run history

## Faster five-minute checks
- **Cloud change tokens and NAS notifications:** After an initial successful complete reconciliation, if all configured recursive SMB watchers have remained continuously connected and no NAS events arrived, a scheduled five-minute check asks Google Drive only for changes. When nothing changed, it avoids re-enumerating 5,000+ NAS files and records a fast-check timestamp instead.
- **Always reconcile when needed:** Any NAS notification or Google Drive change causes a real sync. Disconnected or missing `smbclient notify`, a fresh process without a verified watcher baseline, API errors, and a mandatory 45-minute reconciliation interval all require a full metadata inventory. Syncria never treats a failed probe as evidence nothing changed.
- **No repeated triggering:** Scheduler ticks (every 30 seconds), NAS events, and local filesystem notifications use the same one-at-a-time launch reservation. The existing atomic per-sync running flag remains as a second guard. A 45-second cooldown after completing a sync suppresses immediate self-generated NAS notifications.
- **Honest status:** Shows individual Google Drive and NAS scan counters, the latest successful quick-check timestamp, and whether NAS notifications currently qualify for fast checks. The counts no longer appear to restart mysteriously when progressing between endpoints.

## Full sync run history
- Click any sync name or **View sync details & run history** in My Syncs/Activity.
- Browse chronological run records with completion status, duration, transferred bytes and file-operation totals.
- Open any run to see actual copied/replaced/deleted/unsupported/failed file paths and readable operation reasons, with pagination. Records are kept in the persistent `/data/syncria-index.db`, not just the last 100 live events.
- The Active Syncs view shows transfers that actually started and operations that completed. A separate Test run / Preview remains for planning proposed changes; live sync no longer builds another speculative transfer plan every run.
- **History begins with v0.8.1.** Earlier v0.8.0 runs retain their counters and limited recent events but cannot be reconstructed into a complete file-by-file history retroactively.

## Safety and limitations
- Existing source/destination timestamps and file indexes remain intact; no sync configuration is reset. The database gains a run-history bucket, so back up the persistent `/data` directory before upgrading.
- This does not make a first scan or the mandatory periodic NAS reconciliation instantaneous. SMB monitoring must be functioning continuously; if unavailable, Syncria deliberately falls back to metadata scans rather than silently miss NAS edits.
- Changes to watched NAS paths are reflected as notifications, but network loss, missed events, and startup gaps are covered by the full reconciliation interval. A full scan is required before enabling the fast path after watcher reconnection.
- The one-way/bidirectional direction and **Propagate deletions** settings are unchanged by this update. With deletion propagation enabled, previously synced files deleted on Google Drive can cause the NAS counterpart to be staged for deletion, even in one-way cloud → NAS relationships.
- The per-sync atomic guard protects a running server and the pending-launch reservation prevents duplicate scheduler/watch triggers. Two independent installations with **different** Syncria data directories pointing at the same endpoints require external coordination. Do not configure both as active writers.
- The index change from v0.8.0 remains in place; older pre-index binaries cannot safely use a v0.8.x database. Follow the v0.8.0 backup/rollback guidance.

## Test on NAS
1. Keep the sync paused while upgrading and back up `/data`.
2. Verify v0.8.1 in the interface, then choose your sync and review the deletion settings.
3. Start a controlled sync and confirm the actual source and NAS metadata phases appear and a single run completes.
4. Allow the next scheduled five-minute check with **no changes**. With healthy NAS watchers and valid Drive cursor, it should use the quick check without launching a full NAS walk.
5. Create one safe test file, confirm one triggered run and an accurate file-operation record; check run history survives browser refresh/restart.
6. Disconnect the SMB watcher or test with unsupported notifications and confirm Syncria falls back to a complete metadata inventory, not a claimed fast no-change check.
