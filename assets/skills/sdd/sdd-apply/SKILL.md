---
name: sdd-apply
description: >
  Guides the actual implementation of code changes based on the task plan.
  Focuses on 'Doing' the work within the defined scope.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: apply
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-apply

## Purpose
Direct the implementation process. This skill acts as the bridge between the plan (`tasks.md`) and the actual code edits, ensuring that every change follows the approved design and satisfies the specification.

## Use When
- `sdd-tasks` is approved and Gate 2 is cleared.
- You are ready to generate or guide the modification of source files.

## Inputs
- **Task Plan**: `sdd/{slug}/tasks.md`.
- **Design & Spec**: The "How" and "What" of the change.
- **Code Context**: Direct access to the target files.

## Workflow: The "Implementation Guardrail Protocol"
1.  **Task Selection**: Pick the next available task from the `tasks.md` list.
2.  **Code Generation/Modification**:
    - Apply changes following the architectural patterns from `explore`.
    - Ensure all `spec.md` acceptance criteria are addressed.
    - Avoid "Scope Leak" (don't fix unrelated bugs).
3.  **Self-Correction**: Check that new code doesn't violate existing patterns (e.g., naming, error handling).
4.  **Status Tracking**: Mark tasks as `complete`, `in-progress`, or `failed`.
5.  **Drafting Reports**: Record exactly what was changed for the verify phase.

## Outputs
### Implementation Summary (Markdown)
Must include:
- **Changes Applied**: List of modified files and summary of edits.
- **Unresolved Issues**: Tasks that couldn't be completed or new debt created.
- **Verification Readiness**: Signal that code is ready for `sdd-verify`.

## Boundaries
- **NO Unplanned Changes**: Do not modify files not listed in `tasks.md`.
- **NO Design Drifting**: If a design decision must be changed, you must go back to `sdd-design`.
- **NO Final Verification**: `apply` only checks if code *exists*; `verify` checks if it *works*.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/apply.md` | Log of implementation actions.
- **target: opencode** | `artifacts/sdd/{slug}/changes` | References to the new/modified code.

## Result Envelope
**Status**: `success | partial | blocked`
**Next Recommended**: `sdd-verify`