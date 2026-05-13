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
