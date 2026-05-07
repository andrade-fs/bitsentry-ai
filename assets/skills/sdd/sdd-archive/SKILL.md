---
name: sdd-archive
description: >
  Closes the session by consolidating artifacts and documenting final decisions.
  Focuses on 'Cleanup' and 'Knowledge Transfer'.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: archive
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-archive

## Purpose
Formalize the end of an SDD session. This skill ensures that all knowledge gained during the change is persisted, follow-up tasks are tracked, and the project state is left clean for the next orchestrator run.

## Use When
- `sdd-verify` has issued a verdict (Gate 3).
- The change is either completed, abandoned, or deferred.

## Inputs
- **Final Verdict**: From `sdd-verify`.
- **Session State**: Full `sdd/{slug}/state.yaml`.
- **All Phase Artifacts**: From `init` to `verify`.

## Workflow: The "Session Consolidation Protocol"
1.  **Final Summary**: Synthesize the entire journey from raw request to verified code.
2.  **Artifact Indexing**: Create a final list of all `sdd/` files produced.
3.  **Unresolved Follow-ups**: Extract "deferred" tasks or technical debt into a handoff-ready format.
4.  **Decision Log**: Record key architectural decisions made (and why).
5.  **State Closure**: Mark the session as `closed` or `archived`.

## Outputs
### Archive Manifesto (Markdown)
Must include:
- **Executive Summary**: 2-3 paragraphs for future reference.
- **Decision Registry**: Key pivots or choices.
- **Next Actions**: Recommended follow-up issues or research.

## Boundaries
- **NO Work**: Do not perform any more development or testing.
- **Read-Only Finalization**: Just organize and summarize.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/archive.md` | The closing manifesto.
- **target: local** | `sdd/{slug}/state.yaml` | Final status update to `closed`.
- **target: engram** | `sdd/{slug}/final_state` | Persistent memory of the change for future context.

## Result Envelope
**Status**: `success`
**Next Recommended**: `none`

**Handoffs**:
- **To Issue Creation**: For any follow-up debt.
- **To Branch PR**: To finalize the merge request text.