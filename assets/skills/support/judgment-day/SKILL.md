---
name: support-judgment-day
description: >
  Adversarial quality review skill. Provides a cold, objective audit of 
  artifacts to issue a final quality verdict.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/engram-convention.md
---

# Skill: support-judgment-day

## Purpose
To provide an uncompromising quality gate. This skill re-evaluates all evidence produced during SDD (Implementation/Testing) or SDR (Research/Validation) to ensure no assumptions were made and no critical security or logic gaps remain.

## Use When
- High-risk changes are ready for final approval.
- An extra layer of scrutiny is needed before `sdd-archive` or `sdr-archive`.
- You suspect "hallucination drift" or "lazy validation" in previous phases.

## Workflow: The "Adversarial Audit Protocol"
1.  **Evidence Stress-Test**: Don't just read the "PASS" verdict; verify the logs/output provided in `verify.md` or `validate.md`.
2.  **Assumption Hunting**: Identify statements starting with "It should...", "I assume...", or "Likely...". Convert them into `BLOCKED` status until proven.
3.  **Control Gap Analysis**: Check for missing security controls, edge cases not covered, or performance bottlenecks ignored.
4.  **Risk Quantification**: List every residual risk and categorize it by Impact vs. Probability.
5.  **Final Verdict Issuance**:
    - **PASS**: Flawless evidence, zero critical risks.
    - **PASS WITH NOTES**: Minor technical debt, but safe to proceed.
    - **PARTIAL**: Goal met partially; needs follow-up issues.
    - **BLOCKED**: Fundamental flaws, safety risks, or insufficient evidence.

## Outputs
### Judgment Manifesto (Markdown)
- **Final Verdict**: Large, clear status.
- **Critical Issues**: Non-negotiable points that must be fixed.
- **Residual Risks**: What we are "living with" if we merge/publish.
- **Remediation List**: Concrete steps to reach a `PASS`.

## Boundaries
- **NO Implementation**: This is an audit-only skill. Do not fix the code.
- **NO Politeness**: Be direct, cold, and objective. "Better safe than sorry."

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: `reviews/judgment-day-{slug}.md`
  **action**: upsert
  **summary**: Final adversarial audit and quality verdict.
- **target**: engram
  **key**: `support/judgment-day/{slug}`
  **action**: upsert
  **summary**: Historical record of the quality gate.

## Result Envelope
**Status**: `success | blocked` (Success here means the audit was performed, regardless of the verdict).

**Executive Summary**:
Judgment Day for `{slug}` concluded. Verdict: `{VERDICT}`. `{N}` critical issues identified.

**Next Recommended**: 
- `sdd-archive` (if PASS).
- `sdd-apply` / `sdr-research` (if BLOCKED).

**Handoffs**:
- **to**: `original-flow-orchestrator`
  **reason**: Mandatory remediation required or clearance granted.

**Quality Checklist**:
- [ ] Verdict is backed by specific evidence found in artifacts.
- [ ] No "soft" approvals for high-risk gaps.
- [ ] Remediation steps are actionable.