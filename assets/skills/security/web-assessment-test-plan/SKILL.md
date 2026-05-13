---
name: web-assessment-test-plan
description: Define a safe and approved web assessment test plan without executing tests.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: test-plan
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: web-assessment-test-plan

## Purpose
Create a controlled test matrix for authorized web assessment, including guardrails, approvals, and abort criteria.

## Assessment Session Context
- Intent: transition from Mapping to Test Plan with contract-first constraints.
- Scope and ownership/autorización remain mandatory anchors.
- Controlled Checks are planned only; no runtime, no tooling real, no live requests.
- Expert skills can be referenced as planning aids only.
- Execution mode drives depth: planning_only, dry_run, execute_approved, retest.

## Tooling Policy / Command Safety
This stage defines controls only and defers full command safety semantics to `security/web-assessment-requests`.
- no execution by default
- explicit approval per request
- authorized target required
- exact scope required
- rate limits required
- stop conditions required
- evidence logging required
- no DoS/load testing

Semantics:
- web-assessment-test-plan: define intensidad, rate limits, stop conditions y logging plan.

## Use When
- Mapping is complete and priorities are known.
- The team needs an explicit, safe sequence before any request activity.

## Inputs
- Prioritized web surface map.
- Authorization and scope gates.
- Environment/testing window/intensity constraints and rate limits.

## Workflow
1. Define test objectives per prioritized area.
2. Build test matrix with methods, expected evidence, and risk level.
3. Controlled Checks: attach mandatory approval checkpoints before request execution.
4. Define abort conditions and escalation paths.
5. Validate: confirm prohibited actions remain blocked by default.
6. Hand off approved declarative test plan to requests stage.

## Outputs
- `web-assessment-test-plan.md` with matrix, checkpoints, and abort criteria.

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
  **key/path**: `web-assessment/{slug}/test-plan.md`
  **action**: upsert
  **summary**: Declarative test matrix and guardrail checkpoints.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Test plan defined with explicit checkpoints and non-destructive controls.

**Next Recommended**: `web-assessment-requests`

## Handoffs
- to `security/web-assessment-requests` with approved test matrix.
- Command safety canonical source: `security/web-assessment-requests`.

## Quality Checklist
- [ ] Every test line has evidence expectation and risk label.
- [ ] Approval checkpoints are explicit before requests.
- [ ] Abort conditions include safety and authorization triggers.
