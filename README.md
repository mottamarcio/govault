# GoVault 🛡️

**GoVault** is a modern, local-first, offline password and secrets manager written in Go. Built with uncompromising security standards, GoVault provides both a powerful command-line interface (CLI) for automation and a rich interactive Terminal User Interface (TUI) built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

---

## ✨ Key Features

- **🔒 Zero-Knowledge & 100% Offline:**
  - Strict zero-network guarantee: no telemetry, no cloud sync, no tracking, no external ports opened.
  - State-of-the-art cryptography: **Argon2id** Key Derivation Function (KDF) and authenticated **XChaCha20-Poly1305** / **AES-256-GCM** encryption.
  - Per-record subkey derivation using **HKDF-SHA256** and Authenticated Additional Data (AAD) binding vault IDs, record IDs, and versions.
  - Ephemeral in-memory key management with explicit zeroization (`memguard`-style memory wiping) upon lock, timeout, or exit.

- **🖥️ Modern Terminal User Interface (TUI):**
  - Aesthetic Nord/Catppuccin-inspired dark theme with adaptive split-pane layouts.
  - **Sidebar Navigation:** Category filtering (All Items, Logins, Notes, API Keys, Custom, Trash) with dynamic tag counters.
  - **Incremental Search:** Real-time fuzzy filtering and relevance ranking across secret titles, usernames, URIs, and tags.
  - **Interactive Form Editor:** Multi-field secret creation and editing with `Tab` navigation, validation, and dynamic custom key-value rows (`Ctrl+N` / `Ctrl+D`).
  - **Embedded Generator Overlay:** Live password & Diceware passphrase generator with Shannon entropy calculation (bits) and colorized strength meter.
  - **Modal Dialogs:** Confirmation overlays for destructive operations (move to trash, permanent purge, discard changes) and keybinding help (`?`).

- **⚙️ Comprehensive CLI Suite:**
  - Full CRUD operations with human-friendly formatting and `--json` mode for automated scripting.
  - Secure clipboard driver with timed background auto-clearing (45 seconds default).
  - Built-in `doctor` command to inspect storage health, file permissions, and cryptographic invariants.

- **📦 Portable Encrypted Backups (`.gvault`):**
  - Single-file encrypted container format with authenticated headers, Gzip compression, and integrity checksums.
  - Flexible restore and merge strategies: `overwrite`, `keep-newer`, and `duplicate`.

---

## 🚀 Installation & Build

### Prerequisites
- Go 1.22+ installed.

### Build from Source

```bash
# Clone repository
git clone https://github.com/mottamarcio/govault.git
cd govault

# Compile binary to current directory
go build -o govault ./cmd/govault

# Or install directly to $GOPATH/bin
go install ./cmd/govault
```

---

## 🏃 Quick Start Guide

### 1. Launch the Interactive TUI

```bash
./govault tui
```
*(If the vault does not exist yet, you can initialize it directly or specify `--vault /path/to/vault.db`)*

### 2. Command-Line Usage (CLI)

```bash
# 1. Initialize a new secure vault
./govault init --vault ~/.govault/vault.db

# 2. Generate a strong password or Diceware passphrase
./govault generate --length 24 --symbols --digits
./govault generate --passphrase --words 5

# 3. Add secrets
./govault add --title "GitHub" --type login --username "octocat" --password "S3cr3tP@ss!" --tags "dev,work"
./govault add --title "Stripe API Key" --type apikey --service "Stripe" --secret "sk_live_123456" --tags "prod,payments"
./govault add --title "Server Recovery Key" --type note --notes "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5..."

# 4. Search and list secrets
./govault list
./govault search "github"

# 5. Retrieve secret details (masked by default, use --reveal for plaintext)
./govault get "GitHub" --reveal

# 6. Copy password to clipboard with auto-clear timer (45 seconds)
./govault copy "GitHub" --field password

# 7. Edit an existing secret
./govault edit "GitHub" --username "octocat-v2"

# 8. Soft-delete and manage Trash
./govault delete "GitHub"
./govault trash list
./govault trash restore "GitHub"
./govault trash purge

# 9. Create an encrypted portable backup
./govault backup --out ./my-backup.gvault

# 10. Run offline integrity and security diagnostics
./govault doctor
```

---

## ⌨️ TUI Keyboard Shortcuts

| Shortcut | Description |
| :--- | :--- |
| `Tab` / `Shift+Tab` | Switch focus between sidebar categories, records table, and form fields |
| `j` / `k`, `↑` / `↓` | Navigate items in sidebar, tables, and history |
| `h` / `l` | Jump directly between sidebar and items table |
| `/` | Start live incremental fuzzy search |
| `a` | Add a new secret record (opens Form Editor) |
| `e` | Edit currently selected record |
| `d` | Move record to Trash (prompts for confirmation) / Purge if in Trash |
| `r` | Restore record from Trash |
| `v` | Toggle secret masking (reveal / hide plaintext) |
| `c` | Copy password to clipboard (auto-clears in 45 seconds) |
| `u` | Copy username to clipboard |
| `H` | Open revision history timeline viewer |
| `Ctrl+G` | Open embedded password & passphrase generator |
| `Ctrl+S` | Save record in Form Editor |
| `Ctrl+N` / `Ctrl+D` | Add / remove custom field row in Form Editor |
| `Ctrl+L` | Instantly lock vault and zeroize in-memory keys |
| `?` | Toggle interactive keyboard shortcuts cheatsheet |
| `Esc` | Dismiss active overlay / Cancel editor / Exit search |
| `q` / `Ctrl+C` | Lock vault and quit application |

---

## 🏗️ Architecture & Security Model

```
+---------------------------------------------------------------------------------+
|                                 User Interface                                  |
|            CLI (Cobra)              |        TUI (Bubble Tea + Lipgloss)        |
+---------------------------------------------------------------------------------+
                                      |
+---------------------------------------------------------------------------------+
|                              Application Services                               |
|   VaultService  |  RecordService  |  BackupEngine  |  Clipboard  |  AutoLock    |
+---------------------------------------------------------------------------------+
                                      |
+---------------------------------------------------------------------------------+
|                               Cryptography Engine                               |
|   Argon2id KDF  |  XChaCha20-Poly1305 / AES-256-GCM  |  HKDF-SHA256 Subkeys    |
|   Shannon Entropy Analyzer        |        Memguard-Style Zeroization           |
+---------------------------------------------------------------------------------+
                                      |
+---------------------------------------------------------------------------------+
|                               Storage & Persistence                             |
|       SQLite (WAL Mode, Foreign Keys, Schema Migrations, Revision History)      |
+---------------------------------------------------------------------------------+
```

- **Encryption Envelope:** Master password derives `K_vault` via Argon2id. Every record derives an isolated subkey `K_record = HKDF-SHA256(K_vault, record_id)`.
- **Authenticated Additional Data (AAD):** Prevents ciphertext substitution across vaults, records, or historical versions.
- **Revision History:** Editing a secret archives the previous version in a cryptographically isolated history log for auditability and rollback.
- **Memory Safety:** Master keys and decrypted secrets are held in memory only for the minimum required duration and wiped with `Zeroize()` upon timeout or lock.

---

## 🧪 Running Tests

GoVault includes an extensive automated test suite covering cryptographic invariants, database persistence, concurrent operations, and full TUI workflows.

```bash
# Run all unit and integration tests with Go race detector
go test -v -race ./...

# Run TUI tests specifically
go test -v -race ./internal/tui/...
```

---

## 📄 License

This project is licensed under the **MIT License**.
