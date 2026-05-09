# Final Verify — Phase 3.8B Dynamic Asset Discovery

## Summary

Implemented a read-only dynamic asset discovery API in `internal/capabilities/assets_discovery.go`:

- `DiscoverAssets(root string) (AssetCatalog, error)`

The API discovers assets without hardcoding only SDD/SDR/support and reports structured metadata for:

- Flows (`assets/flows/*.yaml`)
- Skill packs (`assets/skills/<pack-id>/` and `_shared`)
- Skill files (`assets/skills/<pack-id>/**/SKILL.md`)
- Shared contracts (`assets/skills/_shared/*.md`)
- Orchestrator contracts (`assets/orchestrators/*.md`, optional)

## Discovery contract

### Flows
Exposes manifest fields and source path including:
`id`, `name`, `kind`, `selectable`, `top_level_flow`, `family`, `skill_pack`, `orchestrator_skill`, `status`, `triggers`, `contracts`, `requires`, `persistence`, `stages`, `stage_graph`, `handoffs`, `final_artifacts`, `outputs`.

### Skill Packs
Exposes:
`id`, `source path`, `skill count`, `skill files`, `shared flag`, and related flow IDs (inferred from flow manifests by `skill_pack`).

### Skills
Exposes:
`id`, `pack id`, `relative path`, `source path`, parsed title/frontmatter fields, required heading presence/missing lists, and status (`valid`/`invalid`).

### Shared contracts
Exposes file ID, path, and non-empty status.

### Orchestrators
If directory exists, exposes file ID/path/title/non-empty status.
If missing, returns empty list (no failure).

## Tests added

`internal/capabilities/assets_discovery_test.go`

- Discovers all current flow manifests from repository root
- Discovers skill packs (`sdd`, `sdr`, `support`) and `_shared`
- Discovers skills recursively and validates required headings
- Handles missing `assets/orchestrators/` gracefully
- Verifies dynamic behavior with temporary fake flow + fake skill pack fixture

## Scope constraints preserved

- No runtime orchestration implemented
- No autonomous execution implemented
- No OpenCode apply/internal behavior changes
- No external integration requirements introduced
- Existing CLI/TUI behavior preserved

## Validation

- `go test ./...` ✅

## Next phase

Phase 3.8C — selection-aware OpenCode export projection based on discovered assets (still non-runtime and safety-preserving).
