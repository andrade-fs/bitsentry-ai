# Final Verify — Phase 3.5 Capability Service Extraction

## Why this service exists

Phase 3.4 identified a coupling risk in TUI apply flow: TUI directly managed executable invocation details.
Phase 3.5 introduces a small shared capabilities service to centralize reusable capability operations and reduce duplication.

## What logic is centralized

- Load current capability selection from config
- Save capability selection draft to config
- Validate draft selection against modeled registry
- Build plan and OpenCode projection from draft
- Apply dry-run / apply entrypoints (with report path retrieval)
- Read latest report path

## Subprocess usage status

- **Removed from TUI implementation details** (TUI now calls service methods)
- **Still present inside service apply/apply-dry-run**, which delegates to existing CLI apply path for safety and behavior preservation

This is intentional in Phase 3.5 to preserve OpenCode Phase 2.5 apply safeguards and avoid risky rewrite.

## Behavior preservation

- CLI commands remain unchanged.
- TUI still blocks apply/dry-run on unsaved draft.
- Real apply still requires explicit confirmation.
- Validation still blocks real apply.
- Existing report behavior remains discoverable via latest report path.

## Remaining risks

- Service apply still depends on executable path availability.
- Future phase can replace internal subprocess delegation with in-process apply core extraction.

## Verdict

PASS WITH NOTES
