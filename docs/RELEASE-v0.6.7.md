# Syncria v0.6.7 — clearer updates and protected Google-native files

## Self-update handoff
After you initiate an update, Syncria remembers the expected version and waits for the replacement process to report that version as running. Once verified, the browser automatically opens the updated application. The redundant **Open updated Syncria** button has been removed.

Normal background checks and sync status updates do not reload the page. A failed update does not trigger an automatic reload.

## Google Earth, My Maps and other Google-native items
Google Drive contains virtual Google-native items which are not ordinary downloadable files. Some cannot be exported in a supported, lossless format.

These unsupported items now appear in scans without stopping the entire relationship. They are marked as **skipped** in the preview and recent activity, with an explanation. Existing originals on Drive are untouched; the connector does not attempt to upload, overwrite or propagate deletion for those items.

**Important:** Skipped Google-native items are **not backed up to the NAS**. Do not mistake an otherwise successful sync for a complete export of Google Earth/Maps project data. Export must be implemented and verified separately for types that support a trustworthy format.

Regular files and supported folders continue syncing under the existing conflict and recovery rules.
