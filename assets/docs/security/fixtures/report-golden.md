# Title
Source Security Review Report — Golden Fixture (Synthetic)

## Executive Summary
This golden fixture demonstrates the required section order and safe-content reporting contract for source security review outputs.

## Scope
- Static review of synthetic demo components
- No runtime execution or live target interaction

## Methodology
- Read-only source and config inspection
- Taxonomy-driven findings normalization
- Severity by Impact × Likelihood; confidence by evidence quality

## Repository / Application Context
- Example modular backend with synthetic handlers and validators
- Review bounded to repository-visible artifacts only

## Risk Summary
- Critical: 1
- High: 1
- Medium: 1
- Low: 1
- Informational: 1
- Categories represented: Authorization, Session Management, Server-Side Request Forgery, File Upload, Informational

## Findings
- SEC-GOLD-001 — Authorization — Critical / High
- SEC-GOLD-002 — Session Management — High / Medium
- SEC-GOLD-003 — Server-Side Request Forgery — Medium / Medium
- SEC-GOLD-004 — File Upload — Low / High
- SEC-GOLD-005 — Informational — Informational / Low

## Evidence
- Evidence grouped by canonical finding ID and affected component
- Evidence labels preserved (`direct`, `inferred`)

## Remediation Plan
1. Prioritize critical/high authorization and session controls.
2. Harden callback validation and upload validation controls.
3. Normalize error rendering across profiles.

## Verification Steps
1. Add unit tests per remediation acceptance criteria.
2. Re-run source security review fixture checks.
3. Confirm no regression in approved request paths.

## Assumptions and Limitations
- Source-only review: no runtime, infra, or live endpoint validation.
- Confidence reflects available static evidence only.

## Next Steps
- Track remediation by owner and due date.
- Reassess severity/confidence after code changes.

## Appendix
- Safety note: synthetic examples only; no secrets, sensitive data, or exploitable payload instructions.
