---
id: KNOW-002
type: knowledge
status: active
sources:
  - path: ai/raw/threat-model.md
    fingerprint: sha256:eb2640fffaeed55d7a2cbd2a9a40c33654720232ff7fed98470cf6985b238f5d
---

# KNOW-002: Security Threat Model and Trust Boundaries

## Summary

The threat model defines the assets, adversaries, in-scope attack vectors, out-of-scope boundaries, and security assumptions governing GoVault. The primary security objective is ensuring that an adversary who obtains the local SQLite database, encrypted backup files, configuration, or application logs cannot recover protected secrets without the user's master password.

## Known Facts

- Primary Security Goal: An attacker who obtains `vault.db`, `.gvault` backups, config files, or logs cannot decrypt secrets without the master password or active decrypted key material.
- In-Scope Adversaries and Attacks:
  - Offline attacker with disk/database/backup artifact access (lost laptop, stolen backup, backup sync leak).
  - Offline brute-force and dictionary attacks against master passwords.
  - Forensic analysis of database remnants, deleted record tombstones, and SQLite WAL files.
  - Clipboard snoopers (mitigated by automated timed clearing).
  - Terminal shoulder-surfing / accidental terminal history leakage (mitigated by default masking and avoiding passing passwords as CLI command arguments).
- Out-of-Scope Adversaries and Attacks:
  - Full host compromise / root access / kernel-level compromise.
  - Active keyloggers and compromised terminal emulators.
  - Direct live process memory inspection/dumping of an unlocked running GoVault process.
  - Hardware side-channels (Spectre, Meltdown, cold boot attacks).
  - Compromised Go runtime or compromised compiler toolchain.

## Constraints

- Storage encryption alone (e.g. SQLite encryption extensions or full-disk encryption like LUKS/FileVault) is not sufficient; application-level authenticated encryption is mandatory.
- Master passwords must never be stored on disk in plaintext or as reversible hashes.
- Secrets must not be passed via CLI command-line arguments (which appear in `ps` process lists and shell histories); use stdin, prompts, or secure environment variable inputs.
- Vault database files and backups must have restricted filesystem permissions (e.g., `0600` on Unix).

## Unknowns

- Feasibility of best-effort memory zeroization in Go given garbage collection, heap allocations, and string immutability (accepted as best-effort defense-in-depth).

## Conflicts

- None.

## Provenance

- Security goals, attacker models, trust boundaries, and mitigation strategies derived directly from `ai/raw/threat-model.md`.

## Related Topics

- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture
- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Storage Format
- [KNOW-005](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-005-backup-format.md) — Backup Format
