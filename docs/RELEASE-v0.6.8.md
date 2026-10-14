# Syncria v0.6.8 — live preview scanning and duplicate Drive filenames

## Live scan progress

The **Test run · Preview changes** view no longer sits on a static message until everything has finished. It streams actual scan progress while Syncria inventories the source and destination:
- Current folder or filename being enumerated.
- Which storage location is being scanned.
- Discovered item counts for each location.
- A clear transition to comparison and the final proposed changes.
- A **Stop scan** control that closes the read-only request.

Progress is based on real objects encountered rather than a guessed percentage of your storage size. A large 89 GB dataset can take time to enumerate, but you can see what the scan is currently doing. All preview scanning remains read-only.

## Duplicate Google Drive file names

Google Drive permits multiple distinct files with the same name in one folder, including backup names such as `configs.xml.20260316-1302050218.backup`. A NAS cannot represent two sibling files with the same name.

Syncria now identifies colliding siblings and adds a reversible, Drive-ID-derived suffix to each NAS filename. The original Google Drive objects and names are not changed, and writing back to Drive restores the original display name.

If colliding siblings previously used a different NAS path, review the dry-run plan before enabling synchronization: a new mapping may appear as a new file or a removed legacy path. Syncria does not silently merge those objects.

## Existing behavior

Live status on the rest of the application continues as before. The optional preview does not copy, delete, or rename files. Existing sync decisions, recovery options and in-app updates are retained.
