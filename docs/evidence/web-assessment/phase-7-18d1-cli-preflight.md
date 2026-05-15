# 7.18D.1 CLI Preflight Evidence

## Summary
Preflight-only validation for **7.18D.1** was performed.

## Command Path Resolution
- `./bitsentry-ai` with `securityweb` was not available (unknown command).
- Validation proceeded with: `rtk go run ./cmd/bitsentry-ai`.

## Observed Outcomes
- First failure (historical 7.18D.1): `unknown flag: --scope-host`
- Successful run achieved after removing `--scope-host`

Observed JSON fields from successful preflight:
- `execution_backend_available=true`
- `entrypoint_available=true`
- `approval_valid=true`
- `limits_complete=true`
- `would_execute=false`
- `policy_decision=allow`
- `violations=[]`
- `exact_approval_required=true`

## Execution Safety Confirmation
No request was executed, and no transport layer was invoked.

## Guardrails Confirmed
- No `manual-execute`
- No `--confirm-execute`
- No `ExecuteApproved`
- No live request

## Verdict
**PASS WITH NOTES**

Historical note (7.18D.1): `--scope-host` was not yet supported in `manual-preflight`.

## 7.18D.2 Alignment Note
- The 7.18D.1 gap is CLOSED by 7.18D.2 documentation+contract alignment.
- `--scope-host` is now contractual and required for `manual-preflight`.
- Missing scope host now denies with `missing_scope_host`.
- URL hostname mismatch vs scope host now denies with `scope_host_mismatch`.
- Matching behavior is documented as `url.Hostname()` with normalized lower+trim semantics.
- Boundary remains preflight-only: `would_execute=false`, no transport/network/`ExecuteApproved`, no live execution claims.
