---
name: security-init
description: >
  Bootstrap declarative source-security-review with strict read-only-first guardrails.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: init
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
    - _shared/opencode-convention.md
---

# Skill: security-init

## Purpose
Initialize Source Security Review MVP with explicit constraints and route framing for source-only analysis.

## Use When
- A request includes security/appsec/threat/risk signals for repository review.
- The user needs a read-only source security review in OpenCode context.

## Inputs
- User objective and threat/risk concerns.
- Repository context and declared scope hints.

## Workflow
1. Confirm this is source code security review (not pentest/live assessment).
2. Declare guardrails: read-only first, no `.env` access, no secrets handling.
3. Set flow constraints: no exploit execution, no external target testing, no destructive actions.
4. Enforce platform boundaries: OpenCode-first, CLI route decide is debug/plumbing, no runtime flow execution.
5. Emit init envelope with required gates and next stage recommendation.

## Outputs
- `init.md` with objective, non-goals, guardrails, and acceptance boundaries.

## Boundaries
- NO edits by default.
- NO MCP credential mutation.
- NO autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/init.md`
  **action**: upsert
  **summary**: Initialization summary and hard guardrails.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Source-security-review initialized with read-only-first and no-runtime constraints.

**Next Recommended**: `security-scope`

## Handoffs
- to `security/security-scope` when initialization is accepted.

## Quality Checklist
- [ ] Explicitly states source-only (no live target testing).
- [ ] Includes no `.env` access and no secrets policy.
- [ ] Includes `agent.bitsentry.permission.edit = deny` compatibility.
