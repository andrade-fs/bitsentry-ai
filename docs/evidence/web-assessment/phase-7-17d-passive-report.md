## Phase 7.17D — First Owned Target Passive Report Draft

## 1) Title
First Owned Target Passive Report Draft (Single HEAD Request)

## 2) Executive Summary
A single manually approved passive request (`HEAD https://bitsentry.xyz/`) was executed in prior phase 7.17B and then normalized in 7.17C. This draft reports only that existing evidence. One candidate hardening finding was identified (`hf-referrer-weak`) and remains a candidate finding, not confirmed vulnerability.

## 3) Scope
- one approved HEAD request to `https://bitsentry.xyz/`
- request_ref: `first-owned-head-root`
- method: `HEAD`
- no new requests in this phase

## 4) Methodology
- passive reporting only
- source artifacts only:
  - `docs/evidence/web-assessment/phase-7-17b-first-owned-head.md`
  - `docs/evidence/web-assessment/phase-7-17c-normalized-pipeline.md`
- no runtime execution in 7.17D

## 5) Evidence Summary
- status_code: `200`
- redirect_observed: `false`
- approval_id: `approval-first-owned-head-root-001`
- evidence_id preserved: `WEB-EV-first-owned-head-root-approval-first-owned-head-root-001`
- policy_decision: `allow`

## 6) Passive Header Review
Observed from HEAD response evidence:
- `Content-Security-Policy`: present
- `Strict-Transport-Security`: present
- `X-Content-Type-Options`: present (`nosniff`)
- `X-Frame-Options`: present (`SAMEORIGIN`)
- `Referrer-Policy`: `origin-when-cross-origin` (weaker than stricter alternatives)

## 7) Candidate Findings
- ID: `hf-referrer-weak`
- Type: candidate finding, not confirmed vulnerability
- Category: Configuration / hardening
- SeverityHint: `Low`
- ConfidenceHint: `High`
- Notes: observed policy may disclose more referrer context than stricter options depending on site requirements.

## 8) Recommendations
- Consider stricter Referrer-Policy if compatible with product behavior:
  - `strict-origin-when-cross-origin`
  - `same-origin`
  - `no-referrer`
- Validate functional impact before changing policy.
- Keep existing positive controls in place:
  - CSP
  - HSTS
  - X-Content-Type-Options
  - X-Frame-Options

## 9) Limitations
- HEAD-only evidence
- no body fetched
- no robots.txt checked
- no sitemap.xml checked
- no security.txt checked
- no crawling
- no authenticated areas
- no external assets tested
- no confirmed vulnerabilities
- candidate finding only

## 10) Next Steps
- Conservative option: close as passive evidence report.
- Approved expansion option: plan a separate one-request `GET /robots.txt` gate with a new exact approval token.
- Improvement option: perform focused hardening review for Referrer-Policy selection.

## 11) Appendix / Traceability
- request_ref: `first-owned-head-root`
- approval_id: `approval-first-owned-head-root-001`
- evidence_id: `WEB-EV-first-owned-head-root-approval-first-owned-head-root-001`
- traceability chain preserved: `request_ref -> approval_id -> evidence_id`
- WebTestPlan remains planning_only
- no active checks
- no new requests

## 12) Cross-reference (added in 7.18D.4)
For CLI manual-execute evidence normalization of the same owned-target request chain, see:
- `docs/evidence/web-assessment/phase-7-18d3b-cli-manual-execute-head.md`

Historical facts in this 7.17D report remain unchanged.
