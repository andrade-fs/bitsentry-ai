## Phase 7.17E — Close First Owned Target Run / Lessons Learned

## 1) Executive closure
- first owned-target run closed
- no new request in 7.17E
- The first controlled cycle is formally closed with one successful real request and one earlier blocked pre-transport attempt.
- Validated circuit: `Approval -> ExecutionResult -> PassiveHeaderCheck -> Normalization -> Report`.

## 2) What worked
- Exact approval gate worked as intended.
- generic approval inputs were rejected until exact token alignment.
- Final successful run executed exactly one request:
  - `HEAD https://bitsentry.xyz/`
- Guardrails held during successful run:
  - one request only
  - no redirects followed
  - no POST
  - no GET fallback
  - no robots
  - no crawler/scanner

## 3) What failed or almost failed
- Earlier attempt was correctly denied before transport:
  - `missing_max_preview_size blocked pre-transport`
- This prevented unintended network execution with incomplete required limits.

## 4) Guardrails validated
- no fabricated evidence
- No remote evidence was claimed when pre-transport deny occurred.
- Exact approval matching remained mandatory.
- Scope stayed constrained to one approved request.

## 5) Evidence produced
- request_ref: `first-owned-head-root`
- approval_id: `approval-first-owned-head-root-001`
- evidence_id: `WEB-EV-first-owned-head-root-approval-first-owned-head-root-001`
- status_code: `200`
- redirect_observed: `false`
- policy_decision: `allow`

## 6) Candidate findings produced
- `hf-referrer-weak`
- candidate finding not confirmed
- This remains a hardening signal, not a confirmed vulnerability.

## 7) What is NOT proven
- HEAD-only evidence
- no body fetched
- no robots/sitemap/security.txt
- no crawl
- no authenticated areas
- no external assets tested
- no complete application coverage

## 8) Operational lessons
- exact approval UX must be preserved
- entrypoint formalization required
- avoid ad-hoc/local temporary execution runners for production-owned runs
- required limits (including max_preview_size) must be validated pre-transport every time

## 9) Required fixes before expanding scope
- Create a formal manual execution entrypoint with explicit preflight validation.
- Enforce required limit completeness before any transport call.
- Keep strict separation between candidate findings and confirmed findings unless additional approved evidence is collected.
- Maintain one-request gating for initial expansions.

## 10) Recommended next phase
- Phase 7.18A — Manual Execution Entrypoint Formalization
