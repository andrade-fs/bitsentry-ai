---
name: security-review
description: Perform bounded read-only security analysis on source artifacts.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: review
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: security-review

## Purpose
Execute focused, read-only source security review and collect evidence-backed risk candidates.

## Use When
- Threat surface hotspots are already mapped.
- User needs security findings without live target testing.

## Inputs
- Hotspot map from `security-map`.
- Source files, dependency declarations, and non-secret config.

## Workflow
1. Review prioritized files for risky patterns and control gaps.
2. Review dependency versions and supply-chain hygiene signals.
3. Review configuration anti-patterns that increase attack surface.
4. Capture evidence snippets and rationale per candidate finding.
5. Forward candidate findings for normalization/severity.

## Outputs
- `review.md` with candidate findings and evidence pointers.

## Boundaries
- NO edits by default.
- NO secret retrieval.
- NO autonomous execution.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: upsert
  **summary**: Candidate findings from read-only review.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Read-only source review completed with evidence-ready candidate findings.

**Next Recommended**: `security-findings`

## Handoffs
- to `security/security-findings` for severity normalization and mitigation mapping.

## Quality Checklist
- [ ] Findings include evidence references.
- [ ] No pentest/exploit/runtime activity was performed.
- [ ] Review stayed within scoped sources.
