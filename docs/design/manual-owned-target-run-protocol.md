## Manual Owned Target Run Protocol (Phase 7.16D, design-only)

### 1) Purpose / Non-goals

**Purpose**
- Define a strict manual protocol for the first future owned-target request under Level 3 operational gate semantics.
- Preserve policy-first execution boundaries where approvals are exact and per-request.
- Provide a human-checkable checklist and PASS/FAIL protocol before any future real run.

**Non-goals (7.16D)**
- No request real.
- No bitsentry.xyz live testing.
- No target vivo testing.
- No CLI/OpenCode execution.
- No runtime flow execution.
- No crawler/scanner/fuzzing/payload/auth/cookies/POST/redirect-follow behavior.

### 2) Pre-run Checklist (mandatory)

Must be explicitly satisfied before any future manual run:
- target owner declared
- scope host exacto
- in-scope URL exacta
- out-of-scope explícito
- method permitido
- one request only
- rate limit
- timeout
- max response size
- max preview size
- stop conditions
- evidence policy
- redaction policy
- no redirects followed
- no auth/cookies
- no POST
- no crawler/scanner
- no background execution

### 3) Assessment Session Context (example, documentation only)

This is an example protocol context, not executable in 7.16D:

- `session_id: manual-owned-target-run-001`
- `target_owner_declared: true`
- `scope_hosts: ["bitsentry.xyz"]`
- `in_scope_urls: ["https://bitsentry.xyz/"]`
- `out_of_scope: ["all other domains", "subdomains unless explicitly listed", "admin/authenticated areas"]`
- `environment: production-owned`
- `intensity: low_noise`
- `allowed_methods: ["HEAD"]`
- `allowed_paths: ["/"]`
- `prohibited_actions: [POST, auth/cookies, redirects followed, crawler, scanner, fuzzing, payloads, brute force, DoS]`
- `execution_allowed: false by default`

Important:
- `bitsentry.xyz` appears only as an explicit documentation example.
- 7.16D does not perform execution, target testing, or network requests.

### 4) Request Plan Example (manual protocol template)

Preferred first request:
- `request_ref: first-owned-head-root`
- `method: HEAD`
- `url: https://bitsentry.xyz/`
- `purpose: passive header/security posture evidence`
- `expected_evidence: status code, response headers, content type if available, no body required`
- `risk: low`
- `redirects: observe only, do not follow`

Secondary request:
- `request_ref: first-owned-robots`
- `method: GET`
- `url: https://bitsentry.xyz/robots.txt`
- `purpose: passive discovery-file evidence`
- `expected_evidence: status code, small redacted body preview`
- `risk: low`
- `redirects: observe only, do not follow`

### 5) Approval Token (exact)

Accepted exact example:
- `APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/`

generic approval invalid examples:
- `adelante`
- `ok`
- `sí`
- `hazlo`
- `approve all`
- `prueba todo`
- `APPROVE url=https://bitsentry.xyz/` (missing request_ref)
- `APPROVE request_ref=...` with method/url mismatch

### 6) Validation Rules (before any future execution)

- `request_ref` must match
- `method` must match
- `URL` must match
- scope must match
- approval not expired
- `approval_text_or_hash` present
- limits present
- stop conditions present
- execution mode `execute_approved` only at executor stage
- mismatch => deny-before-transport

### 7) Active limits that must be present

- one request only
- method allowlist
- strict path allowlist
- rate_limit_per_minute
- request_budget
- timeout_seconds
- max_response_size_bytes
- max_preview_size_bytes
- stop_conditions
- no redirects followed

### 8) Evidence capture contract

Expected evidence metadata:
- `session_id`
- `request_ref`
- `approval_id`
- `evidence_id`
- timestamp
- policy decision
- redactions applied
- safety notes
- limitations

Evidence handling requirements:
- evidence redacted
- no secrets exposure
- no full sensitive payload persistence by default

### 9) Expected output envelope (future manual run)

- `ExecutionResult`
- `PassiveCheckResult`
- `evidence_id`
- `redactions applied`
- `safety notes`
- `limitations`
- no confirmed finding by default

Optional next steps after manual run:
- build SurfaceMap
- build RiskHypotheses
- build WebTestPlan

### 10) Stop conditions / hard failure conditions

Stop immediately on:
- unexpected 5xx
- timeout
- response too large
- redirect out of scope
- sensitive data detected
- user stop
- policy mismatch
- request budget exhausted

### 11) PASS/FAIL Protocol

PASS if:
- approval exact
- one request only
- no redirects followed
- evidence redacted
- traceability present
- no prohibited behavior

FAIL if:
- multiple requests
- generic approval accepted
- redirect followed
- POST/auth/crawler/scanner used
- evidence stores secrets
- target out of scope
- missing audit trail

### 12) Prohibited behaviors (explicit)

- no autonomous pentest
- no mass scanning
- no DoS/load testing
- no brute force
- no credential attacks
- no exploit chains
- no destructive checks
- no background execution
- no secrets exposure

### 13) Status for 7.16D

- completed / PASS WITH NOTES
- design-only
- no runtime
- no live target execution
- protocol ready
- future operational run pending
