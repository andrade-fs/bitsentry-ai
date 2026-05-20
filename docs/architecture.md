# Architecture Overview (Public MVP)

This document describes the **current public MVP architecture**, not the original bootstrap phase.

## Architectural posture

- **OpenCode/TUI-first** product posture.
- **CLI as support/debug surface**, not primary onboarding UX.
- **Controlled manual MVP**: explicit user confirmation and visible readiness states.
- **No live web execution** in this posture.
- **No `.env` or secrets handling**.

## Runtime shape (high level)

1. User enters via TUI (`bitsentry-ai`) or CLI.
2. System performs environment + capability checks.
3. Install/Setup presents readiness and manual next steps.
4. OpenCode-facing guidance drives follow-up usage.

The key principle is explicit control over hidden automation.

## Package map

- `cmd/bitsentry-ai`
  - Binary entrypoint.
- `internal/tui`
  - Primary guided UX for Install/Setup and readiness signaling.
- `internal/cli`
  - Command surface for support/debug/status paths.
- `internal/app`
  - Composition root and service wiring.
- `internal/config`
  - Local configuration read/write (`~/.bitsentry-ai/config.yaml`).
- `internal/system`
  - OS/arch/shell detection and dependency checks.
- `internal/agents`
  - OpenCode/agent detection and integration hooks.
- `internal/profiles`
  - Profile catalog/selection for local context management.
- `internal/components`
  - Capability registry and metadata wiring.
- `internal/workflows`
  - Flow definitions and routing contracts.
- `internal/logs`
  - Logging initialization and local log management.

## Control boundaries (MVP)

- Readiness and preview signals are first-class outputs.
- Manual intent and operator confirmation are required for impactful steps.
- Security posture avoids autonomous or hidden runtime actions.
- Product messaging must remain honest about current limits.

## Relationship with other docs

- Install and operator flow: [docs/install.md](install.md)
- Public launch readiness: [docs/releases/public-mvp-checklist.md](releases/public-mvp-checklist.md)
- Demo scripts/prompts: [docs/demo/public-mvp-prompts.md](demo/public-mvp-prompts.md)
- Near-term direction: [docs/roadmap.md](roadmap.md)
