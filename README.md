# BitsentryAI

![alt text](docs/image.png)


**Turn OpenCode into a guided security and development workspace.**

BitsentryAI installs local workflows, skills, prompts, commands and safety guardrails into OpenCode, so developers can move from vague AI chats to structured, auditable work: feature design, repository investigation, support triage and security review.

It is built for people who want AI assistance without hidden automation, uncontrolled actions or unsafe security behavior.

> Public MVP status: **local-first, OpenCode-first, guided workflows, safe by default.**

---

## Why BitsentryAI?

AI coding tools are powerful, but they often lack structure.

They can jump into edits too early, forget project context, mix research with implementation, or behave unpredictably during security-related work. BitsentryAI adds a local control layer around OpenCode so AI-assisted sessions become clearer, safer and easier to repeat.

BitsentryAI helps you answer questions like:

* What workflow should this request follow?
* Should the agent research, plan, review or edit?
* Which skills and prompts are relevant?
* What should be read-only?
* What requires confirmation?
* What guardrails apply before touching code, secrets or live targets?

The goal is simple: **less chaotic AI chat, more guided engineering work.**

---

## What it does today

BitsentryAI currently provides:

* **TUI-first installation** for local setup, readiness checks and guided OpenCode integration.
* **Native OpenCode integration** with a `bitsentry` agent, `/bit-*` commands, prompts and local capability packs.
* **Intent-aware routing** for development, research, support and security workflows.
* **Capability packs** containing flows, skills, roles, commands and prompt contracts.
* **Source security review workflows** with read-only-first behavior and structured findings/reporting.
* **Web assessment planning flows** with strong authorization and safety gates.
* **Doctor/status tooling** for troubleshooting local readiness.
* **Safety-first defaults**: no hidden autonomous runtime, no secret access, no uncontrolled live testing.

---

## Core idea

BitsentryAI does not try to replace OpenCode.

It **vitaminizes OpenCode** by installing a structured layer of workflows and guardrails around it:

```text
Developer request
      ↓
Bitsentry route decision
      ↓
Relevant flow / skill / role
      ↓
OpenCode-guided work
      ↓
Structured output, gates and next steps
```

Instead of asking an AI agent to “just do the thing”, BitsentryAI helps it decide whether the request is better handled as:

* a direct answer,
* a software design workflow,
* a repository investigation,
* a support triage,
* a source security review,
* or a scoped assessment planning task.

---

## Quickstart

### 1. Clone the repository

```bash
git clone https://github.com/andrade-fs/bitsentry-ai.git
cd bitsentry-ai
```

### 2. Build the binary

```bash
go build -o bin/bitsentry-ai ./cmd/bitsentry-ai
```

### 3. Run the installer TUI

```bash
./bin/bitsentry-ai tui
```

Use the **Install / Setup** flow to detect OpenCode, export the local Bitsentry pack and install the native OpenCode integration.

The installer is designed to show what it will do before applying changes.

---

## Using it with OpenCode

After installation, open OpenCode and use the `bitsentry` agent.

Example prompts:

```text
@bitsentry Analyze this repository structure.
Do not access .env or secrets.
Return a route decision, relevant files, risks and a suggested next workflow.
```

```text
@bitsentry Start a source security review.
Read-only only. No exploits, no live targets, no code edits.
```

```text
@bitsentry Help me design a new feature using SDD.
Start with scope, assumptions and acceptance criteria.
```

```text
@bitsentry Review this bug and decide whether we need support triage, repository discovery or a design flow.
Do not edit files unless I explicitly approve it.
```

BitsentryAI also installs `/bit-*` command templates into OpenCode, such as:

* `/bit-install-check`
* `/bit-pack-status`
* `/bit-sdd-init`
* `/bit-sdr-capture`
* `/bit-support-triage`

These commands are intended to make common guided workflows easier to start from inside OpenCode.

---

## Main workflows

### Software Design Development — SDD

Use SDD when you want to design or implement a feature with structure.

Typical use cases:

* new feature planning,
* acceptance criteria,
* technical design,
* phased implementation,
* code changes after explicit approval.

Default posture: **non-mutating until apply is approved.**

---

### Software Design Research — SDR

Use SDR when you need to understand a repository, architecture, feature area or technical decision before changing anything.

Typical use cases:

* repository analysis,
* architecture discovery,
* feature feasibility,
* technical debt investigation,
* codebase mapping.

Default posture: **read-only investigation.**

---

### Support flow

Use the support flow when you need to triage a problem, gather evidence, understand symptoms or prepare a clean handoff.

Typical use cases:

* bug triage,
* incident notes,
* support investigation,
* reproduction steps,
* handoff summaries.

Default posture: **diagnose before changing.**

---

### Source Security Review

Use this flow when you want a safe, source-based security review of a repository.

Typical use cases:

* authentication review,
* JWT/session review,
* GraphQL security review,
* file upload review,
* SSRF review,
* XSS review,
* secrets exposure review,
* dependency risk review.

Default posture: **read-only, no exploits, no live targets, no secret access.**

---

### Web Assessment Planning

BitsentryAI includes planning-oriented web assessment flows, but the public MVP intentionally keeps strong safety boundaries.

Typical use cases:

* authorized assessment planning,
* scope definition,
* test plan preparation,
* request review,
* findings/report structure.

Default posture: **planning-first, authorization-gated, no autonomous testing.**

---

## Safety by design

BitsentryAI is intentionally conservative.

Current public MVP guardrails:

* No hidden autonomous runtime.
* No uncontrolled code edits.
* No `.env` or secret parsing.
* No credential extraction.
* No live web execution by default.
* No scanner, crawler or fuzzer integrations.
* No one-click pentest behavior.
* No exploit execution.
* No destructive actions.
* Manual confirmation before impactful changes.

This is not a limitation hidden in fine print. It is a core product decision.

BitsentryAI is designed for controlled local workflows where the user remains in charge.

---

## What gets installed into OpenCode?

A native OpenCode setup can include:

```text
<opencode-config-root>/bitsentry/
├── agents/
│   └── bitsentry.md
├── commands/
│   └── bit-*.md
├── flows/
├── skills/
├── roles/
├── OPENCODE_USAGE.md
└── skill-registry.md
```

BitsentryAI can also update OpenCode configuration to register the `bitsentry` agent and command templates.

The intended OpenCode agent posture is safe by default:

```json
{
  "agent": {
    "bitsentry": {
      "mode": "primary",
      "permission": {
        "edit": "deny",
        "bash": "ask"
      }
    }
  }
}
```

---

## Local checks

Useful commands during development:

```bash
go test ./...
```

```bash
./bin/bitsentry-ai doctor
```

```bash
./bin/bitsentry-ai version
```

The CLI exists mostly as local plumbing, testing and diagnostics. The primary user experience is intended to be **OpenCode + TUI**.

---

## Project status

BitsentryAI is currently in a public MVP stage.

The MVP is focused on:

* local installation,
* OpenCode integration,
* guided workflows,
* route decision previews,
* source security review,
* safe assessment planning,
* public demo readiness,
* strong guardrails.

It is not yet a general-purpose autonomous agent platform, pentest automation engine or scanner.

---

## Who is this for?

BitsentryAI is for:

* developers using OpenCode who want more structure than plain chat,
* security engineers who want safe, scoped, read-only-first review workflows,
* technical leads who want repeatable AI-assisted planning and investigation,
* builders experimenting with local AI agents, MCPs and workflow packs,
* teams that want AI assistance with explicit boundaries and reviewable outputs.

---

## Roadmap direction

The current direction is:

1. Stabilize the OpenCode-first public MVP.
2. Improve guided workflows and safety contracts.
3. Harden source security review and assessment planning.
4. Expand controlled integrations only after the safety model is solid.
5. Later, support additional AI coding environments beyond OpenCode.

BitsentryAI is intentionally built step by step: first control, then capability, then automation only where it is safe and explicit.

---

## Development principles

BitsentryAI follows a few strict principles:

* **OpenCode-first**: current target is OpenCode.
* **Local-first**: install and run locally.
* **User-controlled**: no hidden autonomous behavior.
* **Read-only first**: especially for security and repository analysis.
* **Explicit gates**: risky actions require confirmation.
* **No secret access**: `.env` and sensitive data are out of scope.
* **Structured outputs**: findings, reports and decisions should be reviewable.
* **Composable capabilities**: flows, skills, roles and commands should evolve independently.

---

## Contributing

Contributions should preserve the core safety posture of the project.

Before opening a pull request, make sure:

* `go test ./...` passes,
* no secrets or local credentials are committed,
* OpenCode integration remains safe by default,
* security workflows remain read-only-first unless explicitly gated,
* README/docs accurately describe current behavior.

---

## License

Add project license information here.

---

## One-line summary

**BitsentryAI is a local workflow and safety layer for OpenCode: it turns AI coding sessions into guided, auditable development and security workflows.**
