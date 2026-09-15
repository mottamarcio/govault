---
id: SPEC-012
type: spec
status: ready
parent: FEAT-004
depends_on:
  - SPEC-004
  - SPEC-005
  - SPEC-006
  - SPEC-008
  - SPEC-009
  - SPEC-011
supersedes: []
---

# SPEC-012: Backup Creation, Restore & Merge Application Service

## Intent

Specify the `BackupService` application workflows: atomic backup file creation, export with custom export passphrase vs master password, dry-run backup inspection and verification, recovery snapshot creation, full vault restoration into temporary SQLite database with atomic swap, and non-conflicting entry merge import.

## Requirements

- **R1: Atomic Backup Creation & Custom Passphrase Export:** Implement `BackupService.CreateBackup(ctx, outputPath, opts)`:
  - Collect logical entries, tags, and histories from storage repositories.
  - Support encrypting backup with active vault key / master password OR re-wrapping under a custom one-time export passphrase with fresh Argon2id salt and KDF parameters.
  - Write backup to a temporary file (`.gvault.tmp-*`) with `0600` permissions and perform atomic rename to target path upon successful completion and verification.
- **R2: Structural & Cryptographic Backup Verification:** Implement `BackupService.VerifyBackup(ctx, backupPath, password)`:
  - Perform unauthenticated structural checks (header integrity, magic, version, bounds).
  - If password is provided, perform full cryptographic verification (unwrapping key, checking AAD, decrypting payload, verifying manifest counts vs parsed item counts, and checking for duplicate IDs or broken history references).
- **R3: Full Vault Restore with Atomic Swap & Safety Snapshot:** Implement `BackupService.RestoreBackup(ctx, backupPath, password, opts)`:
  - Pre-validate backup cryptographically before modifying any database state.
  - Automatically create a timestamped recovery snapshot of the existing vault prior to restore.
  - Re-encrypt and insert logical entries, tags, and histories into a temporary SQLite database with fresh per-record nonces.
  - Atomically replace the active vault database upon validation of the restored database.
- **R4: Merge Import Strategy:** Implement `BackupService.MergeBackup(ctx, backupPath, password, conflictStrategy)`:
  - Decrypt backup payload and identify conflicting entries (matching IDs or titles/aliases).
  - Support conflict resolution strategies (`skip`, `overwrite`, or `rename`) to safely import records and tags into an existing unlocked vault without replacing the entire vault database or lineage.

## Acceptance Scenarios

- **Scenario 1: Custom Passphrase Export and Independent Restore**
  - *Given* an unlocked vault and a distinct one-time export passphrase
  - *When* `CreateBackup` is executed with the custom passphrase
  - *Then* the resulting `.gvault` file can be decrypted and verified using the export passphrase alone on another machine without the original master password.
- **Scenario 2: Full Restore with Pre-Existing Vault**
  - *Given* an existing vault database with records A and B, and a backup file containing records C and D
  - *When* `RestoreBackup` is executed with the correct backup password
  - *Then* a recovery snapshot of the current vault is created first, and the active vault database is atomically replaced with the restored vault containing records C and D.
- **Scenario 3: Corrupted Backup Aborts Restore Without Altering Database**
  - *Given* a tampered `.gvault` backup file
  - *When* `RestoreBackup` is executed
  - *Then* restoration aborts during pre-validation, no database tables are modified, and the original vault remains intact.
- **Scenario 4: Merge Import with Conflict Resolution**
  - *Given* an active vault with record "Github" and a backup containing "Github" and "AWS"
  - *When* `MergeBackup` is executed with strategy `skip`
  - *Then* "AWS" is imported into the vault while the existing "Github" record is preserved unmodified.

## Edge Cases

- Attempting to overwrite an existing destination backup file fails unless `Overwrite` option is explicitly set.
- Restore onto an incompatible or locked database handles locks gracefully and cleans up temporary files.
- Backup containing orphaned history records (referencing non-existent `entry_id`) is rejected during restore validation.

## Constraints

- Temporary backup files and databases must use `0600` file permissions.
- Restore must never stream unauthenticated data into the live database.
- All temporary restore artifacts must be cleaned up on failure.

## Non-Goals

- Remote/cloud network backup upload or streaming.
