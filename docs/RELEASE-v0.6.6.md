# Syncria v0.6.6 — Live sync visibility

Syncria now shows actual synchronization activity as it happens. The read-only test run remains available, but it is no longer the main way to learn what the sync engine is doing.

## Live across the application
- A persistent status strip shows running relationships, scan or transfer stage, the current file or folder, and progress.
- The Activity page includes a system-wide live feed of actual file operations, including direction and the relationship involved.
- Drive and NAS scans report names and discovery counts while folders are enumerated.
- The dashboard, sync cards and Activity share the same status stream. An authenticated, local Server-Sent Events connection refreshes every second; a polling fallback is used where reverse proxies block streaming.
- Each sync saves up to 100 actual file operations from its most recent run, once at completion, for review after a restart. Syncria does not write the main config for each file.

## Google Drive filenames containing slashes
A single Drive item can legally be named `Bland, Mike - Security Camera/P2P Setup`. A NAS cannot store a literal slash in one file or folder name.

Syncria now translates names requiring it into unique, reversible NAS-safe names, and restores their original Drive names when creating or updating cloud items. A collision on the destination remains a safe-stop error; no file is silently discarded or overwritten.

The exact FreeFileSync database name `sync.ffs_db` remains excluded as a synchronization sidecar.

## Read-only test runs
**Test run · Preview changes** remains an optional, non-destructive preflight check. Actual live status is available without opening or running a preview.

## Safety and compatibility
Existing conflict decisions, one-way versus two-way rules, deletion recovery, self-contained in-app updates and NAS/cloud connections remain supported. Updates do not require another container or installation service.
