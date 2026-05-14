## Phase 7.17B — First Owned Target HEAD Evidence (factual run record)

### Scope
- Single approved request only.
- No fallback and no additional requests.

### Request
- request_ref: `first-owned-head-root`
- method: `HEAD`
- url: `https://bitsentry.xyz/`
- approval_id: `approval-first-owned-head-root-001`
- evidence_id: `WEB-EV-first-owned-head-root-approval-first-owned-head-root-001`

### ExecutionResult (factual)
- status_code: `200`
- redirect_observed: `false`
- redirect_location: ``
- duration_ms: `369`
- policy_decision: `allow`
- follow_redirects: `false`
- safety_notes:
  - `no full response stored by default`

### Relevant headers observed (redacted contract path)
- `Content-Security-Policy`: present
- `Strict-Transport-Security`: present
- `X-Content-Type-Options`: present (`nosniff`)
- `X-Frame-Options`: present (`SAMEORIGIN`)
- `Referrer-Policy`: `origin-when-cross-origin`

### PassiveHeaderCheck (factual)
- CheckID: `passive_headers_mvp`
- EvidenceID: `WEB-EV-first-owned-head-root-approval-first-owned-head-root-001`
- Notable observation:
  - `Referrer-Policy`: `weak`
- Candidate finding:
  - `hf-referrer-weak` (Low severity, High confidence)

### Guardrail statement
- Candidate finding remained candidate-only.
- No confirmed finding by default.
