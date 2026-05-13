# Title
Web Assessment Report Example — Synthetic Target

## Authorization Summary
- Authorization reference: AUTH-EXAMPLE-001
- In scope confirmation: approved for synthetic staging target only
- Testing window: 2026-05-10T10:00Z to 2026-05-10T12:00Z

## Scope
- https://staging.example.invalid
- Paths: /login, /account, /api/profile

## Out of Scope
- Production hosts
- Third-party SaaS endpoints

## Methodology
- Contract-first, declarative review of authorized request/evidence records

## Tooling and Intensity
- Tool class: Passive inspection, Single-request verification
- Intensity: low

## Request / Evidence Log
- Evidence ID: web-assessment-evidence-001
- Linked finding IDs: WA-F-001

## Risk Summary
- 1 High, 0 Critical, residual risk accepted pending remediation window

## Findings
- ID: WA-F-001
- Title: Missing secure cookie flags on session cookie
- Severity: High
- Confidence: Medium

## Remediation Plan
- Enforce `Secure` and `HttpOnly` flags for session cookies

## Verification Steps
- Re-check cookie response headers after deployment

## Assumptions and Limitations
- No runtime flow execution
- No target testing vivo

## Next Steps
- Track remediation in backlog and schedule re-validation

## Appendix
- Traceability: authorization → scope → request/evidence → finding → report
