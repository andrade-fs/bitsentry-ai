## Controlled HTTP Executor (Phase 7.8C, offline executable contracts)

### 1) Purpose / Non-goals

**Purpose**
- Implement an offline executable boundary for controlled web requests in `internal/securityweb`.
- Enforce strict per-request approval + policy + redaction invariants using deterministic in-memory transport only.

**Non-goals (7.8C)**
- No real network execution.
- No `net/http` client.
- No DNS/TLS/crawler/scanner/runtime flow execution.
- No live target testing.

### 2) Architecture (7.8C)

- `OfflineControlledExecutor` (separate component)
  - Executes only with explicit `ExecutionApproval`.
  - Uses `FakeTransport` only.
- `FakeTransport`
  - Primary key: `request_ref`.
  - Fallback key: `method + url` only when `request_ref` is empty.
- `ExecutionResult`
  - Includes execution metadata + evidence linkage + policy violations.

### 3) ExecutionApproval contract

Required behavior:
- `expires_at` is mandatory and authoritative for expiration checks.
- `ttl_seconds` is metadata/audit (non-authoritative).
- missing `expires_at` => deny.
- expired `expires_at` => deny.

Strict approval matching:
- request ID mismatch => deny.
- method mismatch => deny.
- URL mismatch => deny.

### 4) Execution and policy rules (implemented)

- `execute_approved` requires valid approval.
- GET/HEAD allowed only with valid approval.
- POST denied.
- one request at a time.
- `follow_redirects=false` default.
- out-of-scope redirect denied; new approval required.
- FakeTransport only; no real network.

Policy violation codes were extended additively with prefixes:
- `approval_*`
- `redirect_*`
- `limiter_*`

### 5) ExecutionResult metadata and redaction

`ExecutionResult` includes:
- `evidence_id`
- `body_truncated`
- `body_preview_redacted` (may be empty on denied cases)
- `response_size`
- `max_preview_size`

Redaction coverage (response evidence):
- `Authorization`
- `Cookie`
- `Set-Cookie`
- `Bearer` tokens
- sensitive query params (`token`, `access_token`, `api_key`, `password`, `secret`)

### 6) Testing strategy (offline only)

- Deterministic contract tests in `internal/securityweb/executor_test.go`.
- Static import guards in package tests:
  - no `net/http` import
  - no `os/exec` import
- Repository-level validation via `go test ./...`.

### 7) Relationship to adapter design

This 7.8C implementation operationalizes (offline-only) the execution boundary planned in:
- `docs/design/web-request-adapter.md`

All guardrails remain: no live target testing, no offensive tooling execution, no credential mutation.

### 8) Legacy 7.8B contract anchors (kept for compatibility tests)

This section intentionally preserves prior 7.8B anchor wording:

- `execute_approved only`
- `per-request approval`
- `approval_id`
- `approved_request_id`
- `approved_method`
- `approved_url`
- `expires_at`
- `approval_text_or_hash`
- `follow_redirects=false`
- `no redirects followed by default`
- `out-of-scope redirect requires new approval`
- `GET/HEAD only first MVP`
- `one request at a time`
- `no POST`
- `no payloads`
- `no crawler`
- `no scanner`
- `no background execution`
- `body_preview_redacted`
- `headers_redacted`
- `no full response stored by default`
- `FakeTransport only for future tests`
- `no real network in 7.8B`


## 13) 7.8D Hardening Addendum

This phase remains **offline-only** and adds strict hardening rules before any future network-capable phase.

### Approval hardening
- execute_approved requires valid approval
- approval_id required
- approved_by required
- approval_text_or_hash required
- approved_scope_ref required
- approved_execution_mode required and must be `execute_approved`
- approved_tool_class required and must match request tool class
- approved_intensity required and must match context intensity
- approved_request_id/method/url must match request
- expires_at required and must be valid (TTLSeconds is metadata/audit only)

### Limits and precedence
- MaxRequests required
- RateLimit required
- MaxDuration required
- TimeoutSeconds required
- MaxResponseSizeBytes required
- MaxPreviewSizeBytes required
- StopConditions required
- approval limits cannot exceed session/context limits (`approval_exceeds_context_limits`)

### Redirect policy
- follow_redirects=false
- no redirects followed by default
- redirect_location_invalid when location is empty/invalid/unsafe
- redirect_out_of_scope when redirect target is outside scope

### Traceability
- request_ref -> approval_id -> evidence_id -> execution_result
- ExecutionResult includes ApprovalID, ApprovedBy, ApprovalExpiresAt

### Redaction and storage
- body_preview_redacted
- headers_redacted
- body preview may be empty after safe redaction
- no full response stored by default
- no real network in 7.8D

- approval_scope_missing
- approval_execution_mode_missing
- approval_tool_class_missing
- approval_intensity_missing
- approval_actor_missing
- approval_proof_missing