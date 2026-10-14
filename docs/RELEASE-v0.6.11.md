# Syncria v0.6.11 — Edit your syncs and choose a direction

Syncria now lets you change an existing **Google Drive ↔ NAS**, **Dropbox ↔ NAS**, or **NAS ↔ NAS** relationship without deleting it and building a replacement.

## Edit an existing relationship

On **My Syncs**, select **Edit sync**. The editor loads your current name, source and destination folders, accounts, schedule, conflict preference, delete-and-recovery settings, and (where applicable) NAS sharing options. Make changes and save. No file transfer starts just because you saved settings.

**Important safety rules:**
- Pause the relationship before saving edits, and wait for any running sync to finish.
- Changing just the name, schedule, direction, or conflict/deletion rules **preserves the existing file comparison checkpoints and run history**.
- Changing the source account or folder, NAS source, destination folders, or destination order requires an explicit acknowledgment. The old comparison checkpoints are reset so old paths cannot generate incorrect overwrites or deletion instructions for new locations. Existing user files are not deleted by saving.
- Pending deletion recovery must be resolved before changing endpoints, so an in-flight restore cannot lose its association with the original folder.
- A relationship cannot change between cloud-to-NAS and NAS-to-NAS types; use a new relationship for that different type.

## Choose direction for cloud-to-NAS

Cloud relationships now include a visible choice at creation and editing:

- **Both ways — Cloud ↔ NAS** (the existing default). Eligible changes can travel in both directions.
- **One way — Cloud → NAS**. The cloud is authoritative: NAS-only files are not uploaded, changed cloud files copy down, and an existing NAS copy whose contents differ may be replaced by the cloud version. This includes source-wins behavior when both versions differ during first setup. Review a dry run before first enabling it.

One-way cloud relationships cannot also enable propagation of NAS-side changes between multiple NAS boxes through the cloud.

## NAS-to-NAS behavior stays separate

The NAS-to-NAS editor retains its existing one-way and bidirectional modes and the configured conflict policy. Existing one-way NAS protections against overwriting independently changed destinations remain unchanged.

## More visible status

The sync card now identifies the direction, and the existing read-only preview uses the newly saved direction. Actual syncs use the same direction rule. Tests cover paused edits, saved checkpoints, changed endpoints, and an end-to-end one-way cloud transfer without an unintended reverse upload.

## Still outstanding

This release does **not** claim that the earlier first-run 5,060-item **Review** result is solved. Equal-timestamp files still need content comparison and a trustworthy identical-file classification in the **dry run**; that larger audit remains separate. No claim is made that an entire 89 GB collection was synchronized end-to-end in CI.
