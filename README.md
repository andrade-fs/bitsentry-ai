# BitsentryAI — Public MVP

BitsentryAI is an OpenCode/TUI-first bootstrap and control layer for local, guided security/developer workflows.

This public MVP is intentionally constrained: it prioritizes safe setup, explicit readiness signals, and manual control over automation.

## What is BitsentryAI (MVP)

- A local installer + workflow shell focused on OpenCode as primary UX.
- A TUI-first setup experience (`bitsentry-ai` → Install/Setup).
- A route/flow guidance system with controlled, manual posture.

References:
- Install: [docs/install.md](docs/install.md)
- Architecture: [docs/architecture.md](docs/architecture.md)

## What it does / does not do

### What it does now
- Detects local prerequisites and OpenCode context.
- Guides installation through TUI and shows explicit readiness status.
- Supports CLI commands for support/debug flows (for example `version`, `doctor`).
- Preserves guardrails and explicit manual boundaries.

### What it does not do in this MVP
- No live web execution from this product posture.
- No hidden runtime automation or autonomous exploitation.
- No `.env` parsing or secrets/token handling.
- No over-promised “one-click full assessment” behavior.

## Quickstart (TUI-first)

1. Build/install following [docs/install.md](docs/install.md).
2. Run:
   ```bash
   bitsentry-ai
   ```
3. Open **Install / Setup** and complete the wizard.
4. Review the final readiness verdict (`PASS`, `PASS WITH NOTES`, `FAIL`).
5. Continue in OpenCode with the suggested Bitsentry prompts/commands.

## How flow works (simple)

1. You express intent (chat/TUI context).
2. Bitsentry proposes route/capability guidance.
3. You confirm manually before impactful actions.
4. MVP keeps decisions explicit and auditable.

## Why context profiles matter (SDD / SDR / Security)

BitsentryAI uses the same tool surface, but different intents require different guardrails and outputs.

- **SDD profile:** for building or changing product features. Emphasizes specs, design clarity, acceptance criteria, and disciplined implementation steps.
- **SDR profile:** for investigation, research, and diagnosis. Emphasizes hypothesis framing, evidence collection, and non-mutating discovery first.
- **Security profile (hacking context, controlled):** for source/security review and authorized web-assessment planning. Emphasizes strict scope definition, authorization gates, and low-noise execution with manual approvals.

Why this is useful: teams spend less time context switching, trigger fewer wrong actions, improve traceability of decisions, keep operations safer, and communicate intent more clearly across roles.

MVP constraint reminder: profile selection guides intent and risk gates, but execution remains explicitly manual with no hidden autonomous runtime.

See system shape in [docs/architecture.md](docs/architecture.md).

## MVP readiness taxonomy

- **PASS**: Core setup is ready for intended MVP usage.
- **PASS WITH NOTES**: Usable, but with clear manual follow-ups.
- **FAIL**: Required baseline is missing or invalid.

This taxonomy is intentionally operational and honest: it reports readiness, not marketing confidence.

## Launch kit

- Public MVP checklist: [docs/releases/public-mvp-checklist.md](docs/releases/public-mvp-checklist.md)
- Public demo prompts: [docs/demo/public-mvp-prompts.md](docs/demo/public-mvp-prompts.md)

## CLI support/debug

TUI is primary. CLI remains available for inspection and troubleshooting:

```bash
bitsentry-ai version
bitsentry-ai doctor
```

Use CLI as a support/debug surface, not as the main guided onboarding path.

## Roadmap (now/next)

- **Now (Public MVP):** stable TUI-first setup, OpenCode-first guidance, controlled/manual operations.
- **Next:** incremental hardening and capability expansion without breaking current safety boundaries.

Roadmap detail: [docs/roadmap.md](docs/roadmap.md)
