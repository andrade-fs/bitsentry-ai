---
name: jwt-review
description: Read-only review checklist for JWT issuance, validation, and lifecycle risks.
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

# Skill: jwt-review

## Purpose
Detect practical JWT misuse patterns and produce evidence-backed findings without runtime validation.

## Use When
- Token auth is used (Bearer/JWT access + refresh tokens).
- Security review scope includes auth middleware, gateways, or API edge.

## Inputs
- Token generation/verification code and auth middleware.
- Public key handling, algorithm config, and expiry/rotation policy docs.

## Workflow
1. Identify token issue/verify paths and key management approach.
2. Quick triage checklist:
   - `exp`, `iat`, `iss`, `aud` checks enforced.
   - Algorithm is explicit and not user-influenced.
   - Refresh token rotation/invalidation strategy exists.
   - Revocation path exists for compromised sessions.
3. Risk patterns:
   - Accepting `none`/unexpected algorithm.
   - Missing audience/issuer validation.
   - Excessive token TTL for privileged scopes.
4. Safe evidence capture: claims validation code references and configuration lines (no keys/tokens).
5. Emit actionable candidate finding with severity orientation.

## Outputs
- JWT-focused candidate findings with impact, evidence pointers, likely exploit preconditions, and mitigation.

## Boundaries
- Read-only first; no edits by default.
- No `.env` access or secrets/key material extraction.
- No exploit execution, no external target testing, no destructive actions.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: JWT review candidates and safe evidence references.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
JWT handling was reviewed for validation strictness and lifecycle risk with evidence-safe notes.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity guide (High for auth bypass/token forgery vectors, Medium for lifecycle weakness), false-positive caveats, and fix path.
- To `security/security-report`: summarize residual risk if key/issuer architecture must change.

## Quality Checklist
- [ ] Claims validation coverage documented.
- [ ] Algorithm/key handling reviewed.
- [ ] Common false positives addressed (gateway validates claims upstream).
- [ ] Finding output is actionable and bounded.
