# Final Verify — Phase 3.9 OpenCode Skills Export

## Summary

Implemented selection-aware OpenCode skills export using the existing capabilities workflow:

- `bitsentry-ai capabilities export-preview --target-agent opencode [--select ...]`
- `bitsentry-ai capabilities export --target-agent opencode [--dry-run] [--select ...]`

Pipeline:
1. Discover assets dynamically (`DiscoverAssets`)
2. Build selection-aware projection (`BuildOpenCodeExportProjection`)
3. Execute safe export (`ExecuteOpenCodeSkillsExport`)
4. Persist export report (`WriteOpenCodeSkillsExportReport`)

## Managed target area

Exports write only under:

`<opencode-config-root>/bitsentry/`

Including:
- `skills/_shared/**`
- `skills/<pack-id>/**`
- `flows/<flow-id>.yaml`
- `orchestrators/*.md` (if present)
- `skill-registry.md` (generated)

## Safety behavior

- Dry-run never writes to OpenCode.
- Real export creates backup under `~/.bitsentry-ai/backups/opencode-skills/<timestamp>/`.
- Path traversal prevented by managed-root and relative-path validation.
- Unknown selections become warnings/skipped in projection/report.
- Invalid projected skill blocks real export.

## Reporting

Reports stored in:

`~/.bitsentry-ai/exports/opencode-skills/`

With timestamped file + `latest.yaml`, including:
- dry_run
- status (`preview | exported | partial | failed`)
- selected IDs
- target root
- included flows/packs/skills count
- generated/written files
- warnings/skipped
- backup path

## Scope constraints preserved

- No runtime orchestration execution.
- No OpenCode config mutation outside managed bitsentry area.
- Existing MCP apply behavior remains intact.
- No external service dependencies.

## Validation

- `go test ./...` ✅

## Next phase

Phase 4.0 — Main Bitsentry Orchestrator contract and minimal runtime orchestration MVP with strict guardrails.
