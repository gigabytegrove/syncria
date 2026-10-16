# Syncria v0.8.6-alpha.1 — Connected Computers and Update Channels

**Experimental alpha preview — not a stable release.** The stable production baseline remains **v0.8.5**. This release is opt-in for testers who understand that alpha builds can contain bugs. Back up the entire Syncria `/data` volume before manually moving from Stable to Alpha, especially when using indexed synchronization.

## What's included

- **Update channels:** Stable, Beta and Alpha are selectable from App Updates. The choice persists in Syncria's existing state file and survives restarts.
- **Stable:** Only fully published stable releases. No alpha/beta/RC builds are offered.
- **Beta:** Published stable, beta and release-candidate builds. Alpha builds are excluded.
- **Alpha:** All published releases, including experimental alpha builds.
- **Safe release discovery:** Read GitHub's published Releases API, not arbitrary tags. Ignore draft releases, malformed versions, and releases with inconsistent prerelease flags. Validate the OS/architecture binary and `SHA256SUMS` before offering an update.
- **Correct precedence:** Alpha → Beta → RC → Stable for the same semantic version. The updater never downgrades just because a channel changes.
- **Legacy safety:** Experimental release tags are prefixed with `alpha-` (or `beta-`). Syncria v0.8.5 and earlier do not recognize these tags as valid stable versions and therefore cannot accidentally offer this alpha update.
- **Connected Computers:** Native Windows/Linux agents with explicitly allowed folders; HTTPS certificate pinning, authenticated pairing, direct file browsing, and peer-to-peer copy/move of individual files.
- **Direct transfer integrity:** Receiver checks SHA-256 and byte count before publishing new files; existing files cannot be silently overwritten. The sender's original is removed on a move only after verification.

## How to opt in from Stable v0.8.5

The existing v0.8.5 user interface cannot yet select a channel because the feature is new. It also deliberately does **not** offer alpha releases in its normal update search.

1. Download this alpha release's platform binary and `SHA256SUMS` from the release page.
2. Back up the Syncria data volume and, in v0.8.5, open **App Updates → Advanced options → Manual installation**.
3. Upload the matching alpha binary and `SHA256SUMS` to Syncria's existing verified manual updater.
4. After the new version restarts, open **App Updates**, choose **Alpha** and select **Save channel**.
5. Subsequent updates are discovered from the saved channel automatically. Switching to Stable does **not** force an unsafe downgrade to an older stable version.

The native agent executables use `-agent` mode as documented in [Connected Computers](CONNECTED-COMPUTERS.md). They do not require SMB/NFS access or a second Docker container on the hub.

## Release publishing conventions

The tag for this alpha is **`alpha-v0.8.6-alpha.1`**, while the version reported by running Syncria is **`v0.8.6-alpha.1`**. This deliberately prevents old tag-scanning clients from treating the experiment as Stable.

Future conventions:
- Stable: `v0.8.6` with `prerelease=false`
- Beta: `beta-v0.8.6-beta.1` with `prerelease=true`
- Release candidate: `beta-v0.8.6-rc.1` with `prerelease=true`
- Alpha: `alpha-v0.8.6-alpha.1` with `prerelease=true`

Always attach the entire platform executable set and `SHA256SUMS` to a published GitHub Release. A GitHub tag alone is not an installable update. Draft releases are excluded from every channel.

## Limitations

Alpha includes individual-file agent copy and verified move only. Automated agent sync relationships, whole-folder transfers, WAN relays and production-ready agent lifecycle management remain future development. This alpha should not be deployed to irreplaceable data without independent backups.
