---
name: sdd-orchestrator
description: >
  High-fidelity orchestrator for Spec Driven Development. 
  Manages flow-control, gate-keeping, and state-persistence without implementation.
license: MIT
metadata:
  author: Bitsentry
  version: "1.2"
  family: sdd
  role: orchestrator
  delegate_only: true
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: SDD Orchestrator

## Purpose
Exclusively coordinate the SDD lifecycle by managing the transition between specialized phase skills. It maintains the "Source of Truth" for the session state and enforces quality gates.

## Use When
- **Management**: You need to decide "What happens next?" in a software change.
- **Validation**: You need to ensure a phase's output is sufficient to move forward.
- **Recovery**: A process was interrupted or blocked and needs a pivot or retry.

## Inputs
### Required
- `state.yaml`: Current session snapshot (if exists).
- `last_envelope`: The output from the previous phase skill.
- `skill_registry`: Catalog of available `sdd-*` skills.

## Workflow
*The agent must follow these steps in strict linear order:*

1.  **Reconstruct**: Parse `state.yaml`. If not found, invoke `sdd-init` protocol logic.
2.  **Audit**: Inspect the `last_envelope`.
    - If `status: blocked`, stop and surface the blocker.
    - If `status: success`, verify all declared `artifacts` exist.
3.  **Gate-Check**: Check the **Gating Matrix**:
    - *Gate 1 (Post-Propose)*: Needs explicit user approval to move to Spec/Design.
    - *Gate 2 (Post-Tasks)*: Needs explicit user approval to move to Apply.
4.  **Synthesize**: Update the `artifact_index` and `risk_register` based on the latest phase findings.
5.  **Briefing**: Select the next skill and generate a **Phase Brief** containing only:
    - Target: `{skill-id}`
    - Context: Links to relevant artifacts (not the full content).
    - Objective: The specific success criteria for that phase.

## Phase DAG (Directed Acyclic Graph)
`init` → `explore` → `propose` → **[GATE 1]** → (`spec` & `design`) → `tasks` → **[GATE 2]** → `apply` → `verify` → **[GATE 3]** → `archive`.

## Outputs
Must return a `Result Envelope` that updates the orchestrator's view of the world.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/state.yaml` | The definitive session state.
- **target: engram** | `sdd/{slug}/session_summary` | High-level summary for long-term memory.

## Boundaries
- **NO Engineering**: Do not suggest code, refactors, or technical fixes.
- **NO Deep Dive**: Do not read source code files; rely on `sdd-explore` reports.
- **NO Bypass**: Do not skip `sdd-verify` even for "trivial" changes.
- **Stateless Operation**: Do not rely on hidden memory; everything must be in `state.yaml`.

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Brief status of the flow. (e.g., "Session `auth-fix` moved from `propose` to `spec` after approval").

**Detailed Report**:
- **Current State**: Summary of progress.
- **Validation Audit**: Result of checking the previous envelope.
- **Next Step**: Identification of the next skill to be loaded.

**Artifacts**:
- `sdd/{slug}/state.yaml` (Updated session state).

**Persistence Actions**:
- target: local
  key/path: `sdd/{slug}/state.yaml`
  content summary: Updated state including `{current_phase}` and `{artifact_index}`.

**Next Recommended**: `{next-sdd-skill}`

**Handoffs**:
- **To SDR**: If requirements are too vague for `sdd-spec`.
- **To Judgment Day**: If `sdd-verify` shows recurring quality issues.

**Risks / Gaps**:
- Any detected drift between the proposal and the current design/implementation.


## Handoffs
Compliant with _shared/handoff-contract.md
### 1) To SDR
- **when**: technical research, idea validation, or external comparison is needed.
- **payload (YAML)**:
  - `origin_session_id`: `{slug}`
  - `research_question`: "Specific unknown to solve"
  - `context_artifacts`: [list of paths]
- **expected return**: SDR Result Envelope with validated findings to be integrated into `sdd-explore` or `sdd-spec`.

### 2) To Judgment Day
- **when**: after `sdd-verify` or before `sdd-archive` for critical/high-risk changes.
- **payload (YAML)**:
  - `session_id`: `{slug}`
  - `subject_artifacts`: [design, spec, verify_report]
  - `open_risks`: [list]
- **expected return**: Critique report with verdict: `PASS | PASS WITH NOTES | PARTIAL | BLOCKED`.

### 3) To Go Testing
- **when**: `sdd-verify` requires specialized Go testing patterns (benchmarks, fuzzing, or complex mocks).
- **payload (YAML)**:
  - `target_packages`: [list]
  - `test_evidence`: "current output"
- **expected return**: Go-specific test plan or identified coverage gaps.

### 4) To Issue Creation
- **when**: post-archive follow-ups or deferred technical debt identified.
- **payload (YAML)**:
  - `title`: "Brief description"
  - `body`: "Context + Acceptance Criteria"
  - `priority`: "low | med | high"
- **expected return**: Structured issue draft (Markdown).

### 5) To Branch PR
- **when**: `sdd-verify` is successful and the change is ready for upstream.
- **payload (YAML)**:
  - `branch_name`: `{slug}`
  - `summary`: "Executive summary of changes"
  - `verification_evidence`: "Summary of passing tests"
- **expected return**: PR description and commit plan.


## Quality Checklist
- [ ] Orchestrator logic is strictly delegate-only.
- [ ] Session state is updated and consistent with local artifacts.
- [ ] Gating rules (approvals) are strictly enforced.
- [ ] The next Phase Brief is minimal and deterministic.
