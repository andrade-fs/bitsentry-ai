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

If request is clearly trivial/small, allow direct reasoning path instead of forcing full SDD.

## Inputs
- **Raw Request**: The user's prompt, issue description, or task.
- **Context**: Project name and existing constraints (if any).

## Workflow
0.  **Handshake First (Mandatory)**:
    - If no active flow exists, first classify route: `SDD | SDR | Support | Direct reasoning` and explain briefly why.
    - If route is ambiguous, ask user to choose before proceeding.
    - Ask for **Execution Mode**: `interactive (default) | autonomous-plan | autonomous-apply | direct reasoning`.
    - Ask for **Persistence Mode**: `engram (default if available) | openspec | both | none`.
    - If Engram available, query relevant memory context before meaningful work.
    - If Engram unavailable, state clearly and offer OpenSpec or Engram-ready blocks.
    - Confirm **Mutation/Write Permissions** and **Phase Progression Policy**.
    - Do **NOT** auto-advance to `sdd-explore` until explicitly confirmed.
1.  **Intent Extraction**: Identify the `Action` (e.g., Add), `Object` (e.g., Auth Provider), and `Context`.
2.  **Identity Generation**:
    - **Name**: Human-readable (e.g., `OAuth2 Provider Support`).
    - **Slug**: Deterministic kebab-case (e.g., `oauth2-provider-support`).
3.  **Typing**: Classify as `feature | bugfix | refactor | test | docs | architecture`.
4.  **Guardrail Definition**: Define one **Primary Goal** and at least three **Non-Goals**.
5.  **Scope Mapping**: Categorize items into `In-Scope`, `Out-of-Scope`, and `Unknown`.
6.  **Constraint Capture**: List technical or safety limitations.
7.  **State Initialization**: Set `status: initialized` and point to `sdd-explore` (pending user confirmation in interactive mode).

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
- **NO Persistence Side Effects by Default**: `none` mode means no files, no folders, no memory writes.
- **NO Hidden Reasoning Dumps**: never expose raw `Thinking:` blocks.
- **Compact by Default**: interactive summaries should be short and decision-oriented.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md
- **default (`none`)**: no persistence actions.
- **target: local-openspec** | `openspec/{slug}/state.yaml` | only after explicit user confirmation.
- **target: engram-ready** | `sdd/{slug}/init` draft block only; save only if real Engram write is available and user confirmed.

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Initialized SDD session `{slug}`. Identity and scope established. Ready for `sdd-explore`.

Main chat output should be compact envelope only:
- Phase
- Verdict
- Useful findings (3-5 max)
- Blocking questions (if needed)
- Decision/recommendation
- Next step

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
