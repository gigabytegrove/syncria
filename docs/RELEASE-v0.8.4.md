# Syncria v0.8.4 — Enforce the ten-minute recovery deadline

## Fixed: Expired files were still listed with Restore buttons

Syncria's 10-minute recovery window has always been enforced by the restore API, but the UI previously displayed every stored pending-deletion record until the background purge succeeded. If the NAS or Google Drive was unavailable when cleanup ran, an entry deleted hours earlier could still show **Restore file**, even though restoring it would be rejected.

v0.8.4 separates **recovery eligibility** from **cleanup of staged data**:

- A file is eligible only while its absolute recorded `Expires` time is strictly in the future.
- The server filters expired records from My Syncs and Sync Details before rendering, including on the first page load after upgrading.
- An already-open browser also removes the expired file row **and its Restore POST form** at the deadline, updates the remaining eligible count, and hides empty recovery summaries. Rechecking when the tab becomes visible protects against sleeping background tabs.
- The backend already refuses an expired Undo request; this release preserves that protection.
- Expired staged files remain in the internal cleanup queue when a remote purge fails. Syncria retries cleanup as before without extending the 10-minute Undo window.
- **No change to the 10-minute recovery duration, deletion policies, file transfer logic, or sync scheduling.**

## Test cases

Regression tests cover last-night deletions, an exact expiration timestamp, mixed unexpired/expired entries, UI rendering with failed cleanup, a direct stale Undo request, removal of Restore forms in open browsers, and preservation of the remote cleanup queue.

## Updating

Install v0.8.4 from Syncria **App Updates** once published. Reload My Syncs: files deleted more than 10 minutes ago should no longer appear under Recently Deleted even if their cleanup is pending. The per-sync recovery panel will only appear if at least one file remains eligible.

If the old warning persists after updating, confirm the footer displays v0.8.4 and refresh the page to load the new interface.

This is a focused maintenance patch to the v0.8.3 baseline, not a new interface or sync-engine redesign.
