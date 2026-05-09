# Final Verify — Phase 3.8A Flow Manifest Normalization

## Summary

Normalized `assets/flows/sdd.yaml`, `assets/flows/sdr.yaml`, and `assets/flows/support.yaml` for dynamic declarative discovery.

Key outcomes:
- common dynamic fields present (`kind`, `selectable`, `top_level_flow`, `skill_pack`, `orchestrator_skill`, `triggers`, `contracts`, `persistence`, `stage_graph`)
- pack-prefixed skill refs enforced (`sdd/...`, `sdr/...`, `support/...`)
- SDD explicitly models OpenCode under `requires.targets`, not `requires.mcps`

## Scope constraints preserved

- No runtime orchestration implemented
- No OpenCode apply internal changes
- No new MCP runtime behavior
- No CLI/TUI behavior changes

## Dynamic discovery readiness

Manifests now support future:
- dynamic flow listing/selection
- routing heuristics
- orchestrator phase mapping

## Skill contract restoration

After flow manifest normalization, `TestAllSkillFilesContainRequiredSections` exposed pre-existing skill heading inconsistencies.

Actions taken:
- restored strict heading contract in `internal/capabilities/skills_contract_test.go`
- normalized all affected `assets/skills/**/SKILL.md` files to include required headings exactly:
  - `## Purpose`
  - `## Use When`
  - `## Inputs`
  - `## Workflow`
  - `## Outputs`
  - `## Boundaries`
  - `## Persistence Actions`
  - `## Result Envelope`
  - `## Handoffs`
  - `## Quality Checklist`

Validation:
- `go test ./internal/capabilities -run FlowManifest` ✅
- `go test ./...` ✅

## Next phase

Phase 3.8B — Dynamic asset discovery API (read-only) for CLI/TUI/orchestrator consumption.
