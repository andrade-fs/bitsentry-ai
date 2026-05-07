---
name: sdd-propose
description: >
  Generates a high-level technical proposal for the change. 
  Acts as the primary decision gate before formalizing requirements.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: propose
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-propose

## Purpose
Synthesize the findings from `sdd-explore` into a concrete strategy. This skill defines *how* the goal will be achieved at a high level, allowing stakeholders to approve the direction before investing time in detailed specs or implementation.

## Use When
- `sdd-explore` has successfully mapped the impact area.
- There are multiple ways to solve the problem and a choice must be made.
- You need official approval (Gate 1) to proceed with the technical design.

## Inputs
### Required
- **Session State**: `sdd/{slug}/state.yaml`.
- **Exploration Report**: `sdd/{slug}/explore.md`.
- **Primary Goal**: From `sdd-init`.

### Optional
- Alternative approaches suggested by the user.
- Global architectural constraints.

## Workflow: The "Strategic Alignment Protocol"
1.  **Problem Statement**: Re-articulate the problem based on exploration (why the current state is insufficient).
2.  **Solution Synthesis**: 
    - Formulate the **Proposed Approach** (the "Happy Path").
    - Identify **Alternatives Considered** and why they were discarded.
3.  **Impact Assessment**: 
    - List modules to be modified vs. modules to be created.
    - Identify breaking changes or migration needs.
4.  **Risk & Mitigation**: Identify what could go wrong (performance hits, security gaps) and how to handle it.
5.  **Rollback Strategy**: Define how to undo the change if it fails in production.
6.  **Approval Question**: Formulate a clear "Gate 1" question for the user/orchestrator.

## Outputs
### Change Proposal (Markdown)
Must include:
- **The "Why"**: Refined problem definition.
- **The "How"**: High-level technical strategy.
- **Trade-offs**: What are we gaining and what are we sacrificing?
- **Security/Performance Impact**: Preliminary notes.
- **Rollback Idea**: High-level plan for reversal.

## Boundaries
- **NO Detailed Design**: Do not define specific function signatures or DB schemas (that's for `sdd-design`).
- **NO Implementation**: Do not write code blocks.
- **NO Tasking**: Do not create a step-by-step plan (that's for `sdd-tasks`).
- **Respect Explore**: Do not propose changes to files not identified during exploration without justification.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/proposal.md` | The core proposal document.
- **target: local** | `sdd/{slug}/state.yaml` | Update status to `awaiting_approval`.
- **target: engram** | `sdd/{slug}/proposal_summary` | Executive summary for session memory.

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Proposal for `{slug}` generated. Strategy: `{summary}`. Awaiting Gate 1 approval.

**Detailed Report**:
- **Proposed Strategy**: Core logic of the change.
- **Alternatives**: Why this is better than [X].
- **Approval Gate**: Clear prompt for the orchestrator/user.

**Artifacts**:
- `sdd/{slug}/proposal.md`

**Persistence Actions**:
- target: local
  key/path: `sdd/{slug}/proposal.md`
  content summary: High-level strategy, risks, and alternatives.

**Next Recommended**: `sdd-spec` (after approval)

**Handoffs**:
- **To SDR**: If the proposal requires validating a third-party tool or a non-trivial algorithm.
- **To Judgment Day**: If the proposal is high-risk and needs an adversarial critique before approval.

**Risks / Gaps**:
- List any unresolved trade-offs that require user input.

## Quality Checklist
- [ ] The proposal directly addresses the goal from `sdd-init`.
- [ ] Alternatives were considered and documented.
- [ ] Rollback strategy is present.
- [ ] No implementation-level details (code) are included.
- [ ] Impact on existing dependencies is clear.