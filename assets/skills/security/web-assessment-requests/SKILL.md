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

## Assessment Session Context
- This is the canonical gate for execution mode changes: planning_only -> dry_run -> execute_approved -> retest.
- planning_only and dry_run remain non-operational in this phase.
- execute_approved requires explicit approval before live requests/tooling.
- Retest intent must preserve original scope/authorization and evidence log continuity.

## Tooling Policy / Command Safety
This skill is the canonical source for web-assessment request/tooling safety policy.

### Core Safety Anchors (mandatory)
- no execution by default
- explicit approval per request
- authorized target required
- exact scope required
- allowed tools required
- prohibited actions required
- rate limits required
- stop conditions required
- evidence logging required
- no exploit execution
- no destructive actions
- no DoS/load testing
- no credential attacks
- no mass scanning
- no out-of-scope scanning
- no secrets exposure
- no brute force
- no password spraying
- no aggressive fuzzing
- no exfiltration

### Tool Classes (explicit)
- Passive inspection
- Single-request verification
- Low-noise mapping
- Authenticated test with provided test credentials
- Prohibited / requires separate explicit approval

### Request Safety Contract
Every proposed request MUST include:
1. Authorization evidence reference.
2. Exact scope anchor (target/environment/path/asset).
3. Tool class and specific allowed tool selection.
4. Intensity/rate profile and stop condition reference.
5. Expected evidence artifact(s) and logging destination.
6. Explicit per-request approval record before execution.
7. Assessment Session Context execution mode and rationale.

Default deny rules:
- Deny when authorization is missing, stale, or ambiguous.
- Deny when scope is broad/implicit/open-ended.
- Deny when tool class maps to prohibited activity without separate explicit approval.
- Deny when evidence logging plan is absent.
- Deny when stop conditions are undefined.
- Deny when requested action conflicts with hard guardrails.

### Canonical Delegation Rule
Other web-assessment skills MUST summarize only the subset relevant to their stage and defer full command safety semantics to `security/web-assessment-requests`.

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
- [ ] Core Safety Anchors are preserved with exact tokens.
- [ ] Tool Classes are listed explicitly and mapped to approvals.
