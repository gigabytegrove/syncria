# Syncria v0.8.6-alpha.3 — Connected Computers and Verified Channel Upgrades

**Current recommended Alpha preview.** Stable v0.8.5 is unchanged. This is a corrective superseding release of alpha.2 with the same separate computer-management and transfer-engine functionality.

## Everything included in the alpha.2 feature preview

- **My Computers:** organized inventory, live online/offline checks, shared-folder details, three-step Windows/Linux agent onboarding and securely verified pairing.
- **Computer Transfers:** dedicated workspace, dual file browsers, filename search, up to 100-file copy/move batches, queue, persistent history, actual receiver bytes, cancel, retry, clear record and job status filters.
- **Independent workers:** each copy or move runs in an isolated native OS process, maximum three at once. Peer-to-peer data streams directly between the source and destination agents; the hub manages job control only.
- **Transfer safety:** encrypted certificate-pinned HTTPS, scoped short-lived tickets, no-overwrite destination publishing, exact size and SHA-256 checks before any move-source deletion.
- **Resilient transfers:** persistent 0600 journal under `/data/runtime`. Jobs interrupted by a hub/container restart are flagged and require explicit retry; no unsupervised replay.
- Windows/Linux native agents are built from the Syncria executable; no SMB/NFS shares are required.

## Corrected in alpha.3

**Alpha/Beta updater restart confirmation:** GitHub prerelease tags intentionally have `alpha-` and `beta-` safety prefixes so older Stable clients cannot offer experimental builds. Previously the post-restart verifier compared the full GitHub tag (e.g. `alpha-v0.8.6-alpha.2`) against the executable version (`v0.8.6-alpha.2`) and could report that a successful installation failed. The new executable and web UI now compare the normalized version while preserving the original signed-release tag for discovery and checksum downloads.

This also corrects the first upgrade from alpha.1, because the *new* executable performs the confirmation after the supervisor restart.

## Update

If you are already running **alpha.1 or alpha.2**, open **App Updates**, choose **Alpha**, save, check for updates and install **v0.8.6-alpha.3**. The prior alpha can incorrectly mark restart verification as failed, but the new alpha.3 binary recognizes the prefixed tag correctly. Refresh the browser after restart if the old page remains.

If you are still on **v0.8.5 Stable**, opt in manually using **App Updates → Advanced options → Manual installation** with this release's matching Linux executable and `SHA256SUMS`, then select the Alpha channel after restart.

Back up the persistent Syncria `/data` volume before crossing channels. No GitHub container-image replacement is required; the existing supervisor activates `/data/runtime/current`.

See [Connected Computers](CONNECTED-COMPUTERS.md) and the detailed [alpha.2 feature notes](RELEASE-v0.8.6-alpha.2.md).

## Limitations

Folder-recursive transfer, resume within a partially copied file, unattended agent auto-update, WAN relays and automatic bidirectional folder synchronization remain future work. This prerelease handles regular-file transfer batches, not full LANHub feature parity.

## Verification

Before publication, this release must pass Go tests (including a channel-prefixed version recovery regression test), race tests, vet, UI build, seven platform binaries with SHA256SUMS, and live copy/move/no-overwrite integration using two HTTPS agents and a separate native worker executable.
