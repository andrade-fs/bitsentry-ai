## Manual Execution Entrypoint Execute Design (Phase 7.18B, design-only)

### 1) Scope and status
- Manual Execution Entrypoint Execute Design
- manual-execute is not implemented in 7.18B
- 7.18B freezes the future contract only.
- No runtime execution is introduced in this phase.

### 2) Future command proposal (not implemented in 7.18B)
Future command surface:
- `bitsentry-ai securityweb manual-execute`

Required flags (future):
- `--request-ref`
- `--method`
- `--url`
- `--approval`
- `--scope-host`
- `--timeout-seconds`
- `--max-response-size-bytes`
- `--max-preview-size-bytes`
- `--request-budget`
- `--rate-limit-per-minute`
- `--stop-condition` (repeatable)
- `--confirm-execute` (mandatory)

Hard rule:
- `--confirm-execute required`

### 3) Safety and validation contract
- preflight deny stops before transport
- one request only
- HEAD only first MVP
- path / only first MVP
- request budget must be `1`
- exact approval token required and exact tuple match (`request_ref + method + url`)
- no redirects followed
- no auth/cookies
- no POST
- no payloads
- no crawler/scanner
- no background execution

### 4) Execution flow contract (future 7.18C)
1. Run `ManualPreflight` service.
2. If preflight deny => stop.
3. Build `AssessmentSessionContext`.
4. Build `PlannedRequest`.
5. Build `ExecutionApproval`.
6. Invoke `OfflineControlledExecutor.ExecuteApproved`.
7. Use `securitywebhttp.Transport`.
8. Produce `ExecutionResult`.
9. Run `EvaluatePassiveHeaders`.
10. Print human/JSON report.

### 5) Output contract (human + --json)
Required fields:
- `execution_backend_available`
- `entrypoint_available`
- `preflight_decision`
- `executed`
- `request_ref`
- `approval_id`
- `evidence_id`
- `method`
- `url`
- `status_code`
- `redirect_observed`
- `redirect_location`
- `headers_redacted`
- `body_truncated`
- `duration_ms`
- `redactions_applied`
- `safety_notes`
- `passive_header_check`
- `candidate_findings`
- `limitations`
- `pass_fail`

Denied output rules:
- `executed=false`
- `policy_decision=deny`
- include `violations`
- No claim execution unless ExecutionResult exists
- No fabricated evidence

### 6) Canonical result states (frozen for 7.18C)
- `PREFLIGHT_DENIED`
- `APPROVAL_DENIED`
- `EXECUTION_DENIED_PRE_TRANSPORT`
- `TRANSPORT_ERROR`
- `EXECUTED`
- `COMPLETED_WITH_CANDIDATES`
- `COMPLETED_NO_CANDIDATES`

### 7) Evidence and finding handling rules
- Candidate findings are not confirmed vulnerabilities.
- Do not run passive checks over denied empty execution results unless explicitly marked invalid/non-remote.
- No claim execution unless `ExecutionResult` exists.
- No fabricated evidence.

### 8) Future test plan (implementation phase)
Target phase:
- 7.18C httptest-only implementation

Required tests in 7.18C/7.18D pipeline:
- preflight deny does not call transport
- exact approval can execute only one HEAD /
- generic approval denied
- missing `--confirm-execute` denied
- missing `max_preview_size` denied
- non-HEAD denied
- non-root denied
- transport result produces `ExecutionResult`
- passive check runs only on executed result
- denied result does not create fake findings
- JSON output sets `executed=true` only after execution
- no external network in automated tests (httptest only)

### 9) Phase split recommendation
- 7.18B = execute design only (this document)
- 7.18C = manual-execute implementation with httptest-only tests
- 7.18D = optional owned-target manual execution

Anchor:
- 7.18D owned-target manual execution


### 10) 7.18C implementation update (completed / PASS)
- `manual-execute` is implemented in 7.18C.
- Historical note retained: 7.18B remained design-only and did not implement runtime command paths.
- Automated tests are `httptest`/fake-transport only.
- No external network is used in automated tests.
- No live target execution is part of 7.18C.
- No OpenCode runtime execution is part of 7.18C.
- `--confirm-execute` is mandatory.
- `executed=true` is emitted only when `ExecutionResult` exists.
- Deny paths do not run `PassiveHeaderCheck`.
- No fabricated evidence is allowed in deny or execute paths.
- `internal/securityweb` remains offline-safe and does not import `net/http`.
- 7.18D remains the future owned-target CLI gate phase.


### 11) Phase 7.18D.2 preflight scope-host alignment (manual-preflight, contract-only)
- `--scope-host` is contractual and required for `manual-preflight`.
- Missing scope host must deny with `missing_scope_host`.
- URL hostname mismatch against scope host must deny with `scope_host_mismatch`.
- Host matching semantics are defined as `url.Hostname()` with normalized lower+trim behavior before comparison.
- Boundary unchanged: preflight-only contract (`would_execute=false`), no transport/network invocation, no `ExecuteApproved`, no live execution.
