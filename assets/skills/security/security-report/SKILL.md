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

## Markdown Report Contract (Required Sections)
The final report MUST include these exact section names (in English):

- Title
- Executive Summary
- Scope
- Methodology
- Repository / Application Context
- Risk Summary
- Findings
- Evidence
- Remediation Plan
- Verification Steps
- Assumptions and Limitations
- Next Steps
- Appendix

## Use When
- Findings gate passed and the review needs closure.
- A decision-ready report is needed for remediation planning.

## Inputs
- `findings.md` normalized findings.
- Prior stage artifacts (`scope.md`, `map.md`, `review.md`).

## Taxonomy Consumption Contract
`security-report` MUST consume the official findings taxonomy from `security-findings`.

### Risk Summary (Required Consumption)
- Aggregate by Category (official enum), Severity, and Confidence.
- Show Impact × Likelihood rationale anchors for top risks.

### Findings (Required Consumption)
- Preserve each finding Category exactly as provided by `security-findings` taxonomy.
- Preserve Severity and Confidence enums exactly.
- Preserve deduplicated canonical finding IDs and grouped evidence references.
- Preserve assumptions/limitations relevant to exploitability confidence.

## Workflow
1. Summarize review scope, method, and guardrail compliance.
2. Present severity distribution, category distribution, and key findings.
3. Validate all required report sections by exact heading token before finalizing.
4. Document residual risks and assumptions.
5. Recommend remediation order and owner-ready next actions.
6. Include explicit statement: this is source review, not live pentest.

## Outputs
- `report.md` with final source security review summary.

## Boundaries
- read-only first.
- OpenCode-first.
- no runtime flow execution.
- no autonomous mode.
- no edits by default.
- agent.bitsentry.permission.edit = deny.
- no .env access.
- no secrets.
- no exploit execution.
- no external target testing.
- no destructive actions.
- no MCP credential mutation.
- CLI debug/plumbing only.
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
- explicit handoff consumed: `security-findings -> security-report`.
- to `support/issue-creation` for remediation issue drafting.
- to `support/judgment-day` for adversarial review on critical risk.

## Quality Checklist
- [ ] Explicitly states source-only review boundaries.
- [ ] Includes all required report sections with exact English names.
- [ ] Findings section consumes the minimum finding contract from `security-findings`.
- [ ] Explicit handoff consumed from `security-findings -> security-report`.
- [ ] Risk Summary consumes official taxonomy categories and Severity/Confidence distributions.
- [ ] Findings preserve category/severity/confidence enums from the findings contract.
- [ ] Residual risks are documented.
- [ ] Recommendations are prioritized and actionable.
