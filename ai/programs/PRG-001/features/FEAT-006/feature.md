---
id: FEAT-006
type: feature
status: active
parent: PRG-001
---

# FEAT-006: Interactive Terminal User Interface (TUI)

## Capability

Provides an interactive, keyboard-driven Terminal User Interface built with Bubble Tea, Bubbles, and Lip Gloss for browsing, searching, editing, and managing secrets visually.

## User Value

Provides an intuitive, terminal-native visual workspace for managing credentials, inspecting secret histories, copying values, and generating passwords without memorizing CLI syntax.

## Scope

- Bubble Tea Model-View-Update (Elm architecture) navigation:
  - Unlock Screen: Master password prompt with masked input and error feedback.
  - Main Dashboard: Sidebar for categories/tags and main table/list of secrets.
  - Detail View: Formatted inspection of usernames, masked passwords, URLs, custom fields, and revision history.
  - Record Editor: Interactive forms for creating/editing secrets, adding custom key-value pairs, and managing tags.
  - Generator Overlay: Interactive password & passphrase generator with live entropy preview.
  - Status Bar: Auto-lock countdown timer, active vault indicator, and contextual keybinding hints.
- Clipboard shortcuts: Hotkeys (e.g. `c` for password, `u` for username) triggering background timed auto-clearing.
- Inactivity auto-lock timer and secure screen clearing on lock/exit.

## Non-Goals

- Web or GUI desktop interfaces.
- Direct database or crypto implementations inside Bubble Tea models.

## Constraints

- Pure presentation: TUI models must interact with the application solely through Application Services.
- Sensitive values must remain masked on screen unless the user explicitly toggles visibility (e.g., `v` key).
- Graceful terminal restoration (alternate screen buffer exit) on exit or unexpected signals.

## Relevant Knowledge

- [KNOW-001](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-001-product-requirements.md) — Product Requirements and CLI/TUI Capabilities
- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model and Trust Boundaries
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — GoVault Software Architecture and Package Design

## Open Questions

- None.
