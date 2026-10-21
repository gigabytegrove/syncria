# Syncria 1.0.0 — Capture-time NAS version integrity

**Development only. Approved scope:** destination-owned file versioning/recovery for agent-to-NAS jobs. Existing cloud/NAS and NAS/NAS behaviors are unchanged.

## Preservation contract

Before a verified agent-to-NAS replacement, `nasversion.PromoteVerified` moves the previous NAS file into its own uniquely named NAS archive directory, computes SHA-256 of the archived file with a bounded streaming buffer, and writes/re-reads an adjacent `.syncria-proof.json` containing the job ID, archive ID, original relative path, length and checksum. The replacement is published only after that capture record has been verified. Capture failure attempts to restore the previous live name. If rollback also fails, the old archive location is reported for operator recovery. Controller storage holds metadata, not file payloads.

The read-only archive list reports `capture_proof_present` and `capture_sha256`; **proof presence is not a claim that the entire archived file was re-hashed while listing**. The central restore API now calls `RestoreCopyVerified` and rejects legacy or missing-proof archives, so removing a new archive's proof cannot downgrade to weaker verification. The archived data is compared to its capture-time SHA-256 before publishing a separate recovered filename. The compatibility primitive `RestoreCopy` still supports explicit manual legacy recovery but returns `capture_verified=false`.

Retention remains two-step: inspect an exact plan, then explicitly confirm its digest. New checksum-bearing archives are re-hashed before deletion, and their proof and empty directories are removed after payload deletion. An interrupted cleanup may leave an empty directory or proof-only directory, never grounds for deleting a live file.

## Security and alpha blockers

These are SHA-256 integrity records **not digital signatures**; an attacker who controls both archived content and metadata can forge both. SMB share rename semantics, power-loss recovery, same-name concurrency, actual NAS disk behavior, and performance on multi-gigabyte files still require testing. The full Go 1.25 controller, cloud/NAS regression suite, installation and upgrade/rollback tests are pending. This is not a production-ready version or formal alpha release.

## Verification

```bash
go test -race -count=3 ./nasversion
go vet ./nasversion
```

The tests cover capture-time digest creation, verified restore, same-size/same-timestamp tampering, missing/corrupt proof, failed proof capture and rollback, preview-bound retention, orphan cleanup, and legacy compatibility.
