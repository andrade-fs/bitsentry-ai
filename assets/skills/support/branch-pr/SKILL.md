---
name: support-branch-pr
description: >
  Generates a high-fidelity branch strategy and PR narrative. 
  Prepares all documentation for Git/GitHub operations without execution.
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

# Skill: support-branch-pr

## Purpose
Formalize the transition from "Implementation/Verification" to "Integration". This skill generates the technical naming, commit sequence, and Pull Request documentation needed to merge work into the main codebase.

## Use When
- `sdd-verify` is complete with a PASS verdict.
- You need a structured plan for a human or a CI agent to perform Git operations.
- You want to ensure PR descriptions are consistent and evidence-based.

## Inputs
### Required
- **Session State**: `sdd/{slug}/state.yaml`.
- **Implementation Summary**: `sdd/{slug}/apply.md`.
- **Verification Evidence**: `sdd/{slug}/verify.md`.

## Workflow
1.  **Branch Naming**: Generate a deterministic branch name (e.g., `feature/oauth2-provider` or `fix/mem-leak-storage`).
2.  **Commit Sequencing**: Break the `apply.md` changes into a logical commit plan (e.g., "Refactor interface", "Implement logic", "Add tests").
3.  **PR Narrative Construction**:
    - **Title**: Concise and descriptive.
    - **Description**: "What" changed and "Why".
    - **Visual Check**: Reference any UI/UX changes if applicable.
4.  **Evidence Embedding**: Consolidate the `verify.md` results into a "Validation Evidence" section for the PR body.
5.  **Risk & Rollback**: Define the risk of the merge and the exact command/steps to revert.

## Outputs
### Branch/PR Plan (Markdown)
Must include:
- **Target Branch**: `main | master | develop`.
- **New Branch Name**: `{type}/{slug}`.
- **Commit Plan**: List of messages and targeted files.
- **PR Body**: Full Markdown template ready to copy/paste.
- **Verification Linkage**: References to local `verify.md`.

## Boundaries
- **NO Git Operations**: Do not run `git checkout`, `git commit`, or `git push`.
- **NO API Calls**: Do not interact with GitHub/GitLab APIs.
- **NO Code Edits**: The code is already written; this skill only plans the delivery.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: `planning/pr/{slug}.md`
  **action**: upsert
  **summary**: Complete PR narrative and branch strategy.

## Result Envelope
**Status**: `success | partial`

**Executive Summary**:
Branch/PR plan generated for `{slug}`. Ready for human execution.

**Persistence Actions**:
- target: local
  key/path: `planning/pr/{slug}.md`
  content summary: PR description, branch name, and commit sequence.

**Next Recommended**: `none` (End of automated support flow)

**Handoffs**:
- **to**: `human-operator`
  **reason**: Ready for Git execution and manual review.

**Risks / Gaps**:
- Identify if the PR is too large (Mega-PR warning) and suggest splitting.

## Quality Checklist
- [ ] Branch name follows the `{type}/{slug}` convention.
- [ ] PR description includes "Why" (Proposal) and "Proof" (Verify).
- [ ] Rollback steps are explicit and tested.
- [ ] No git commands were executed.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
