# Final Verify — Phase 3.4 TUI Capability Selector MVP

## Verdict
PASS WITH NOTES

## What was verified

- TUI Capabilities screen supports draft selection for preset/MCPs/flows/skills.
- Validate (`v`) runs against the current draft using capability validation logic.
- Plan preview (`l`) builds plan/projection from current draft.
- Save (`s`) persists capability config explicitly.
- Dry-run apply (`d`) and real apply (`a` + `y`) use existing safe CLI apply path.

## Critical fix applied during final verify

- Apply flow previously persisted draft implicitly before apply.
- Now apply/dry-run are blocked if draft has unsaved changes (`capDirty=true`).
- User must save draft (`s`) before apply/dry-run.

## Safety behavior

- Real apply requires explicit two-step confirmation (`a` then `y`).
- Real apply is blocked when validation fails.
- Screen messaging clarifies:
  - OpenCode is the only real target currently.
  - Skills/flows remain declarative in this phase.
  - Postgres MCP is modeled but skipped in real apply.

## Notes

- TUI apply invokes current executable (`os.Executable`) to run CLI apply command.
- This is acceptable for MVP; future phase may extract a shared capabilities service for tighter integration.
