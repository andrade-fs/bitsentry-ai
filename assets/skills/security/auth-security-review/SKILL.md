---
name: auth-security-review
description: Focused checklist for authentication and session-control weaknesses in source review.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: review-support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: auth-security-review

## Purpose
Provide practical, read-only triage for authentication and session security findings during `source-security-review`.

## Use When
- Login, session, MFA, password reset, or access-control code is in scope.
- `security-map` identified auth boundaries or identity trust paths.

## Inputs
- Auth-related source files, middleware, route guards, and config (non-secret only).
- Session/token policy docs and reviewer scope constraints.

## Workflow
1. Confirm entry points: login/signup/reset/session refresh/logout.
2. Run quick triage checklist:
   - Account lockout/rate limiting present.
   - Password reset tokens are short-lived and single-use.
   - Session invalidation on logout/password change.
   - MFA bypass paths absent for privileged actions.
3. Check risk patterns:
   - Missing server-side authorization checks.
   - Long-lived sessions without rotation.
   - Weak recovery flow enabling account takeover.
4. Capture safe evidence (file path, function name, pseudocode snippet, no secrets).
5. Draft actionable finding and handoff to `security-findings`.

## Outputs
- Candidate auth finding entries with: risk statement, impact scope, safe evidence pointer, and mitigation direction.

## Boundaries
- Read-only first; NO edits by default.
- NO `.env` access, NO secrets retrieval/exposure.
- NO exploit execution, NO runtime flow execution, NO autonomous mode.
- NO external target testing, NO destructive actions, NO MCP credential mutation.
- OpenCode-first; CLI is debug/plumbing only.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: Auth-focused candidate findings and evidence pointers.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Auth controls were reviewed with bounded checklists and evidence-safe candidates.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (often High for auth bypass, Medium for session-hardening gaps), false-positive notes, and remediation hints.
- To `security/security-report`: include residual risk if architectural change is required.

## Quality Checklist
- [ ] Includes when/where auth logic was inspected.
- [ ] Includes quick triage checklist results.
- [ ] Includes common false-positive check (client-only guard vs server enforcement).
- [ ] Produces actionable finding format (what/impact/evidence/fix).
