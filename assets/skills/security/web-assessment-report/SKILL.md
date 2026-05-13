---
name: web-assessment-report
description: Define final report contract for authorized web assessment in declarative mode.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: report
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: web-assessment-report

## Purpose
Produce a consistent final reporting contract summarizing risk posture, findings, and residual risk without enabling operational runtime behavior.

## Use When
- Findings contract is finalized.
- Stakeholders need executive and technical reporting structure.

## Inputs
- Normalized findings schema and calibration rules.
- Scope definition and assumptions/limitations.
- Guardrails and prohibited-action context.

## Workflow
1. Define report sections (executive summary, scope, methodology constraints, findings, remediation, residual risk).
2. Require traceability from each finding to scope and evidence anchors.
3. Add explicit limitations and non-executed areas.
4. Define remediation prioritization and ownership hints.
5. Emit final declarative report contract and closure checklist.

## Outputs
- `web-assessment-report.md` with final report structure and closure checklist.

## Boundaries
- Authorization required before live target interaction.
- Exact scope required.
- Explicit permission before requests.
- No external target testing without authorization.
- No exploit execution by default.
- No destructive actions.
- No DoS/load testing.
- No credential attacks.
- No secrets exposure.
- No MCP credential mutation.
- No runtime flow execution.
- No autonomous mode.
- OpenCode-first.
- CLI debug/plumbing only.
- `agent.bitsentry.permission.edit = deny`.

## Persistence Actions
- **target**: local
  **key/path**: `web-assessment/{slug}/report.md`
  **action**: upsert
  **summary**: Final report contract with residual risk and closure structure.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Report contract defined with traceability, limitations, and residual risk communication.

**Next Recommended**: `support/issue-creation`

## Handoffs
- to `support/issue-creation` for remediation tracking artifact creation when requested.

## Quality Checklist
- [ ] Report sections are explicit and complete.
- [ ] Every finding requires traceable evidence anchors.
- [ ] Limitations and residual risks are mandatory.
