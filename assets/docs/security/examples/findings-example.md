# Findings Example (Source Security Review)

This example demonstrates the **minimum finding contract** and safe documentation style for `source-security-review`.

## Minimum Finding Contract (Exact Tokens)
- ID
- Title
- Severity
- Confidence
- Category
- Affected files
- Affected component
- Evidence
- Impact
- Likelihood
- Remediation
- Verification
- References / Notes

## Allowed Enums
Severity:
- Critical
- High
- Medium
- Low
- Informational

Confidence:
- High
- Medium
- Low

## Example Findings (Synthetic / Safe)

### Finding A
- ID: SEC-EX-001
- Title: Demo admin route missing ownership check
- Severity: High
- Confidence: High
- Category: Authorization
- Affected files: `internal/demo/handlers/admin_handler.go`
- Affected component: `demo-admin-handler`
- Evidence:
  - direct: conditional allows authenticated users without role check on `/admin/export`
  - inferred: exported dataset can include other tenant metadata
- Impact: Unauthorized data export in demo environment pattern
- Likelihood: Medium (requires authenticated session but no privileged role)
- Remediation: Add explicit role/ownership check before export branch
- Verification: Unit test non-admin actor receives forbidden response
- References / Notes: Synthetic scenario, no live target, no exploit execution

### Finding B
- ID: SEC-EX-002
- Title: Sample callback validation allows broad internal hostnames
- Severity: Medium
- Confidence: Medium
- Category: Server-Side Request Forgery
- Affected files: `internal/sample/callback/validator.go`
- Affected component: `sample-callback-validator`
- Evidence:
  - direct: hostname suffix allow-list accepts broad internal namespace
  - inferred: service could reach internal metadata-style endpoints in some deployments
- Impact: Internal request pivot risk in misconfigured environments
- Likelihood: Medium (depends on deployment/network segmentation)
- Remediation: Enforce strict host allow-list + resolved IP policy
- Verification: Add negative tests for private ranges and non-approved domains
- References / Notes: Example only; no external target testing

## Deduplication Rule Example
- Candidate C1 and C2 reference the same root cause (missing ownership check in `demo-admin-handler`).
- Canonical output keeps only `SEC-EX-001` and attaches both evidence pointers.

## Evidence Grouping Example
- Group by finding ID and affected component.
- Keep `direct` evidence lines separate from `inferred` reasoning notes.

## Assumptions / Limitations Example
- Assumption: authentication middleware is enabled in all environments.
- Limitation: runtime network policy was not evaluated (source-only review).
