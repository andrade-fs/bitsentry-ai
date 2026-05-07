---
name: sdd-tasks
description: >
  Converts technical design into a sequenced, atomic task list.
  Focuses on 'When' and 'In what order' the work happens.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: tasks
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
  
---

# Skill: sdd-tasks

## Purpose
Deconstruct the approved `sdd-design` and `sdd-spec` into a granular, ordered execution plan. This skill ensures that implementation is systematic, reduces merge risks, and identifies logical dependencies.

## Use When
- `sdd-design` and `sdd-spec` are finalized.
- You need a roadmap for the `sdd-apply` phase.

## Inputs
- **Technical Design**: `sdd/{slug}/design.md`.
- **Requirements**: `sdd/{slug}/spec.md`.
- **Exploration Report**: `sdd/{slug}/explore.md`.

## Workflow: The "Tactical Planning Protocol"
1.  **Atomic Breakdown**: Identify the smallest possible units of work (e.g., "Create interface X", "Add field Y to DB").
2.  **Dependency Mapping**: Determine which tasks must happen first (e.g., "Data layer before API layer").
3.  **Target Assignment**: Explicitly map tasks to specific files or modules identified in `explore`.
4.  **Verification Steps**: For each task, define how to know it is "done" (e.g., "Compile check", "Unit test pass").
5.  **Rollback Points**: Identify "Safe Points" where the implementation could be paused or reverted.

## Outputs
### Implementation Plan (Markdown)
Must include:
- **Phase-based Task List**: Groups of tasks (Setup, Core, Integration, Cleanup).
- **File/Module Targets**: Exact paths to be modified.
- **Dependency Graph**: Clear order of operations.

## Boundaries
- **NO Implementation**: Do not write the code.
- **NO Architectural Changes**: Follow the `design.md` strictly. If the design is flawed, block the skill.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/tasks.md` | The execution roadmap.
- **target: engram** | `sdd/{slug}/task_state` | Tracking for the implementation progress.

## Result Envelope
**Status**: `success | partial | blocked`
**Next Recommended**: `sdd-apply` (Needs Gate 2 Approval)