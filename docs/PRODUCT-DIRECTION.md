# Syncria — Locked Product Direction (2026-10-20)

## Product identity
Syncria remains a self-hosted **file synchronization engine**. Its goal is to keep designated files and folders synchronized across approved endpoints (cloud, NAS, and optionally user devices) while preserving file integrity, privacy and recoverability. Syncria does not become a cloud-storage provider, unrestricted remote file browser, remote desktop or general endpoint-management system.

## Current operational baseline
User reports several days of reliable operation and working large-file transfers. Do not continue to label large files as a currently reproducible blocker without new evidence. Persistent bbolt indexing and Google Drive change-cursor support already exist; do not design a redundant second index. Keep indexing efficiency, correctness after restarts, and SMB/NAS enumeration costs under observation and verify with measured real-world runs. Interrupted transfers/resume, safe rollback, and change-inventory optimizations are still separate acceptance dimensions.

## Optional device agents — extend existing sync engine
Allow optional agents on Windows, Linux and (where supported) macOS, with Android later as a native device endpoint. Agents use the **same sync-job and destination model**, indexing principles, scheduler, operation history, conflict handling and safety policies as NAS/cloud relationships. Do not build an unrelated file-management/control-plane product.

- Enrollment through explicit pairing approval, revocable per-device credentials and scoped access; easy installation and safe disconnection.
- Expose only user-approved folders, never whole drives by default. Make granted permissions, online/offline status, last-seen, capacity and pending work visible.
- Support selected-folder computer-to-NAS and computer-to-computer sync. Optionally provide one-off authorized **Send file** transfers that do not create a permanent sync.
- Prefer secure, authenticated, encrypted transport and enforce path boundaries, resource limits, device ownership, audit events and least privilege.
- Preserve verified checkpoints; resume/retry safely after sleep, intermittent links and restart without rewriting healthy destinations or misidentifying an offline folder as deleted.
- Keep unattended sync and user-initiated transfer actions distinct. Support bandwidth/scheduling controls where beneficial.
- Updates should use verified builds and maintain compatibility, with agents functional on LAN without a cloud account. Any remote-network connectivity is optional and must not weaken authentication.

## Android scope
A full Android management application is **not required** for Syncria's existing NAS/cloud goals. A future **optional native Android companion/agent** can contribute approved phone photos or folders as sources and targets, offer transfer initiation/status, and provide helpful notifications. Respect Android file permission APIs, background execution, charging/network policies and OS-managed scheduling. Do not promise instantaneous background syncing. The existing browser UI remains the primary management interface.

## Useful additions consistent with the core
Selectively synced folders, meaningful sync-health alerts, file version history where endpoints support it, bandwidth/schedule policies, better conflict resolution, device-to-device transfer, robust large-file interruption recovery and improved NAS change discovery are in scope when they reuse the same engine and earn actual acceptance. Never claim a feature finished solely because a PR exists.

## Guardrails
No unrestricted access to a user's computer. No hidden file collection. No automatic replication of an entire endpoint. No silent destructive deletion on partial/offline scans. No replacement of the existing cloud/NAS synchronization foundation; additions are opt-in. A sync is not an independent backup, so retain accurate deletion and restore warnings.

## Acceptance
Tests must cover permission isolation, revocation, offline/reconnect, safe file operations, conflicts, resume, updates, indexed state and performance on real NAS samples. Keep device agents optional. This document records approved direction, **not a claim that all proposed features have shipped**.
