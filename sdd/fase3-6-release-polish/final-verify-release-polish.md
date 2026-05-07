# Final Verify — Phase 3.6A Capability Preview Core + Release Polish

## Final installer status

- Capability preview/validate/plan/report summary logic is now shared under `internal/capabilities`.
- CLI capability apply uses shared preview summary and report writer/reader.
- TUI continues using shared capability service for selection/validate/plan/apply orchestration.

## Commands verified

- `capabilities status`
- `capabilities inspect <id>`
- `capabilities configure`
- `capabilities validate`
- `capabilities plan`
- `capabilities apply --dry-run`
- `capabilities apply`
- `capabilities report latest`

## TUI flow verified

- Draft selection/editing remains in-memory until save
- Validate and plan preview remain read-only
- Dry-run/apply behavior preserved

## Limitations

- Real apply remains on the existing safe OpenCode path (no internal apply-core rewrite in this phase)
- Postgres MCP remains modeled/skipped
- Skills/flows remain declarative in apply semantics

## Next recommended phase

Phase 3.7 — Core Skills Pack
