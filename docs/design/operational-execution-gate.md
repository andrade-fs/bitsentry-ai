## Operational Execution Gate (Phase 7.16C, design-only)

### 1) Purpose / Non-goals

**Purpose**
- Define the gate design required before any future owned-target real request execution is allowed.
- Preserve the current policy model where `OfflineControlledExecutor` remains the single Policy Enforcement Point (PEP).
- Keep this phase contract-first and documentation-first.

**Non-goals (7.16C)**
- No real requests.
- No live target execution.
- No CLI/OpenCode execution path changes.
- No runtime flow execution.
- No crawler/scanner/fuzzing/payload/auth/cookies/POST/redirect-follow behavior.

### 2) Operational Gate Levels

- **Level 0: planning_only**
  - Planning/validation artifacts only.
  - No execution intent.

- **Level 1: dry_run request plan**
  - Request-plan artifacts and policy/evidence templates only.
  - No network execution.

- **Level 2: httptest execute_approved**
  - Synthetic integration execution boundary for tests only.
  - Explicit per-request approval required.

- **Level 3: manual owned-target execute_approved**
  - Future manual gate for a first owned-target real request.
  - Explicit human approval and strict limits required.

- **Level 4: controlled multi-request passive discovery**
  - Future only.
  - Not enabled in 7.16C.

- **Level 5: active checks**
  - Future only.
  - Not enabled in 7.16C.

**7.16C scope note**
- This phase defines design up to Level 3 only.
- It does not implement operational execution.

### 3) Assessment Session Context (authorization context only)

Required fields for operational design context:
- `session_id`
- `target_owner_declared`
- `scope_hosts`
- `in_scope_urls`
- `out_of_scope`
- `environment`
- `intensity`
- `allowed_methods`
- `allowed_paths`
- `prohibited_actions`
- `rate_limit`
- `request_budget`
- `timeout`
- `max_response_size`
- `max_preview_size`
- `stop_conditions`
- `evidence_policy`
- `redaction_policy`
- `execution_allowed=false by default`

Mandatory rule:
- A session context never authorizes execution by itself.
- Real execution requires explicit per-request approval.

### 4) Operational ExecutionApproval (per-request only)

Required fields:
- `approval_id`
- `session_id`
- `request_ref`
- `method`
- `url`
- `scope_ref`
- `approved_by`
- `approved_at`
- `expires_at`
- `approval_text_or_hash`
- `max_requests`
- `max_duration`
- `rate_limit`
- `stop_conditions`
- `expected_evidence`
- `dry_run_reference`

Rules:
- **per-request approval required**
- **no generic approval**
- request authorization must match exact tuple: `request_ref + method + url`
- mismatch or expiry => deny-before-transport

### 5) manual boundary (future Level 3 gate)

Before any owned-target execution, present and confirm:
- exact request plan
- exact method and URL
- purpose
- risk
- expected evidence
- effective limits

Mandatory confirmation token:
- `APPROVE request_ref=... method=... url=...`

Rules:
- `APPROVE request_ref=... method=... url=...` is required for concrete request execution.
- Ambiguous confirmations such as “adelante”, “ok”, “sí”, “hazlo” are not sufficient when one or more concrete requests are pending.
- Any approval/request mismatch must deny-before-transport.

### 6) First future owned-target run profile (manual only)

First future real run is constrained to:
- first owned-target run one request only
- `HEAD /` preferred
- HEAD / preferred
- `GET /robots.txt` secondary
- GET /robots.txt secondary
- target ownership declared
- in-scope required
- no redirects followed
- no auth/cookies
- no POST
- no crawler/scanner
- timeout/max response/max preview active
- evidence redacted

### 7) Failure / stop rules

Execution must stop on:
- unexpected 5xx
- timeout
- response too large
- redirect out of scope
- sensitive data detected
- user stop
- policy mismatch
- request budget exhausted

### 8) Audit trail contract

Minimum audit trail fields:
- `session_id`
- `request_ref`
- `approval_id`
- `evidence_id`
- `result_id` (if modeled)
- timestamp
- policy decision
- redactions applied
- safety notes
- limitations

### 9) Non-goals (operational safety)

- no autonomous pentest
- no mass scanning
- no DoS/load testing
- no brute force
- no credential attacks
- no exploit chains
- no destructive checks
- no background execution
- no secrets exposure

### 10) Status for 7.16C

- Completed / PASS WITH NOTES
- design-only
- no runtime
- no live target execution
- operational run pending future phase


## 11) 7.16D Manual Owned Target Run Protocol (design-only)

Design reference added:
- `docs/design/manual-owned-target-run-protocol.md`

This protocol defines the manual pre-run checklist, exact approval token validation, PASS/FAIL criteria, and evidence/audit expectations for a first future owned-target run without enabling runtime behavior in 7.16D.
