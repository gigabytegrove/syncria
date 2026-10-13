# Syncria v0.5.1 — Live updater progress and full UI refresh

**Primary update method:** In Syncria, open **Updates → Check again → Download & install update**. The updater runs entirely inside the existing Syncria installation; no extra Docker services or manual compilation.

This release focuses on the already-approved updater and v0.5.0 workflow, not new synchronization protocols or features.

- Replaced the old generic “Update staged. Syncria is restarting; refresh shortly.” form handler with the same tracked update process used by the main Updates view.
- Added a redesigned **Updates** workspace with exact running/target versions, file name, downloaded/total bytes, measured SHA-256 hash verification, elapsed server time, and durable timestamped events.
- Update phase progression and percentages are grounded in actual operations: indeterminate for non-measurable steps, measured percentages only for byte-oriented download, upload and SHA-256 verification. No arbitrary overall progress percentages.
- Displays detailed error states and loss of connectivity through restart. A release is shown as successfully installed only after the new running process reports the target version.
- Manual binary recovery updates are now tracked in the same status history.
- Cohesive updated styling across Overview, Sync relationships, Accounts, Storage, Activity, Updates and Settings, while retaining all existing functions and the approved Syncria branding.
- Clarifies direct SMB connection status (a configured native connection is not a Linux-mounted volume).

Your existing cloud accounts, sync relationships, NAS connections, storage indexes, conflict-resolution and 10-minute recovery settings remain in place.

**Note:** The page displayed by the *already-running v0.5.0 binary* may still show its older update message once when installing v0.5.1. Reload after installation to activate the v0.5.1 interface; subsequent releases use the new updater presentation.

Alternate installation methods, including standalone release binaries, Docker Compose and GHCR images (subject to package permissions), are retained.
