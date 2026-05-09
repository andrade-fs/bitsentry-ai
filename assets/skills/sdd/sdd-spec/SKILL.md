---
name: sdd-spec
description: >
  Defines behavioral requirements and acceptance criteria. 
  Focuses on 'What' the system must do to satisfy the proposal.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: spec
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
  
---

# Skill: sdd-spec

## Purpose
Translate the approved proposal into a set of rigorous, testable requirements. This skill defines the expected behavior, edge cases, and success metrics that the implementation must satisfy.

## Use When
- `sdd-propose` has been approved (Gate 1).
- You need to establish the "Contract of Truth" for what constitutes a successful change.

## Inputs
- **Approved Proposal**: `sdd/{slug}/proposal.md`.
- **Session State**: `sdd/{slug}/state.yaml`.
- **Repo Context**: To ensure specs align with existing business logic.

## Workflow
1.  **User Stories / Functional Requirements**: Break down the proposal into discrete units of behavior.
2.  **Acceptance Criteria (AC)**: Define exact conditions that must be met (e.g., "API must return 404 if ID is missing").
3.  **Edge Case Mapping**: Identify "sad paths", boundary conditions, and error states.
4.  **Non-Functional Requirements**: Define performance targets, security constraints, and observability needs.
5.  **Validation Matrix**: Create a checklist that connects requirements to future verification.

## Outputs
### Specification Document (Markdown)
Must include:
- **Functional Requirements**: Detailed behavior list.
- **Acceptance Criteria**: Testable "Given/When/Then" or bulleted conditions.
- **Constraints Checklist**: Security, performance, and compatibility specs.

## Boundaries
- **NO Technical Design**: Do not specify classes, DB tables, or code structures.
- **NO Implementation**: Do not write code.
- **NO Tasking**: Do not list implementation steps.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/spec.md` | The definitive specification.
- **target: engram** | `sdd/{slug}/requirements` | Indexed ACs for the verify phase.

## Result Envelope
**Status**: `success | partial | blocked`

**Next Recommended**: `sdd-design`

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.

## Quality Checklist
- [ ] Required heading present for contract compliance.
- [ ] Guidance remains declarative.
- [ ] No runtime behavior changes introduced.
