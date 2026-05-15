## Title
Phase 7.18D.4 — CLI Execution Evidence Normalization

## Executive Summary
This document records the first CLI manual-execute owned-target run as a documentation evidence normalization artifact. The run outcome is recorded as executed=true with a completed result state, and PASS is based on traceable CLI evidence only.

## Command Context
- command family: `bitsentry-ai securityweb manual-execute`
- target class: owned target
- request type: `HEAD` root request (`request_ref=first-owned-head-root`)
- phase boundary: documentation/evidence normalization only

## Approval / Scope
- approval_id: `approval-first-owned-head-root`
- request_ref: `first-owned-head-root`
- evidence_id: `WEB-EV-first-owned-head-root-approval-first-owned-head-root`
- scope model: single approved owned-target HEAD root request
- no new requests in 7.18D.4

## ExecutionResult
- executed: `true`
- state: `COMPLETED_WITH_CANDIDATES`
- status_code: `200`
- redirect_observed: `false`

## PassiveHeaderCheck Summary
PassiveHeaderCheck is included only after a real ExecutionResult. In this evidence chain, PassiveHeaderCheck follows the recorded execution result and is not produced for deny/non-executed paths.

## Candidate Findings
- `hf-referrer-weak` is recorded as a candidate finding.
- candidate finding, not confirmed vulnerability.

## Safety Controls
- no fabricated evidence
- PassiveHeaderCheck only after real ExecutionResult
- evidence_id preserved; evidence traceability preserved through request/approval/evidence identifiers
- no new requests in 7.18D.4

## Limitations
The following areas were not evaluated in this phase artifact:
- body content
- robots
- sitemap
- security.txt
- crawl coverage
- authenticated areas

## PASS/FAIL
- pass_fail: `PASS`
- status: PASS WITH NOTES (documentation normalization context)

## Traceability
- request_ref: `first-owned-head-root`
- approval_id: `approval-first-owned-head-root`
- evidence_id: `WEB-EV-first-owned-head-root-approval-first-owned-head-root`
- execution state chain: `executed=true -> state=COMPLETED_WITH_CANDIDATES -> status_code=200`

## Next Steps
- Keep this artifact as the normalized CLI evidence record for 7.18D.4.
- If scope is expanded later, require a new approval token and separate evidence chain; do not mutate this historical record.
