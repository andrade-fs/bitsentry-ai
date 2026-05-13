---
name: web-assessment-requests
description: Prepare explicit request-authorization contract and templates for web assessment.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: requests
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: web-assessment-requests

## Purpose
Define the request approval workflow and template structure, ensuring no live request is executed by default.

## Use When
- Test plan is approved and safety boundaries are explicit.
- A formal request gate is needed before any target interaction.

## Inputs
- Approved test matrix and checkpoints.
- Tool allowlist and prohibited actions list.
- Rate limits and credential-handling policy.

## Workflow
1. Define mandatory fields for per-request authorization.
2. Provide request-template format including purpose, scope tie, and expected evidence.
3. Define deny conditions for out-of-scope/high-risk requests.
4. Add audit trail expectations for approvals/rejections.
5. Hand off request contract to findings stage.

## Outputs
- `web-assessment-requests.md` with request contract, templates, and deny rules.

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
  **key/path**: `web-assessment/{slug}/requests.md`
  **action**: upsert
  **summary**: Request authorization and template contract.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Request gate and templates defined; execution remains explicitly disabled by default.

**Next Recommended**: `web-assessment-findings`

## Handoffs
- to `security/web-assessment-findings` with approved request-contract assumptions.

## Quality Checklist
- [ ] Explicit approval required before each request.
- [ ] Deny conditions cover out-of-scope and prohibited actions.
- [ ] No runtime tooling/request execution is implied.
