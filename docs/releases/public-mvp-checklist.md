# Public MVP Release Checklist (Phase 7.20)

## Scope Gate (must pass)
- [ ] No Phase 8 work included.
- [ ] No real web execution features.
- [ ] No crawler/scanner/fuzzing/tool integrations.
- [ ] No POST/auth/cookies/live runtime flow execution.
- [ ] No `.env` or secrets handling.

## Product Positioning Gate
- [ ] OpenCode/TUI is described as the primary UX.
- [ ] CLI is described only as support/debug/status.
- [ ] Messaging does not imply autonomous runtime execution.

## Install/Setup Done-Summary Gate
- [ ] Shows OpenCode detected/config root.
- [ ] Shows Bitsentry pack status.
- [ ] Shows native agent status.
- [ ] Shows commands status.
- [ ] Shows security flows availability.
- [ ] Shows edit deny contract (`agent.bitsentry.permission.edit=deny`).
- [ ] Shows MCP config preview-only status (when applicable).
- [ ] Ends with final result (`PASS` / `PASS WITH NOTES` / `FAIL`).

## Doctor Gate
- [ ] `bitsentry-ai doctor` includes concise MVP readiness block.
- [ ] Readiness block includes the same core statuses where feasible.
- [ ] Doctor ends with `PASS` / `PASS WITH NOTES` / `FAIL`.

## Docs Gate
- [ ] README has public MVP quickstart and explicit constraints.
- [ ] README references `docs/demo/public-mvp-prompts.md`.
- [ ] Roadmap includes Phase 7.20 entry and keeps Phase 8 untouched as next.

## Verification Gate
- [ ] `go test ./...` is green.
- [ ] New/updated tests cover done/review messaging and doctor verdict behavior.

## Public Demo Readiness Notes
- Preferred story: "controlled manual MVP" not "automated pentest engine".
- Use PASS taxonomy honestly: PASS / PASS WITH NOTES / FAIL.
- If notes exist, list them explicitly and avoid over-claiming capability.
