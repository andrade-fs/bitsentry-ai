---
name: security-findings
description: Normalize findings with severity, confidence, and mitigation guidance.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: findings
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: security-findings

## Purpose
Convert review candidates into actionable security findings with clear severity and remediation direction.

## Use When
- `security-review` has candidate findings.
- A formal findings gate is required before final reporting.

## Inputs
- Candidate findings with evidence.
- Scope and risk rubric from previous stages.

## Workflow
1. Deduplicate and normalize finding statements.
2. Assign severity and confidence using declared rubric.
3. Add impact, likelihood, and exploit preconditions (conceptual, no execution).
4. Add mitigation guidance and verification hints.
5. Output findings manifest for final report.

## Outputs
- `findings.md` with normalized findings register.

## Boundaries
- NO vulnerability proof-of-concept execution.
- NO runtime mutation or environment manipulation.
- NO credential or secret operations.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/findings.md`
  **action**: upsert
  **summary**: Normalized findings with severity and mitigations.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Findings gate completed with severity-ranked, actionable entries.

**Next Recommended**: `security-report`

## Handoffs
- to `security/security-report` for final synthesis.
- to `support/issue-creation` when follow-up tracking is required.

## Quality Checklist
- [ ] Every finding has severity and evidence.
- [ ] Mitigations are actionable and bounded.
- [ ] No secrets/runtime/pentest behavior introduced.
