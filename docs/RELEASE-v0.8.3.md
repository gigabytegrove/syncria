# Syncria v0.8.3 — Recovery dialog and consistent connection layouts

## Recently deleted is no longer a huge inline list

A sync with pending recoverable deletions now shows a compact **Recently deleted** summary and its file count. Choose **Review & restore** to open a properly bounded, accessible native dialog.

- Search by filename or folder across all currently eligible recovery entries.
- Scroll a long list inside the dialog rather than expanding the sync card or entire page.
- See each exact path and deadline.
- Restore an individual file with the **existing** /api/recovery/undo action. Its security/authentication and expiration semantics have not been changed.
- Close with the button, backdrop click, or Escape; focus returns to the previous control.
- Recovery is also available from the individual sync's Overview tab, including when the main My Syncs view is hidden.

**Important:** Recovery eligibility is temporary. The existing ten-minute window is unchanged. The dialog does not allow bulk restore or automatically restore anything.

## Cloud Accounts and connection wizard polish

- Google Drive and Dropbox cards use consistent heights, layout, internal alignment, and full-width account-add footers.
- Accounts, Test, and Remove controls now line up cleanly with consistent sizes and spacing.
- The **Connect your cloud files** wizard has deliberate separation from the provider cards, correctly padded headers and steps, a narrower readable form, appropriately aligned buttons, and mobile-friendly spacing.
- Matched corrections for NAS connection form spacing.
- Desktop and narrow/mobile browser captures now include the **open Google Drive wizard** and the **open recovery dialog**, not just their closed states.

## What hasn't changed

- Sync engine, indexed comparison, watcher scheduling, paths, account authorization, conflict/deletion policy, stored run history, and backend recovery logic.
- Existing sync names, credentials, files and account mappings.
- The prior v0.8.2 sync-detail tabs and navigation.

## Test after updating

1. Back up the persistent /data directory, then install v0.8.3 in App Updates.
2. Open **Cloud Accounts**, connect an account wizard, and verify card/buttons/field spacing at desktop and phone widths.
3. Open **My Syncs**, then select **Review & restore** on any relationship with recent deletions.
4. Search a file path and verify the matching row and recovery deadline; do not click Restore unless you intend to undo that deletion.
5. Check **Open sync → Overview** for the same recovery option. Escape, Close, and backdrop should dismiss without changing a file.
6. Confirm normal syncs, previews and historical runs remain accessible.

Automated Go, race, static analysis, browser and UI checks verify the code. Your own NAS remains unchanged until you install the update.
