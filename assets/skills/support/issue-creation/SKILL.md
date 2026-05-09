---
name: support-issue-creation
description: >
  Drafts structured, high-fidelity issues for tracking work. 
  Focuses on clarity, testable criteria, and scope boundaries.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
---

# Skill: support-issue-creation

## Purpose
Standardize the creation of work units. This skill ensures that every issue has enough context for an agent (or human) to execute it without back-and-forth, defining exactly what "done" looks like.

## Use When
- A technical debt item is identified during `sdd-archive`.
- A research finding in `sdr-validate` requires a follow-up implementation.
- You need to break a large goal into smaller, manageable chunks.

## Workflow
1.  **Context Synthesis**: Distill the problem statement into a concise "Why".
2.  **Identity Creation**: 
    - **Title**: Action-oriented (e.g., `Implement ZFS Scrub Automation`).
    - **Slug**: Follows the kebab-case standard for file persistence.
3.  **Boundary Definition**:
    - **In-Scope**: Explicit list of deliverables.
    - **Out-of-Scope**: Explicit "Non-Goals" to prevent scope creep.
4.  **Acceptance Criteria (AC)**: Define testable conditions (e.g., "Script returns 0 if pool is healthy").
5.  **Classification**: Suggest labels (`bug`, `feature`, `refactor`, `security`) and priority (`P0` to `P3`).
6.  **Technical Notes**: Include any known risks or specific files to watch (from `sdd-explore`).

## Outputs
### Issue Template (Markdown)
Must be a clean, copy-paste ready block including:
- **Title**: `[TYPE] Title`
- **Body**: Problem, Scope, and ACs in a structured format.
- **Metadata**: Suggested labels and priority.

## Boundaries
- **NO API Integration**: Do not attempt to connect to GitHub, Jira, or Trello.
- **NO Implementation**: Do not write the code to solve the issue.
- **No Vague ACs**: Every criterion must be something a human or test can verify as "True/False".

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: `planning/issues/{slug}.md`
  **action**: upsert
  **summary**: Structured issue draft for GitHub/GitLab.
- **target**: engram
  **key**: `support/issues/{slug}`
  **action**: upsert
  **summary**: Backlog item metadata.

## Result Envelope
**Status**: `success | partial`

**Executive Summary**:
Issue draft for `{slug}` created. Ready for backlog ingestion.

**Persistence Actions**:
- target: local
  key/path: `planning/issues/{slug}.md`
  content summary: Complete issue body with acceptance criteria.

**Next Recommended**: `support/branch-pr` (if implementation starts immediately).

**Handoffs**:
- **to**: `human-operator`
  **reason**: Manual creation in the issue tracker.

**Quality Checklist**:
- [ ] Acceptance criteria are binary (Pass/Fail).
- [ ] Labels and priority are logically assigned.
- [ ] Non-goals are clearly defined to protect the implementation agent.

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.

## Quality Checklist
- [ ] Required heading present for contract compliance.
- [ ] Guidance remains declarative.
- [ ] No runtime behavior changes introduced.
