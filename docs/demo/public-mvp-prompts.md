# Public MVP Demo Prompts (Phase 7.20)

Use these prompts for public demos (LinkedIn/blog/video) while staying inside MVP boundaries.

## Demo Goals
- Show OpenCode/TUI-first install and readiness UX.
- Show clear guardrails and honest constraints.
- Avoid any claim of autonomous runtime or live web execution.

## Prompt 1 — Install/Setup walkthrough
```text
I want a clean public MVP walkthrough.
Open Install / Setup and explain each step briefly.
At the end, summarize OpenCode detection, pack status, native agent status, commands status,
security flows availability, edit permission deny, MCP preview-only, and final result.
```

## Prompt 2 — Safety posture statement
```text
Before changing anything, state the safety boundaries for this MVP:
- no .env or secrets
- no live runtime flow execution
- no crawler/scanner/fuzzing integrations
- no POST/auth/cookies runtime execution
- preserve agent.bitsentry.permission.edit=deny
Then continue with non-mutating checks only.
```

## Prompt 3 — Doctor status as support/debug
```text
Run bitsentry-ai doctor and explain the MVP readiness block.
Keep CLI framing as support/debug/status, not primary UX.
Return PASS / PASS WITH NOTES / FAIL and why.
```

## Prompt 4 — OpenCode handoff narrative
```text
Assume install finished.
Give me a publication-ready handoff script for OpenCode:
- what is installed
- what remains manual
- how to use /bit-* commands
- what not to claim in this MVP
```

## Prompt 5 — Constraints QA check
```text
Do a strict constraints check for Phase 7.20 Public MVP Polish.
List any text or behavior that could be interpreted as Phase 8, live web execution,
or autonomous security testing.
```

## Demo checklist (during presentation)
- Show TUI Install/Setup as PRIMARY path.
- Show done summary with explicit status bullets.
- Show doctor readiness section as support/debug.
- Explicitly state preview-only MCP behavior.
- Explicitly state no `.env`/secrets and no live runtime web execution.
