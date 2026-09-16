---
id: FEAT-007
type: feature
status: active
parent: PRG-001
---

# FEAT-007: TUI First-Run Setup & Interactive Vault Initialization Wizard

## Capability

Automatic detection of uninitialized or missing vault database files upon launching the TUI, seamlessly presenting a dedicated interactive First-Run Setup Screen (`ScreenInit`) to guide the user through establishing and confirming their initial master password, generating the cryptographic envelope, running database migrations, and transitioning directly into the unlocked vault dashboard without requiring manual CLI `govault init` pre-configuration.

## User Value

New users and users opening fresh vault paths experience a welcoming, intuitive onboarding flow instead of a confusing "invalid password" prompt against an uninitialized database. Users receive clear warnings regarding master password non-recoverability, real-time password strength feedback, and automatic session progression upon setup completion.

## Scope

- **Vault Initialization Detection:**
  - Inspect vault database path and metadata table state during TUI initialization.
  - Automatically route to `ScreenInit` when no vault metadata is found (or database does not exist), and to `ScreenUnlock` when an initialized vault exists.
- **Interactive First-Run Wizard (`ScreenInit`):**
  - Master password input field with masked character entry (`bubbles/textinput`).
  - Password confirmation field enforcing exact character matching before allowing submission.
  - Real-time password entropy calculation and strength badge indicator (Weak -> Strong).
  - Explanatory security notice emphasizing offline zero-knowledge architecture (no "forgot password" reset mechanism).
  - Submit action (`Enter` / `[Initialize Vault]`) that invokes `VaultService.Init`, derives master key via Argon2id, executes SQLite migrations, unlocks session, and transitions directly to `ScreenDashboard`.
  - Cancel / quit action (`q` / `Ctrl+C`).
- **Memory Security & Invariants:**
  - Immediate zeroization of initialization password buffers in memory.
  - Enforcement of minimum 8-character password constraint.

## Non-Goals

- Cloud account creation, telemetry opt-ins, or remote synchronization onboarding.
- Recovery phrase (BIP-39) export during initial setup (deferred to dedicated backup tools).

## Constraints

- 100% offline with zero network connectivity.
- Enforce memory zeroization across all password text inputs.
- Initialize database strictly following Schema V1 and Argon2id KDF parameters matching CLI `govault init`.

## Relevant Knowledge

- `ai/memory/constitution.md`: Security Invariants, Offline Invariant, Memory Zeroization.
- `ai/raw/architecture.md`: Argon2id KDF, Authenticated Cipher Envelope, SQLite Storage.

## Open Questions

- None.
