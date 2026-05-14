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

## 9) 7.9A Future Real HTTP Transport Boundary (design-only)

This phase is design-only and keeps runtime behavior unchanged.

### Explicit package separation

- `internal/securityweb` = core offline-safe.
- `internal/securitywebhttp` = future real transport package (not implemented in 7.9A).

### Policy Enforcement Point (PEP)

- **Policy Enforcement Point**
- **OfflineControlledExecutor is the single Policy Enforcement Point**
- **real transport must not decide policy**
- **deny before transport**
- **transport receives only policy-approved requests**

This means the future transport layer executes only what the executor has already approved under policy; it never upgrades, relaxes, or bypasses approvals.

### Execution contract preserved for future transport

- `execute_approved only`
- `per-request approval`
- `GET/HEAD only first MVP`
- `follow_redirects=false`
- `no redirects followed by default`
- `no full response stored by default`
- `body preview capped and redacted`

### Testing strategy for future real transport

- `no external network tests`
- `httptest only for future real transport tests`

7.9A does not introduce `net/http`, real requests, runtime flow execution, crawler behavior, or live target testing.

### 10) 7.9B Real HTTP transport skeleton (httptest-only)

Phase 7.9B introduces a minimal real transport implementation in:

- `internal/securitywebhttp/transport.go`

Boundary and invariants:

- **OfflineControlledExecutor remains the single Policy Enforcement Point**.
- **transport does not enforce scope/approval/method/tool/rate/budget policy**.
- transport executes request I/O only (HTTP send/receive + normalized response mapping).
- redirect not followed by default.
- Location captured.
- body preview capped.
- `BodyTruncated` marked when cap is exceeded.
- no full response stored by default.
- timeout required in transport constructor.
- transport errors normalized safely (no sensitive target details by default).

Testing contract for 7.9B:

- `httptest` only.
- no external network tests.
- GET/HEAD pass coverage.
- redirect no-follow + Location capture coverage.
- timeout enforcement coverage.
- static import guard in `internal/securitywebhttp/static_imports_test.go` denying `os/exec` and `syscall`.


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

## 14) 7.9C Executor Transport Injection (implemented)

Phase 7.9C keeps architecture boundaries intact while enabling runtime transport injection:

- `OfflineControlledExecutor` now depends on `securityweb.HTTPTransport` interface directly.
- **PEP remains unique in executor**:
  - approval/policy denies happen before transport execution,
  - transport remains policy-agnostic I/O only.
- Approved `GET/HEAD` requests can execute through injected real transport (e.g. `securitywebhttp.Transport`) in `httptest`-only tests.
- Deny paths guarantee **transport is not invoked**:
  - invalid approval,
  - denied method (`POST`),
  - out-of-scope target.
- Redirect behavior preserved:
  - observed and recorded,
  - not followed by default.
- Evidence chain remains explicit:
  - `request_ref -> approval_id -> evidence_id` in `ExecutionResult`.
- `BodyTruncated` semantics aligned:
  - executor now preserves truncation when either preview cap truncates OR the underlying transport already marks `BodyTruncated=true`.


## 14) 7.9D Integration Hardening & Traceability Verify

- Added explicit invalid URL deny code: `request_url_invalid` (empty URL, parse error, unsupported scheme, empty host).
- Maintains **deny before transport** for invalid URL, missing approval, expired approval, out-of-scope, POST, missing limits, prohibited tool class.
- EvidenceID strategy hardened to include `request_ref + approval_id` to reduce collisions across repeated approvals.
- Traceability contract reinforced: `request_ref -> approval_id -> evidence_id -> execution_result`.
- Result metadata required in tests: `RequestID`, `ApprovalID`, `EvidenceID`, `Method`, `URL`, `StatusCode`, `RedirectObserved`, `BodyTruncated`, `RedactionsApplied`, `SafetyNotes`.
- Redirect handling hardening remains: `follow_redirects=false`, `redirect_location_invalid`, `redirect_out_of_scope`.
- Boundary unchanged: offline-only + httptest-only integration tests, no external network tests.

## 15) 7.10A Passive Headers Check MVP (implemented)

This phase adds a **pure passive header analysis module** in `internal/securityweb`:

- `check_headers.go`
- `check_headers_test.go`

Contract:
- input is `ExecutionResult` metadata only (no new request execution).
- output is `HeaderCheckResult` containing observations and conservative candidate findings.
- `evidence_id` is required for traceability and is propagated to each candidate finding.
- headers are evaluated case-insensitively.

Safety/boundary invariants preserved:
- no CLI/OpenCode execution changes.
- no crawler/scanner/fuzzing/payloads.
- no POST/auth/cookies.
- no redirect following changes.
- no runtime flow execution.
- no external network tests.

Behavior highlights:
- missing CSP is not auto-High (conservative low/medium posture).
- missing HSTS is only applicable on HTTPS.
- clickjacking finding is deduplicated when CSP `frame-ancestors` covers missing X-Frame-Options.
- explicit limitations are emitted when context is insufficient.
- HEAD-only posture can emit a limitation note: `GET fallback may be needed in a future approved request` (planning-only; no execution).

## 16) 7.10B Passive Discovery Findings MVP (implemented)

This phase normalizes a shared passive findings contract for future passive checks.

Shared model (internal/securityweb):
- `PassiveObservation`
- `CandidateFinding`
- `PassiveCheckResult`
- `ObservationStatus` now includes: `present`, `missing`, `weak`, `not_applicable`, `needs_context`
- `PassiveCheckID`, `FindingCategory`, `SeverityHint`, `ConfidenceHint`

Compatibility:
- `HeaderCheckResult` is now an alias of the shared result model:
  - `type HeaderCheckResult = PassiveCheckResult`
- `EvaluatePassiveHeaders(...) HeaderCheckResult` remains stable.

Traceability and calibration:
- observation and candidate finding include `EvidenceID` linkage.
- candidate findings include `RelatedObservationIDs` and `SourceCheckID`.
- candidate findings are not confirmed findings; severity/confidence are hints.

## 17) 7.11 Robots/Sitemap/Security.txt Passive Check MVP (implemented)

Added passive discovery-file checks in `internal/securityweb/check_discovery_files.go`:
- `EvaluatePassiveRobots(ExecutionResult)`
- `EvaluatePassiveSitemap(ExecutionResult, inScopeHosts []string)`
- `EvaluatePassiveSecurityTxt(ExecutionResult, now func() time.Time)`

Contract and boundaries:
- consume existing `ExecutionResult` only.
- return `PassiveCheckResult` only.
- no new requests, no URL following, no crawler/scanner behavior.
- no external validation of discovered links or contacts.

Traceability:
- all observations and candidate findings preserve `EvidenceID` linkage.
- candidate findings include `RelatedObservationIDs` and `SourceCheckID`.

Source check IDs:
- `passive_robots_mvp`
- `passive_sitemap_mvp`
- `passive_securitytxt_mvp`

## 18) 7.12 Surface Mapping from Passive Responses MVP (implemented)

Added passive surface mapping in `internal/securityweb/surface_map.go`:
- `BuildSurfaceMap(results []ExecutionResult, checks []PassiveCheckResult, scopeHosts []string) SurfaceMap`

Contract:
- consumes only existing passive evidence (`ExecutionResult` + `PassiveCheckResult`).
- produces an aggregated `SurfaceMap` with hosts/urls/paths/signals/candidate areas.
- deterministic map id for MVP: `surface-map-static-mvp`.

Boundaries:
- no network requests.
- no URL following.
- no crawling/scanning.
- no confirmed findings emitted by surface mapping.

## 19) 7.13 Risk Hypotheses + Expert Skill Routing from SurfaceMap (implemented)

Added non-executing hypothesis layer in `internal/securityweb/risk_hypotheses.go`:
- `BuildRiskHypothesesFromSurfaceMap(SurfaceMap) RiskHypothesisSet`

Contract:
- hypotheses are triage proposals, not confirmed findings.
- priority is triage order, not vulnerability severity.
- confidence is derived from passive evidence only.
- suggested checks are planning/dry-run only.

Safety boundaries preserved:
- no network requests,
- no active checks,
- no crawling/scanning/fuzzing,
- no runtime execution.

## 20) 7.14 Web Test Plan Builder from Risk Hypotheses (implemented)

Added planning-only test-plan layer in `internal/securityweb/test_plan.go`:
- `BuildWebTestPlanFromHypotheses(h RiskHypothesisSet, ctx *AssessmentSessionContext) WebTestPlan`

Contract:
- converts hypotheses into safe planning artifacts only.
- no requests are executed or prepared as request-ready by default.
- global execution mode defaults to `planning_only`.
- `ProposedMethod` remains empty by default in MVP.
- required approvals are future-facing and generated only for `blocked_needs_approval` items.

Safety boundaries preserved:
- no network requests,
- no active checks,
- no crawler/scanner/fuzzing,
- no confirmed findings.

## 21) 7.15 Controlled Check Plan / Request Plan Bridge (implemented)

Added controlled dry-run bridge in `internal/securityweb/check_plan_bridge.go`:
- `BuildControlledCheckPlanFromWebTestPlan(plan WebTestPlan, ctx AssessmentSessionContext, baseTarget string) ControlledCheckPlan`

MVP conversion map (planning-only):
- `headers-hardening-review-dry-run` -> `HEAD /`
- `robots-exposure-review-dry-run` -> `GET /robots.txt`
- `sitemap-exposure-review-dry-run` -> `GET /sitemap.xml`
- `securitytxt-governance-review-dry-run` -> `GET /.well-known/security.txt`

Safety/contract guarantees:
- execution mode fixed to `dry_run`
- `WouldExecute=false` always
- no executor invocation
- no transport invocation
- no active checks
- no POST/query/payload/crawler behavior
- non-convertible items emit `bridge_item_not_convertible_yet`
- policy decisions are retained even when dry-run denies execution

## 22) 7.16A Execute Approved Controlled Check against httptest-only (integration tests)

Added internal integration coverage (`internal/securityweb/integration_controlled_check_test.go`) to validate a controlled transition path:

`WebTestPlanItem -> PlannedRequest (bridge dry_run) -> explicit ExecutionApproval -> ExecuteApproved (executor) -> ExecutionResult -> PassiveCheckResult`

Scope and boundaries:
- httptest-only transport execution for integration scenarios.
- no external network tests.
- no CLI/OpenCode/runtime flow execution.
- no execute_approved auto-transition from planning; approval is explicit in tests only.

Key assertions:
- original `ControlledCheckPlan` remains `dry_run` and `WouldExecute=false`.
- executor path uses `execute_approved` only with explicit per-request approval.
- deny-before-transport holds for invalid/expired approval.
- traceability chain preserved: request_ref -> approval_id -> evidence_id -> passive_check_result.


## 16) 7.16B End-to-End Passive Pipeline over `httptest` (implemented)

Phase 7.16B validates the internal integration loop using real `securitywebhttp.Transport` against `httptest` only:

`ExecuteApproved -> passive checks -> SurfaceMap -> RiskHypothesisSet -> WebTestPlan`

Key enforced constraints in integration tests:
- `ControlledCheckPlan` remains `dry_run` + `WouldExecute=false`
- transition to `execute_approved` occurs only inside test with explicit per-request approval
- no external network, no CLI/OpenCode execution, no runtime flow execution
- GET/HEAD only, no redirects followed, no auth/cookies, no crawler/scanner
- traceability preserved: `request_ref -> approval_id -> evidence_id` and propagated into passive/surface/risk/plan artifacts
- final `WebTestPlan.ExecutionMode` remains `planning_only`
- generated outputs remain hypothesis/planning artifacts (no confirmed findings)


## 23) 7.16C Operational Execution Gate Design (design-only)

Design reference added:
- `docs/design/operational-execution-gate.md`

7.16C clarifies staged operational gating without enabling runtime behavior:
- Level 0 `planning_only`
- Level 1 `dry_run`
- Level 2 `httptest execute_approved`
- Level 3 `manual owned-target execute_approved` (future manual gate)
- Levels 4-5 future-only (not enabled)

Contract highlights:
- `execution_allowed=false by default`
- per-request approval required
- no generic approval
- exact manual token required for concrete requests: `APPROVE request_ref=... method=... url=...`
- any mismatch => deny-before-transport
- first owned-target run one request only (`HEAD /` preferred, `GET /robots.txt` secondary)

Boundary remains unchanged in 7.16C:
- design-only
- no runtime
- no live target execution
