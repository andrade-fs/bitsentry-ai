# Architecture Overview (Phase 1)

Phase 1 is a **Bootstrap MVP**: a modular foundation with working CLI/TUI, environment checks, profile management, and placeholder registries for future capabilities.

## Package map

- `cmd/bitsentry-ai`
  - Binary entrypoint.
- `internal/cli`
  - Cobra command tree (`version`, `doctor`, `agents`, `profiles`, `profile use`, `components`, `config path`).
- `internal/app`
  - Application wiring and shared runtime services.
- `internal/config`
  - Config path resolution and config persistence (`~/.bitsentry-ai/config.yaml`).
- `internal/system`
  - OS/arch/shell detection and dependency checks.
- `internal/agents`
  - Agent detector contracts and OpenCode detection.
- `internal/profiles`
  - Default profile catalog and profile selection/persistence support.
- `internal/components`
  - Component registry (stub entries for future phases).
- `internal/workflows`
  - Workflow registry (stub entries for future phases).
- `internal/tui`
  - Minimal Bubble Tea based interactive UI.
- `internal/logs`
  - Logger initialization and log path handling.

## Design intent

The architecture is intentionally modular so CLI and TUI consume the same domain services. This avoids duplicated behavior and keeps future workflow/component implementation isolated behind registries and interfaces.

## Scope note

In Phase 1, `components` and `workflows` are **stubs only**. They communicate roadmap intent and command shape, but do not execute real SDR/SDD/Red Team/Bug Bounty automation yet.
