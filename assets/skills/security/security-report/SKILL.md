---
name: security-report
description: Generate final source security review report with guardrail compliance.
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

# Skill: security-report

## Purpose
Produce the final source security review report, including residual risk and next actions.

## Use When
- Findings gate passed and the review needs closure.
- A decision-ready report is needed for remediation planning.

## Inputs
- `findings.md` normalized findings.
- Prior stage artifacts (`scope.md`, `map.md`, `review.md`).

## Workflow
1. Summarize review scope, method, and guardrail compliance.
2. Present severity distribution and key findings.
3. Document residual risks and assumptions.
4. Recommend remediation order and owner-ready next actions.
5. Include explicit statement: this is source review, not live pentest.

## Outputs
- `report.md` with final source security review summary.

## Boundaries
- NO flow runtime execution.
- NO changes to credentials, secrets, or environment state.
- NO destructive actions.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/report.md`
  **action**: upsert
  **summary**: Final source security review report and residual risk log.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Final source security review report generated with risk-ranked findings and remediation guidance.

**Next Recommended**: `support/issue-creation` or close.

## Handoffs
- to `support/issue-creation` for remediation issue drafting.
- to `support/judgment-day` for adversarial review on critical risk.

## Quality Checklist
- [ ] Explicitly states source-only review boundaries.
- [ ] Residual risks are documented.
- [ ] Recommendations are prioritized and actionable.
