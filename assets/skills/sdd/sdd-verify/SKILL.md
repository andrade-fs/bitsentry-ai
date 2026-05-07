---
name: sdd-verify
description: >
  Quality gate that validates the implementation against the original specification.
  Focuses on 'Correctness' and 'Spec Compliance'.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: verify
  status: declarative
  requires:
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-verify

## Purpose
Act as the final quality filter. This skill ensures that the code produced in `sdd-apply` strictly adheres to the behavior defined in `sdd-spec` and the technical integrity defined in `sdd-design`.

## Use When
- `sdd-apply` is complete.
- You need a PASS/FAIL verdict before merging or archiving.

## Inputs
- **Acceptance Criteria**: `sdd/{slug}/spec.md`.
- **Implementation Log**: `sdd/{slug}/apply.md`.
- **Test Evidence**: Logs, test results, or build status from the environment.

## Workflow: The "Rigorous Validation Protocol"
1.  **Spec Compliance Matrix**: Map each Acceptance Criterion (AC) to a test result or code verification.
2.  **Regression Check**: Verify that existing functionality remains intact (based on `explore` dependencies).
3.  **Static Analysis Review**: Check for linting, types, and architectural violations.
4.  **Risk Review**: Re-evaluate the risks identified in `propose` to see if they were mitigated.
5.  **Verdict Generation**: Determine the final status: `PASS`, `PASS WITH NOTES`, `PARTIAL`, or `BLOCKED`.

## Outputs
### Verification Report (Markdown)
Must include:
- **Compliance Checklist**: Status of each AC (Met/Not Met).
- **Test Results**: Summary of automated or manual verification.
- **Residual Risks**: Any "known issues" that remain.
- **Final Verdict**: Clear signal for the orchestrator.

## Boundaries
- **NO Implementation**: Do not fix bugs found during verification (report them instead).
- **NO Spec Changes**: Do not "re-interpret" the spec to make a failing test pass.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/verify.md` | The final quality report.
- **target: engram** | `sdd/{slug}/verification_verdict` | Final outcome for session history.

## Result Envelope
**Status**: `success | partial | blocked`
**Next Recommended**: `sdd-archive` (or return to `sdd-apply` if blocked)