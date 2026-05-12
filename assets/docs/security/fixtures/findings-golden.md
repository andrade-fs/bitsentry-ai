# Findings Golden Fixture (Safe / Synthetic)

## Finding 1
- ID: SEC-GOLD-001
- Title: Example emergency admin path bypasses tenant authorization
- Severity: Critical
- Confidence: High
- Category: Authorization
- Affected files: `internal/example/admin/emergency_handler.go`
- Affected component: `example-admin-emergency-handler`
- Evidence:
  - direct: conditional path lacks tenant boundary verification before privileged action
  - inferred: operation may impact all tenant-scoped records
- Impact: Cross-tenant privileged operation possible in synthetic scenario
- Likelihood: High (path available to authenticated operators)
- Remediation: Enforce tenant ownership check and privileged-role policy gate
- Verification: Add tests validating tenant isolation and forbidden response paths
- References / Notes: Synthetic fixture; no runtime exploit, no live targets

## Finding 2
- ID: SEC-GOLD-002
- Title: Demo JWT refresh flow accepts stale session marker
- Severity: High
- Confidence: Medium
- Category: Session Management
- Affected files: `internal/example/auth/refresh_handler.go`
- Affected component: `example-auth-refresh`
- Evidence:
  - direct: session freshness marker not validated on refresh branch
  - inferred: replay window may increase for compromised refresh tokens
- Impact: Session replay risk in synthetic auth lifecycle
- Likelihood: Medium (depends on token exposure conditions)
- Remediation: Enforce freshness marker validation + token rotation guard
- Verification: Add replay-focused unit tests for stale marker rejection
- References / Notes: Source-only observation, exploitability depends on deployment controls

## Finding 3
- ID: SEC-GOLD-003
- Title: Sample callback validator permits internal domain wildcard
- Severity: Medium
- Confidence: Medium
- Category: Server-Side Request Forgery
- Affected files: `internal/example/callback/validator.go`
- Affected component: `example-callback-validator`
- Evidence:
  - direct: suffix match logic allows broad internal domain acceptance
  - inferred: potential internal service reachability in permissive networks
- Impact: Internal request pivot opportunity in misconfigured environments
- Likelihood: Medium (requires network path and callable endpoint)
- Remediation: Strict host allow-list + DNS/IP resolution policy checks
- Verification: Add negative tests for private ranges and wildcard-adjacent domains
- References / Notes: No external testing performed

## Finding 4
- ID: SEC-GOLD-004
- Title: Demo upload path trusts extension without MIME cross-check
- Severity: Low
- Confidence: High
- Category: File Upload
- Affected files: `internal/example/upload/upload_handler.go`
- Affected component: `example-upload-handler`
- Evidence:
  - direct: extension validation exists but MIME signature check absent
  - inferred: malformed content may enter downstream processing
- Impact: Limited risk in synthetic upload flow due to isolated storage
- Likelihood: Low (further controls reduce practical exploitability)
- Remediation: Add MIME signature validation and size/type policy checks
- Verification: Unit tests for mismatched extension vs MIME signatures
- References / Notes: Bounded by synthetic storage controls

## Finding 5
- ID: SEC-GOLD-005
- Title: Example debug error shape discloses stack-trace metadata in non-prod profile
- Severity: Informational
- Confidence: Low
- Category: Informational
- Affected files: `internal/example/http/error_renderer.go`
- Affected component: `example-error-renderer`
- Evidence:
  - direct: verbose debug branch includes internal frame metadata in response object
  - inferred: operational visibility may increase during troubleshooting
- Impact: Informational leakage pattern awareness for hardening backlog
- Likelihood: Low (profile gated, non-production only)
- Remediation: Keep verbose details server-side and return generic client message
- Verification: Test profile-gated error response contract
- References / Notes: No sensitive values included in fixture

## Deduplication Example
- Duplicate candidates for missing tenant authorization in `example-admin-emergency-handler` are merged under `SEC-GOLD-001`.

## Evidence Grouping Example
- Evidence is grouped per finding ID and affected component with `direct` and `inferred` labels.

## Assumptions / Limitations
- Assumption: synthetic middleware chain is enabled as documented.
- Limitation: no runtime environment, infra policy, or live endpoint testing was performed.
