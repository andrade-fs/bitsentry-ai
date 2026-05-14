## First Owned Target HEAD Dry-Run Gate (Phase 7.17A, design-only)

### 1) Scope and boundary
- no request is executed in 7.17A
- design-only
- no runtime
- no live target execution
- no CLI/OpenCode execution
- no crawler/scanner/fuzzing/payload/auth/cookies/POST/redirect-follow

### 2) Assessment Session Context (dry-run artifact)
- `session_id: manual-owned-target-run-001`
- `target_owner_declared: true`
- `scope_hosts: ["bitsentry.xyz"]`
- `in_scope_urls: ["https://bitsentry.xyz/"]`
- `out_of_scope: ["all other domains", "all subdomains unless explicitly listed", "admin/authenticated areas", "third-party assets"]`
- `environment: production-owned`
- `intensity: low_noise`
- `execution_allowed=false by default`

### 3) Active request plan (only one request)
- one request only
- `request_ref: first-owned-head-root`
- `method: HEAD`
- `url: https://bitsentry.xyz/`
- `HEAD https://bitsentry.xyz/`
- purpose: passive header/security posture evidence
- expected evidence: status code, response headers, content type if available, redirect location if observed
- risk: low
- redirects: observe only, do not follow

### 4) Secondary option (not active in this gate)
- `first-owned-robots` is not part of the active gate
- It remains a future/secondary alternative only and is not bundled with 7.17A.

- first-owned-robots is not part of the active gate

### 5) Limits (must be explicit)
- one request only
- timeout: 10s
- max response size: small
- max preview size: 0 (HEAD-first posture)
- rate limit: 1 request / manual approval
- stop conditions:
  - timeout
  - unexpected redirect out of scope
  - policy mismatch
  - user stop
  - sensitive data observed

### 6) Approval draft and exact token
- `approval_id: approval-first-owned-head-root-001`
- `request_ref: first-owned-head-root`
- `method: HEAD`
- `url: https://bitsentry.xyz/`

exact approval token is required:
- `APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/`

generic approval rejected examples:
- `adelante`
- `ok`
- `approve all`
- `hazlo`
- `APPROVE` without `request_ref`/`method`/`url`
- any token with method/url/request_ref mismatch

### 7) Expected result shape (future execution phase only)
If a future phase explicitly executes after exact approval:
- `ExecutionResult`
- `EvidenceID`
- `redactions_applied`
- `safety_notes`
- `limitations`
- Passive header check result
- no confirmed finding by default

### 8) PASS/FAIL Protocol
PASS if:
- no request is executed in 7.17A
- exact approval token is required
- request plan has one request only
- scope and limits are explicit
- generic approval rejected

FAIL if:
- any request executes
- generic approval is accepted
- more than one request is planned
- POST/auth/cookies/crawler/scanner appear
- redirect follow is allowed

### 9) Status
- completed / PASS WITH NOTES
- design-only
- no runtime
- no live target execution
- dry-run gate ready
- future manual execution pending


### 10) 7.18A Formal preflight command

Formal entrypoint for manual dry-run validation:

`bitsentry-ai securityweb manual-preflight --request-ref first-owned-head-root --method HEAD --url https://bitsentry.xyz/ --approval "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/" --dry-run`

Notes:
- 7.18A preflight does not execute requests (`would_execute=false`).
- Any approval mismatch returns deny.


### 11) 7.18B transition note

- 7.18B defines execute-entrypoint contract only (design-only).
- No `manual-execute` implementation is added in this phase.
- Continue using `manual-preflight` as the only formal command path.
