# Final Verify — Phase 3.8C Selection-aware OpenCode Export Projection

## Summary

Implemented a read-only projection API that maps selected capability/flow/pack IDs to discovered assets and computes what would be exported under the OpenCode managed area.

Implemented in:
- `internal/capabilities/opencode_export_projection.go`

Key APIs:
- `BuildOpenCodeExportProjection(catalog AssetCatalog, selectedIDs []string)`
- `GenerateSkillRegistry(projection OpenCodeExportProjection)`

## Inclusion behavior

Given `DiscoverAssets(...)` output + selected IDs, projection includes:
- `_shared` contracts when any flow/pack is selected
- selected flow manifests
- selected skill packs
- discovered stage skills referenced by selected flows
- support skills from `requires.support_skills`
- support pack when selected flow dependencies require it (safe MVP inclusion)
- optional orchestrator contract docs if present
- generated file preview: `bitsentry/skill-registry.md`

Alias compatibility:
- `bitsentry-sdd -> sdd`
- `bitsentry-sdr -> sdr`
- `bitsentry-support -> support`

Future dynamic flows/packs are supported when selected ID matches discovered flow or pack.

## Registry generation behavior

Generated registry preview includes deterministic sections:
- Included Flows
- Included Skill Packs
- Included Skills
- Shared Contracts
- Handoffs
- Persistence Roots
- Loading Rules

It explicitly mentions `Result Envelope` in loading rules.

## Tests added

`internal/capabilities/opencode_export_projection_test.go` validates:
- `bitsentry-sdd` selection includes shared + sdd (+ support dependency behavior)
- `bitsentry-sdr` includes sdr and not sdd
- `bitsentry-support` includes support flow/pack
- dynamic fake flow/pack selection works without hardcoded family dependency
- unknown selection yields warnings/skipped (no panic)
- generated registry contains core deterministic sections

## Scope constraints preserved

- Read-only projection only
- No OpenCode config writes
- No apply behavior changes
- No runtime orchestration/autonomous execution
- No external service requirements

## Validation

- `go test ./...` ✅

## Next phase

Phase 3.9 — actual selection-aware OpenCode export/apply integration using this projection model with existing safety guardrails.
