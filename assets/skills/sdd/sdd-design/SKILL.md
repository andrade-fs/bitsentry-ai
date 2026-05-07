---
name: sdd-design
description: >
  Defines the technical architecture and data flow for the change. 
  Focuses on 'How' the system will be built to satisfy the spec.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: design
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
  
---

# Skill: sdd-design

## Purpose
Create the technical blueprint for the change. This skill maps the specification to specific code structures, data models, and integration points, ensuring the solution is idiomatic and robust.

## Use When
- `sdd-spec` is complete (or being run in parallel).
- You need to decide on specific technical implementations before writing code.

## Inputs
- **Specification**: `sdd/{slug}/spec.md`.
- **Exploration Findings**: `sdd/{slug}/explore.md`.
- **Session State**: `sdd/{slug}/state.yaml`.

## Workflow: The "Architectural Blueprint Protocol"
1.  **Component Modeling**: Identify new/modified classes, interfaces, or modules.
2.  **Data Flow & Schema**: Define DB changes, API contracts (JSON/Protobuf), and state transitions.
3.  **Integration Strategy**: Detail how the new logic hooks into existing entry points identified in `explore`.
4.  **Security & Error Design**: Define specific error types, validation logic, and permission checks.
5.  **Testing Strategy**: Outline how the design supports unit, integration, and property-based testing.

## Outputs
### Technical Design (Markdown)
Must include:
- **Architecture Diagram/Description**: High-level structural view.
- **Data Contracts**: API signatures and schema changes.
- **Internal Logic**: Pseudo-code or logic flow for complex algorithms.
- **Testing Plan**: Technical approach for verification.

## Boundaries
- **NO Implementation**: Do not generate the actual source code files.
- **NO Spec Changes**: If a spec is impossible to implement, trigger a `blocked` status; do not silently change the spec.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/design.md` | The technical blueprint.
- **target: engram** | `sdd/{slug}/architecture` | Structural reference for the apply phase.

## Result Envelope
**Status**: `success | partial | blocked`

**Next Recommended**: `sdd-tasks`