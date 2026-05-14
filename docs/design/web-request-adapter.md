## Controlled Web Request Adapter (Phase 7.7, contract-first)

### 1) Purpose / Non-goals

**Purpose**
- Define a canonical architecture contract for a future **Controlled Web Request Adapter** that enforces web-assessment safety policy before any request intent can move forward.
- Keep Phase 7.7 as **documentation + contractual anchors only**.

**Non-goals (Phase 7.7)**
- No Go runtime implementation.
- No stubs Go.
- No flow execution.
- No network requests.
- No crawler real.
- No curl/httpx/nuclei/browser execution.

### 2) Relationship to web-assessment lifecycle

This adapter is a future execution boundary aligned to the current declarative lifecycle:

`web-assessment-init -> web-assessment-scope -> web-assessment-recon-plan -> web-assessment-map -> web-assessment-test-plan -> web-assessment-requests -> web-assessment-findings -> web-assessment-report`

In 7.7 it only formalizes contract mapping; execution remains disabled by default.

### 3) Contract sources

- `assets/flows/web-assessment.yaml`
  - Defines gates/stage contract and route boundaries.
- `assets/skills/security/web-assessment-requests/SKILL.md`
  - Defines canonical operational policy (approval, scope, tool class, deny rules).
- Evidence/Report contracts
  - `assets/skills/security/web-assessment-findings/SKILL.md`
  - `assets/skills/security/web-assessment-report/SKILL.md`

**Source-of-truth statement**
- `web-assessment.yaml` defines gates/stage contract.
- `web-assessment-requests/SKILL.md` defines policy operativa canónica.
- `web-request-adapter.md` translates these contracts into future architecture.

### 4) Architecture

- **PolicyEvaluator**
  - Validates authorization/scope/tool class/mode constraints.
  - Produces `PolicyDecision` or `PolicyViolation`.
- **DryRunPlanner**
  - Converts approved intent into deterministic `PlannedRequest` artifacts without execution.
- **EvidenceRecorder**
  - Produces traceable evidence metadata and lifecycle chain continuity.
- **Redactor**
  - Ensures no secrets in logs/evidence outputs.
- **Executor (future, out of scope 7.7)**
  - Reserved for controlled operational phases only.

### 5) Conceptual API

- **ControlledWebRequestAdapter**
  - `Plan(ctx AssessmentSessionContext, req PlannedRequest) -> (PolicyDecision, []PolicyViolation)`
  - `DryRun(ctx AssessmentSessionContext, req PlannedRequest) -> (EvidenceEntry, []PolicyViolation)`
  - `ExecuteApproved(...)` reserved for future phases.

- **PlannedRequest**
  - target/url, method, headers policy view, expected evidence plan, stop-condition references.

- **AssessmentSessionContext**
  - authorization anchors, scope anchors, execution mode, rate profile, request budget, retest linkage.

- **PolicyDecision**
  - allow/deny + normalized rationale + violated controls list.

- **EvidenceEntry**
  - evidence ID, request intent metadata, redaction status, linkage to findings/report.

- **PolicyViolation**
  - machine-readable violation code + human-readable reason.

### 6) Execution modes

- **planning_only**: plan and validate only; no requests.
- **dry_run**: simulate contractual request path + evidence plan; no network.
- **execute_approved**: future gated execution mode requiring explicit per-request approval.
- **retest**: future mode preserving scope/authorization continuity and prior evidence linkage.

### 7) Enforcement points

- scope validation
- scheme validation
- redirect policy
- method policy
- rate limits
- request budget
- timeout
- max response size
- stop conditions
- redaction
- evidence IDs

Operational anchor rules (contractual, still non-operational in 7.7):
- in-scope target required (authorized target required)
- explicit approval before requests (explicit approval per request)
- GET/HEAD default
- no secrets in logs
- out-of-scope redirects denied

### 8) Error model

- `ErrMissingAuthorization`
- `ErrScopeViolation`
- `ErrOutOfScopeRedirect`
- `ErrMissingRateLimit`
- `ErrMissingStopConditions`
- `ErrToolClassNotAllowed`
- `ErrExecutionModeDenied`
- `ErrMissingEvidencePlan`

### 9) Evidence model + Markdown template

Minimum conceptual fields per evidence entry:
- `Evidence ID`
- `Session/Mode`
- `Authorization Ref`
- `Scope Ref`
- `Planned Request Ref`
- `Policy Decision`
- `Redaction Applied`
- `Linked Finding IDs`
- `Notes / Assumptions`

Template:

```md
## Evidence Entry
- Evidence ID: WEB-EV-0001
- Session/Mode: <session-id> / planning_only|dry_run|execute_approved|retest
- Authorization Ref: <auth-ref>
- Scope Ref: <scope-ref>
- Planned Request Ref: <request-ref>
- Policy Decision: allow|deny (+ reason)
- Redaction Applied: yes|no
- Linked Finding IDs: <list>
- Notes / Assumptions: <notes>
```

### 10) Layered test strategy

- **Layer 1 (current 7.7)**: static contract anchors in tests (no runtime/network).
- **Layer 2 (7.7B)**: offline Go stubs for conceptual API and policy decision wiring.
- **Layer 3 (7.8+)**: passive discovery and controlled request orchestration tests, still safety-gated.

### 11) Future phases

- 7.7B offline Go stubs
- 7.8 Passive Discovery MVP
- 7.9 Controlled Crawler MVP
- 7.10 Safe Check Modules

### 12) 7.7B implemented contracts mapping

This section maps the canonical 7.7 design into an **offline-only Go contract package**:

- Package: `internal/securityweb/`
- Runtime boundary: **sin network execution** (no net/http client, no crawler, no runtime flow execution)

Design component -> 7.7B type/contract mapping:

- `ControlledWebRequestAdapter` -> interface with only:
  - `Plan(...)`
  - `Validate(...)`
  - `RenderEvidenceTemplate(...)`
  - `RedactEvidence(...)`
- `PolicyEvaluator` -> `DefaultPolicyEvaluator` + `PolicyDecision` + `PolicyViolation`
- `DryRunPlanner` -> `DefaultDryRunPlanner` (deterministic plan shaping only)
- `EvidenceRecorder` -> `DefaultEvidenceRecorder` (template rendering + redaction flow)
- `Redactor` -> `DefaultRedactor` (headers/query-sensitive fields redaction)

Core offline domain types implemented:

- `AssessmentSessionContext`
- `PlannedRequest`
- `PolicyDecision`
- `PolicyViolation`
- `EvidenceEntry`
- `ExecutionMode`
- `Intensity`
- `ToolClass`
- `RequestMethod`

Policy rules implemented in `policy.go` (contract-only enforcement):

- `planning_only` never executes
- `dry_run` never executes
- `execute_approved` requires explicit approval
- `retest` requires existing finding/check linkage
- target must be in scope
- scheme must be `http`/`https`
- `GET`/`HEAD` allowed-by-default posture maintained via deny list for unsafe methods
- `POST` denied by default
- unsafe/destructive methods denied by default
- rate limit, request budget, timeout, max response size required for executable modes
- stop conditions required
- evidence plan required
- prohibited tool class denied
- out-of-scope redirects denied by default
- no secrets in evidence/logs through redaction contracts

Explicitly out of 7.7B scope:

- no `ExecuteApproved`
- no network execution
- no live target testing/tooling execution

### 13) 7.8A Passive Discovery Planner (offline-only)

7.8A introduces an offline `DiscoveryPlan` artifact in `internal/securityweb` for deterministic passive discovery planning only.

Implemented MVP request set (fixed order):
- `HEAD /`
- `GET /`
- `GET /robots.txt`
- `GET /sitemap.xml`
- `GET /.well-known/security.txt`

Explicitly excluded in 7.8A MVP:
- `/favicon.ico`
- `/.well-known/change-password`

7.8A safety/behavior rules:
- planner operates only as offline planning contract (no runtime execution)
- valid operating modes are `planning_only` and `dry_run` for this planner
- `would_execute = false` always
- GET/HEAD only, no payloads
- no dynamic query params, no fuzzing, no wordlists, no crawling
- no automatic redirect following
- every planned request is policy-validated
- every planned request receives an evidence template entry with stable reference IDs

### 14) 7.8B Controlled HTTP Executor design reference

Phase 7.8B adds a dedicated design-only contract document:

- `docs/design/controlled-http-executor.md`

7.8B is intentionally limited to docs + lightweight contractual anchors and does **not** introduce runtime executor behavior, `net/http`, or real network execution.

### 15) 7.8C Offline Controlled Executor implementation note

7.8C introduces an offline executable component in `internal/securityweb`:

- `OfflineControlledExecutor` (separate from planner/adapter)
- `ExecutionApproval` with `expires_at` authoritative and `ttl_seconds` as metadata/audit
- `FakeTransport` deterministic lookup (primary `request_ref`, fallback `method+url` only if `request_ref` is empty)
- additive violation-code prefixes: `approval_*`, `redirect_*`, `limiter_*`
- `ExecutionResult` metadata contract including `evidence_id`, `body_truncated`, `body_preview_redacted`, `response_size`, `max_preview_size`

Boundary remains strict: no real network, no `net/http`, no runtime flow execution.

### 16) 7.9A boundary note (future real transport)

- The executor boundary remains policy-first and offline-safe in `internal/securityweb`.
- Future real transport is explicitly separated into `internal/securitywebhttp`.
- Policy Enforcement Point remains unique: `OfflineControlledExecutor is the single Policy Enforcement Point`.

### 17) 7.9B minimal real transport note

- `internal/securitywebhttp` contains the minimal real HTTP transport skeleton for request I/O.
- PEP remains unique in internal/securityweb.
- transport remains policy-agnostic (no approval/scope/tool/method/rate/budget decisions).
- tests remain `httptest`-only with no external network usage.
- Real transport must execute only policy-approved requests and must not decide policy.


- See also: `docs/design/controlled-http-executor.md` for the dedicated execution-layer design and 7.8D hardening constraints.

### 18) 7.10A Passive Headers Check MVP mapping

A first passive check module is now modeled in `internal/securityweb` as a pure post-execution analysis contract:

- Input: `ExecutionResult` / response metadata.
- Output: `HeaderCheckResult` with:
  - `observations` (`present | missing | weak | not_applicable`)
  - conservative `candidate findings`
  - explicit `limitations`
  - evidence linkage via `evidence_id`

The module does not execute requests and does not alter approval/policy transport flow.
It consumes already approved/executed evidence only.

### 19) 7.10B Passive findings normalization mapping

Passive post-execution checks now target a common result pipeline:

`ExecutionResult -> PassiveObservation[] -> CandidateFinding[] -> PassiveCheckResult`

This enables future checks (robots/sitemap/security.txt/TLS/redirect/cookies/forms) to share the same evidence-linked and dedup-friendly contract without adding network execution.

### 20) 7.11 Passive discovery files mapping

The passive post-execution pipeline now includes discovery-file analyzers for:
- robots.txt
- sitemap.xml
- /.well-known/security.txt

These analyzers are pure evidence consumers (`ExecutionResult -> PassiveCheckResult`) and keep execution boundaries unchanged (no new requests, no runtime execution).

### 21) 7.12 Surface map aggregation mapping

Passive evidence now supports a non-executing aggregation layer:

`ExecutionResult + PassiveCheckResult[] -> SurfaceMap`

This layer captures observed surface topology and candidate areas for future hypothesis stages while preserving strict no-network boundaries.

### 22) 7.13 Risk hypotheses mapping

A passive-only hypothesis layer now maps:

`SurfaceMap -> RiskHypothesisSet`

This layer routes candidate areas to suggested expert skills and dry-run next checks without executing requests or confirming vulnerabilities.

### 23) 7.14 Web test plan mapping

Passive planning pipeline now includes:

`RiskHypothesisSet -> WebTestPlan`

This step prepares non-executing, safety-gated test plan items with traceability and explicit scope/approval blockers.
