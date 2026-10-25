# Syncria 1.0 — End-to-end approved-folder receive (development)

## What is implemented

The native Go agent now supports the **complete read-only-source to explicitly locally authorized destination path** on Linux:

1. The source operator enrolls the peer, locally approves a source folder, and starts its read-only peer service with an owner-controlled token. The peer service requires a recent controller heartbeat and TLS outside loopback.
2. The receiver enrolls independently and locally approves a destination folder with `-read-only=false` (default grants remain read-only). The destination folder can limit paths using `-select` and `-exclude`.
3. `peer receive` authenticates to the source, inventories/requests the specified file, commits resumable SHA-256 verified chunks in a private receiver-side staging store, and then requests local publication.
4. The publisher reads only **completed verified transfers**, validates the destination approval, rehashes staged bytes, traverses directories using Linux handle-relative `openat` with `O_NOFOLLOW`, writes a private sibling temporary file, and uses atomic `linkat` to publish **only when the destination does not already exist**. If the name exists or a directory is a symlink, it fails closed. No deletion or remote overwrite operation exists.
5. Files are placed in the owner's selected destination. Verification manifests remain in the staging store until manually removed; a failed promotion preserves the verified staged file for review.

## Typical commands

On the source computer, grant the folder and initialize the peer token locally (only once):

```sh
syncria-agent grants add -name Photos -root /home/user/Pictures
syncria-agent peer init
syncria-agent peer serve -listen 0.0.0.0:9774 -tls-cert /path/peer.pem -tls-key /path/peer-key.pem
```

The two computers must already have been enrolled with the Syncria controller. Source peer keys must be shared out of band with the specific trusted receiver; never place keys on the command line or into a log. Agent grants are not silently created from the controller.

On a Linux receiving computer, configure the writable destination **locally**:

```sh
syncria-agent grants add -name Received -root /srv/received -read-only=false
syncria-agent peer receive \
  -peer https://source.example:9774 \
  -grant SOURCE_GRANT_ID -relative photos/image.jpg -source-id SOURCE_DEVICE_ID \
  -destination-grant RECEIVER_GRANT_ID
```

The receiver prompts privately for the source peer token. If interrupted during download, rerunning the command resumes verified transfer chunks. On success, the console prints the destination grant ID, byte count, and SHA-256.

`peer stage` remains useful for downloading only into private staging. `peer publish -destination-grant GRANT_ID -transfer TRANSFER_ID` atomically delivers a previously completed staging session, without network access. These operations require the locally enrolled agent identity.

## Explicit limitations and blockers

- **Linux destination publishing only.** Windows and macOS builds compile and fail closed for this command until platform-native race-resistant atomic no-overwrite writers are implemented and tested. Enrollment, staging and read-only source functionality retain their existing platform support.
- **Opt-in, user-initiated transfer:** this is a real file transfer, but is **not yet a scheduled, centrally managed Syncria synchronization job**. Job policies, conflict/version history and NAS/cloud adapters must still be integrated into the main engine. The existing NAS/cloud jobs are unchanged.
- Existing destination files are never replaced, even if contents match. There is no automatic propagation of deletions or renames, and no agent-controlled write API.
- Resumable staging consumes local disk space and is not an independent backup. Operators must manage completed staging sessions. Centralized retention and cleanup policies remain future work.
- A source peer token currently authorizes all locally granted read-only roots served by that peer. Per-relationship or per-root credential scoping and centralized credential lifecycle remain release blockers.
- The full Go 1.25 application, UI browser tests, platform runtime tests, installer upgrades, and production migration regression suite are still required. Do **not** publish a stable 1.0.0 release based on this pass alone.

## Tests

Run the package-focused suite:

```sh
./scripts/verify-agent-receive.sh
```

The tests exercise real loopback peer HTTP transfers, multi-chunk payloads, durable staging, verified final files, nested directory publication, replays, preexisting destination protection, race/collision attempts, revoked permissions, symlink escapes, read-only/selected folder policies, corrupted staging, canceled requests, and the native CLI receive command. On a Linux build agent, the file publisher itself is tested under Go's race detector.
