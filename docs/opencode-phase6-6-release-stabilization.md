# OpenCode Phase 6.6 — Release Stabilization

Status: **PASS WITH NOTES**

## Objective

Close Phase 6 with release-grade stabilization focused on documentation alignment, QA evidence consolidation, demo readiness, and release-candidate criteria for OpenCode-first usage.

## Scope

- Consolidate Phase 6 guardrails and closure evidence into one master release document.
- Align README quickstart + phase summary language with OpenCode-first UX.
- Mark roadmap progress through 6.6 and set next step to release candidate/tag preparation.
- Cross-link QA artifacts for route preview, MCP readiness, and MCP config preview.
- Define a clear end-to-end demo flow for OpenCode users (TUI-first).

## Non-Scope

- No major features.
- No architecture refactor.
- No runtime/autonomous execution engine activation.
- No credential/token/secret automation.
- No mutation of MCP credentials.

## Go / No-Go Checklist

- [x] Master closure doc created for Phase 6.6.
- [x] README updated with OpenCode-first quickstart and Phase 6.6 summary.
- [x] Roadmap updated to reflect 6.1–6.6 statuses and release-candidate next step.
- [x] QA docs cross-linked to Phase 6.6 closure.
- [x] Guardrails explicitly documented and consistent.
- [x] Demo flow documented for release validation.
- [x] Final test pass captured (`go test ./...`).

## QA Evidence

- Automated test command:
  - `go test ./...`
- Linked QA artifacts:
  - `docs/opencode-phase6-route-preview-qa.md`
  - `docs/opencode-phase6-5-mcp-config-preview-qa.md`
  - `docs/opencode-phase5-manual-qa.md` (historical regression baseline)

## Verified Guardrails

The following guardrails remain explicit and unchanged for the release candidate:

1. No `.env` access.
2. No secret/token/credential reads or writes.
3. No hardcoded credentials/API keys/passwords.
4. No MCP credential mutation.
5. `agent.bitsentry.permission.edit=deny` contract preserved.
6. CLI is debug/plumbing only, not primary UX.
7. OpenCode-only scope (no new agent targets).
8. No autonomous runtime activation.
9. No flow execution in preview/installer path.
10. No skill execution in preview/installer path.
11. MCP config preview is read-only/no-write with explicit future confirmation+backup gates.

## Demo Flow (Release Candidate)

1. Launch TUI: `bitsentry-ai`.
2. Open **Install / Setup**.
3. Choose **Install Everything** (or update/reinstall when already installed).
4. Validate **MCP Readiness** summary in review step.
5. Validate **MCP Config Preview** section is marked **PREVIEW ONLY** and shows no-write contract fields.
6. Complete install/update and verify done/control-panel summary:
   - OpenCode detection/config root
   - Bitsentry pack status
   - Native integration status
   - Manual notes and backup references
7. In OpenCode chat, run a route+capability preview prompt (example):
   - `@bitsentry Quiero mejorar el wizard del TUI`
8. Confirm visible route/capability preview is shown before deep planning and respects non-mutation gates.

## Release Candidate Criteria

- All Phase 6 QA docs are consistent and cross-referenced.
- README and roadmap communicate the same OpenCode-first boundaries.
- Guardrails are explicit and non-contradictory across docs.
- Demo flow is reproducible without CLI-first guidance.
- Automated tests pass on current branch.

## Pending Risks

- Manual OpenCode behavioral checks still depend on model/runtime behavior at execution time; docs reduce ambiguity but cannot fully eliminate runtime variance.
- Some historical docs still mention broad CLI command trees; this is acceptable as long as product positioning remains TUI/OpenCode-first and CLI is framed as debug/plumbing.

## Closure Verdict

**PASS WITH NOTES**

- Pass: required documentation, QA cross-linking, guardrail consolidation, roadmap/readme alignment, and final automated test validation are complete.
- Notes: final release sign-off should include one human-run OpenCode smoke pass using the demo flow above before tagging.
