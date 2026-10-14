# Syncria v0.6.9 — make the first sync choose the newer copy

Syncria now handles files already present on **both** sides of a new cloud-to-NAS relationship more intelligently.

## First-run comparison

Previously, Syncria labeled every shared file as **Review: both sides have an untracked file**, even when one copy had clearly been modified much later. A matching first-run file also triggered the cloud relationship's implicit manual conflict rule.

Now:

- A first-time cloud-to-NAS match compares both file modification timestamps.
- When one file is clearly newer, the preview identifies **which location wins**, the intended direction and the timestamp of each copy.
- The preview states **Replace if contents differ**, because it is a read-only metadata check. During the real sync, Syncria checks actual file content first and does not overwrite identical copies.
- If timestamps are missing, identical or less than two seconds apart, Syncria does **not** guess. It flags the file for manual resolution.
- File-level overrides and the configured one-way NAS rules continue to take priority. Existing NAS-to-NAS relationships retain their conflict policy.
- The first-run policy is separate from subsequent two-sided conflicts. Existing cloud relationships historically had an automatic manual policy that should not have forced thousands of first-run reviews. This release corrects that without recreating the relationship.
- The cloud sync wizard now offers **Use the file with the newer modification time** (default) or **Ask me which copy to keep**.

## Less work with large folders

For regular Google Drive files that have an MD5 checksum, first-run content checks can hash the NAS copy and compare it to Drive's checksum rather than downloading both file streams. Different file sizes skip the expensive content read. Dropbox's special content hash is *not* treated as a normal MD5/SHA256 value.

For an 89 GB collection, this reduces unnecessary Drive traffic, but comparing a large set of files may still take time.

## Safety

Modification time is evidence of recency, not a guarantee that a file contains the correct version. NAS copies sometimes acquire fresh timestamps when moved or restored. Review the **Test run · Preview changes** decisions carefully before your first live run. Actual changes use the same winner rules as the preview, with content checks before replacing files.

All existing account connections, bidirectional sync, deletion recovery, live activity and offline-bundled updater behavior remain intact.
