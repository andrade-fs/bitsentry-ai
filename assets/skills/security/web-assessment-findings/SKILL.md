---
name: web-assessment-findings
description: Define structured findings contract for authorized web assessment outcomes.
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

# Skill: web-assessment-findings

## Purpose
Normalize how web-assessment findings are captured, validated, and communicated with evidence and confidence.

## Tooling Policy / Command Safety
This stage consumes only authorized evidence and follows canonical request/tooling policy in `security/web-assessment-requests`.
- no execution by default
- authorized target required
- exact scope required
- prohibited actions required
- evidence logging required
- no secrets exposure

Semantics:
- web-assessment-findings: findings solo con evidencia autorizada.

## Use When
- Request-stage contract exists and permitted activity boundaries are explicit.
- Findings need consistent structure for final reporting.

## Inputs
- Request approval/audit assumptions.
- Evidence expectations and severity rubric.
- Scope and non-goal constraints.

## Workflow
1. Define mandatory finding fields (title, scope anchor, evidence, impact, likelihood, confidence).
2. Define severity and confidence calibration anchors.
3. Define deduplication and evidence grouping rules.
4. Require assumptions/limitations for uncertain or partial evidence.
5. Hand off normalized findings contract to report stage.

## Evidence Contract (minimum, exact anchors)
- Evidence ID
- Source
- Target / URL
- In scope confirmation
- Authorization reference
- Request method
- Request purpose
- Tool class
- Intensity
- Timestamp / testing window
- Result summary
- Relevant headers / status / behavior
- Safety notes
- Redactions
- Linked finding IDs
- Limitations

## Traceability Contract
- authorization → scope → request/evidence → finding → report

## Outputs
- `web-assessment-findings.md` with findings schema and calibration rules.

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
  **key/path**: `web-assessment/{slug}/findings.md`
  **action**: upsert
  **summary**: Findings schema, calibration, and evidence rules.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Findings contract defined with calibration and evidence requirements.

**Next Recommended**: `web-assessment-report`

## Handoffs
- to `security/web-assessment-report` with normalized findings contract.
- Command safety canonical source: `security/web-assessment-requests`.

## Quality Checklist
- [ ] Finding schema includes scope, evidence, and confidence.
- [ ] Calibration anchors are explicit and reusable.
- [ ] Assumptions/limitations are mandatory when evidence is partial.
