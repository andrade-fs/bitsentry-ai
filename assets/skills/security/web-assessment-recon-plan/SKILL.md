---
name: web-assessment-recon-plan
description: Build bounded recon strategy for authorized web assessment without runtime execution.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: recon-plan
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: web-assessment-recon-plan

## Purpose
Define a structured reconnaissance approach for an authorized web target while staying strictly declarative and non-operational.

## Use When
- Authorization and exact scope are explicitly validated.
- The assessment requires planning of surface discovery before any live interaction.

## Inputs
- Authorized target ownership/permission evidence.
- Exact in-scope and out-of-scope definitions.
- Environment, testing window, and intensity constraints.

## Workflow
1. Confirm hard gates remain active from prior stages.
2. Define recon objectives, hypotheses, and evidence collection expectations.
3. Specify allowed/prohibited request classes and required approvals.
4. Produce a staged recon checklist with stop conditions.
5. Hand off a declarative recon plan to mapping stage.

## Outputs
- `web-assessment-recon-plan.md` with objectives, checklist, approvals, and stop conditions.

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
  **key/path**: `web-assessment/{slug}/recon-plan.md`
  **action**: upsert
  **summary**: Declarative recon strategy and approval contract.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Recon strategy defined with explicit approvals and non-operational guardrails.

**Next Recommended**: `web-assessment-map`

## Handoffs
- to `security/web-assessment-map` with approved recon strategy constraints.

## Quality Checklist
- [ ] Recon plan is declarative only (no tooling/runtime execution).
- [ ] Approval points are explicit before any request activity.
- [ ] Stop conditions and prohibited actions are explicitly stated.
