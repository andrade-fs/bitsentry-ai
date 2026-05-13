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

## Tooling Policy / Command Safety
This stage reports boundaries and evidence provenance, and references canonical policy in `security/web-assessment-requests`.
- no execution by default
- authorized target required
- exact scope required
- explicit approval per request
- rate limits required
- stop conditions required
- evidence logging required
- no exploit execution
- no destructive actions

Semantics:
- web-assessment-report: reporta límites, autorización, intensidad y evidencia.

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

## Report Contract (minimum, exact anchors)
- Title
- Authorization Summary
- Scope
- Out of Scope
- Methodology
- Tooling and Intensity
- Request / Evidence Log
- Risk Summary
- Findings
- Remediation Plan
- Verification Steps
- Assumptions and Limitations
- Next Steps
- Appendix

## Traceability Contract
- authorization → scope → request/evidence → finding → report

## Outputs
- `web-assessment-report.md` with final report structure and closure checklist.
- MUST include explicit reporting of authorization state, testing intensity profile, applied limits, and evidence anchors.

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
- Command safety canonical source: `security/web-assessment-requests`.

## Quality Checklist
- [ ] Report sections are explicit and complete.
- [ ] Every finding requires traceable evidence anchors.
- [ ] Limitations and residual risks are mandatory.
