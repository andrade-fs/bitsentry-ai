---
name: sdd-init
description: >
  Bootstrap a deterministic SDD change context. Normalizes raw requests into 
  stable session metadata and identity.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: init
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-init

## Purpose
Initialize the "Source of Truth" for an SDD session. This skill transforms a vague or raw user intent into a structured, traceable, and unique initialization package that governs the rest of the flow.

## Use When
- Starting a new feature, bugfix, refactor, or architectural change.
- A request needs to be scoped and named before any exploration or code reading occurs.
- You need to establish the `change-slug` that will index all future artifacts.

## Inputs
- **Raw Request**: The user's prompt, issue description, or task.
- **Context**: Project name and existing constraints (if any).

## Workflow
1.  **Intent Extraction**: Identify the `Action` (e.g., Add), `Object` (e.g., Auth Provider), and `Context`.
2.  **Identity Generation**:
    - **Name**: Human-readable (e.g., `OAuth2 Provider Support`).
    - **Slug**: Deterministic kebab-case (e.g., `oauth2-provider-support`).
3.  **Typing**: Classify as `feature | bugfix | refactor | test | docs | architecture`.
4.  **Guardrail Definition**: Define one **Primary Goal** and at least three **Non-Goals**.
5.  **Scope Mapping**: Categorize items into `In-Scope`, `Out-of-Scope`, and `Unknown`.
6.  **Constraint Capture**: List technical or safety limitations.
7.  **State Initialization**: Set `status: initialized` and point to `sdd-explore`.

## Outputs
### Initial State Snapshot (YAML)
```yaml
change:
  slug: "normalized-kebab-case-id"
  name: "Human Readable Name"
  type: "feature | bugfix | ..."
  goal: "Primary objective"
  non_goals: ["list", "of", "exclusions"]
  scope:
    in: []
    out: []
    unknown: []
  status: "initialized"
phase:
  current: "init"
  next: "sdd-explore"
artifacts:
  local_root: "sdd/{slug}/"
  engram_root: "sdd/{slug}/"
```

## Boundaries
- **NO Deep Exploration**: Do not read codebase files yet.
- **NO Implementation**: Do not suggest how to fix the issue.
- **NO Spec/Design**: Do not define requirements or technical architecture.
- **Minimal Interaction**: Infer defaults; only block if the request is total gibberish.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md
- **target: local** | `sdd/{slug}/state.yaml` | content summary: Initialized state object with slug and primary goal..
- **target: engram** | `sdd/{slug}/init` | content summary: Bootstrap metadata..

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Initialized SDD session `{slug}`. Identity and scope established. Ready for `sdd-explore`.

**Detailed Report**:
- **Identity**: Name, Slug, Type.
- **Summary of Intent**: Refined goal and scope.
- **Risk Assessment**: Initial unknowns that could block the flow later.

**Artifacts**:
- `sdd/{slug}/state.yaml`
- `sdd/{slug}/init.md`

**Persistence Actions**:
- target: local
  key/path: `sdd/{slug}/state.yaml`
  content summary: Initialized state object with slug and goal.

**Next Recommended**: `sdd-explore`

**Handoffs**:
- **To SDR**: If the request is 80% research and 20% implementation.
- **To Blocked**: If no goal or slug can be responsibly derived.

**Risks / Gaps**:
- List of "Unknowns" that `sdd-explore` MUST address.

## Quality Checklist
- [ ] `change-slug` is strictly kebab-case and unique.
- [ ] At least 3 `non-goals` are defined to prevent creep.
- [ ] `sdd-explore` is the explicit next step.
- [ ] Persistence paths use the new `slug` consistently.
- [ ] No code or architectural design was attempted.
```

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
