# Title
Source Security Review Report — Synthetic Demo Application

## Executive Summary
This report summarizes source-level security observations using synthetic scenarios only. No runtime execution, live target testing, or exploit execution was performed.

## Scope
- Repository snapshot review for demo handlers/configs
- Static code and config inspection
- Contract-aligned findings normalization

## Methodology
- Read-only first review
- Findings normalization via official taxonomy
- Severity calibration using Impact × Likelihood
- Confidence calibration by evidence quality (`direct` vs `inferred`)

## Repository / Application Context
- Demo multi-component backend with sample handlers and validators
- Review limited to source artifacts included in repository

## Risk Summary
- High: 1 (Authorization)
- Medium: 1 (Server-Side Request Forgery)
- Confidence distribution: High (1), Medium (1), Low (0)

## Findings
- SEC-EX-001 — Authorization — High / High
- SEC-EX-002 — Server-Side Request Forgery — Medium / Medium

## Evidence
- Grouped by finding ID and component
- Evidence labels preserved: `direct`, `inferred`

## Remediation Plan
1. Enforce role/ownership checks for export endpoints.
2. Restrict callback target validation to strict approved hosts and resolved IP policy.

## Verification Steps
1. Add unit tests for forbidden paths and negative callback validations.
2. Confirm no regression in valid request paths.

## Assumptions and Limitations
- Source-only assessment; no infrastructure/runtime validation.
- Network segmentation assumptions may affect exploitability likelihood.

## Next Steps
- Convert high/medium findings into remediation tracking issues.
- Re-run source review after mitigations are merged.

## Appendix
- This document uses synthetic examples and does not include secrets, sensitive data, or live targets.
