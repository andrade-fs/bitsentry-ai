# OpenCode Convention (Contract)

## Scope

No OpenCode internal modifications in this phase.

## Artifact references

- Skills/flow outputs should be referenceable by path and short summary.
- Prefer deterministic paths:
  - `~/.bitsentry-ai/artifacts/sdd/{change-name}/{phase}.md`
  - `~/.bitsentry-ai/artifacts/sdr/{topic-id}/{phase}.md`

## Apply safety compatibility

- Keep existing safe apply path unchanged
- Preserve backup/merge semantics
- Preserve managed-vs-skipped MCP behavior

## Report compatibility

- Capability reports remain at: `~/.bitsentry-ai/exports/capabilities/apply/`
- `latest.yaml` remains canonical pointer
