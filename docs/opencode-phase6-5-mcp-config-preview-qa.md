# OpenCode Phase 6.5 MCP Config Preview QA

Status: **PASS**

Related closure doc:
- `docs/opencode-phase6-6-release-stabilization.md`

## Scope

Validate Phase 6.5 preview-only MCP config behavior in OpenCode path with strict safety constraints.

## Verified

- Added preview model fields:
  - `current_config_state`
  - `exists`
  - `readable`
  - `invalid_error`
  - `current_mcp_config_detected`
  - `mcp_readiness_state`
  - `proposed_safe_changes`
  - `preserved_keys`
  - `preserved_mcp_entries`
  - `warnings`
  - `manual_steps`
  - `would_write`
  - `requires_confirmation`
  - `backup_required`

- Preview builder is pure/read-only:
  - no writes
  - no backup creation
  - no mutation of `opencode.json`
  - no `.env` reads
  - no credential/token insertion

- Scenario handling:
  - missing `opencode.json`
  - invalid `opencode.json`
  - readable `opencode.json` without MCP
  - readable `opencode.json` with existing MCP entries

- Guardrail behavior:
  - existing MCP entries are preserved in preview (`preserved_mcp_entries`)
  - future apply remains confirmation-gated and backup-required (`requires_confirmation=true`, `backup_required=true`)
  - no change to `agent.bitsentry.permission.edit=deny` contract

- TUI integration:
  - Install review step displays MCP preview section marked **PREVIEW ONLY**
  - Done/control-panel step displays preview contract fields and manual steps

## Notes

- This phase intentionally does not implement MCP credential writes or automatic apply.
- Any sensitive setup remains explicit manual steps until a future gated apply phase.
