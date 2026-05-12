# bitsentry-ai

`bitsentry-ai` is the BitSentry local CLI/TUI bootstrap tool. It provides an intent-aware orchestration system that routes user requests to appropriate development workflows (SDD, SDR, Support), manages AI agent integrations (OpenCode), and installs capability packs for persistent memory, documentation lookup, and skill-based task execution.

## Table of Contents

1. [Overview](#1-overview)
2. [Installation](#2-installation)
3. [TUI Reference](#3-tui-reference)
4. [Integrations](#4-integrations)
5. [Intent-Aware Orchestration](#5-intent-aware-orchestration)
6. [Architecture](#6-architecture)
7. [CLI Reference](#7-cli-reference)
8. [Profiles](#8-profiles)
9. [Development](#9-development)

---

## 1. Overview

### What is bitsentry-ai?

A bootstrap tool that:

- **Detects** local AI tooling (OpenCode, Engram, Context7, MCPs)
- **Orchestrates** development workflows through intent routing
- **Installs** capability packs (skills, flows, MCPs) to target agents
- **Manages** profiles for different work contexts

### Intent-Aware Orchestration System

```
User Intent (natural language)
    │
    ▼
┌─────────────────────────────────────┐
│  Route Decision Engine              │
│  - Intent classification (8 intents)│
│  - Confidence scoring               │
│  - Decision contract               │
└─────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────┐
│  Router → Planner → Session Store  │
│  - Flow selection (SDD/SDR/Support) │
│  - Execution plan generation        │
│  - Session persistence              │
└─────────────────────────────────────┘
    │
    ▼
Execution Phases (per flow)
```

### Phase 6.2 Capability Selection Preview (OpenCode-first)

- Route preview now includes capability selection preview fields before any deep discovery/planning:
  - `primary_skills`, `secondary_skills`, `deferred_skills`
  - `primary_roles`, `secondary_roles`
  - `capability_reason`, `capability_gates`
- Preview remains strictly non-executing/non-persistent:
  - `no_edits_in_preview`
  - `no_persistence_in_preview`
  - `no_flow_execution_in_preview`
- OpenCode native prompt contract is PRIMARY UX for route + capability preview.
- `bitsentry-ai route decide` stays debug/plumbing parity only (not the principal user workflow).

### Phase 6.4 MCP Readiness & Validation (OpenCode-only)

- Added a simple, testable MCP readiness model used by Engram/Context7 and wizard summaries:
  - `status`
  - `detected_evidence`
  - `blockers`
  - `manual_hints`
  - `safe_usable`
- Readiness states now clearly distinguish:
  - `configured`
  - `detected`
  - `missing`
  - `modeled_only`
  - `not_implemented`
  - `manual_step_needed`
- Install/Setup now surfaces readiness in review + done screens with manual next steps when applicable.
- No secret/token reads or writes were added; MCP credential mutation remains out of scope.

### Phase 6.5 OpenCode MCP Config Preview (OpenCode-only)

- Added a PREVIEW-ONLY MCP config model for `opencode.json` safety planning:
  - `current_config_state`, `exists`, `readable`, `invalid_error`
  - `current_mcp_config_detected`, `mcp_readiness_state`
  - `proposed_safe_changes`, `preserved_keys`, `preserved_mcp_entries`
  - `warnings`, `manual_steps`
  - `would_write`, `requires_confirmation`, `backup_required`
- Preview builder is pure/read-only:
  - no writes
  - no backups created
  - no config mutation
  - no `.env` reads
  - no secrets/tokens injection
- TUI Install/Setup now shows this model in review + done/control-panel screens and labels it explicitly as **PREVIEW ONLY**.
- Future apply remains separate and gated by explicit confirmation + backup requirement.

### Phase 6.6 Release Stabilization (OpenCode-only)

- Consolidated release closure in `docs/opencode-phase6-6-release-stabilization.md`.
- Stabilized OpenCode-first product framing:
  - TUI **Install Everything** as main user path
  - MCP Readiness + MCP Config Preview as visible safety checkpoints
  - Route+Capability Preview as primary in-chat OpenCode behavior
- Confirmed CLI remains debug/plumbing only and not primary UX.
- Preserved non-mutation guardrails (`agent.bitsentry.permission.edit=deny`, no flow/skill execution, no autonomous runtime, no MCP credential mutation).

### Phase 7.0 Compact Direct Mode & Agent Token Efficiency (OpenCode-only)

- Updated native OpenCode prompt contract to support **Compact Direct Mode** for direct/atomic/concrete requests with clear intent.
- Compact Direct Mode now avoids printing full Route Decision Preview and full Capability Preview blocks when they add no value.
- Full Route Decision Preview remains mandatory for open/ambiguous/exploratory/planning/architecture/security/multi-file/sensitive requests.
- Sensitive direct requests keep compact output but require brief explicit confirmation before mutation-risk actions.
- Preserved guardrails and boundaries:
  - no `.env` or secrets handling
  - no MCP credential mutation
  - no runtime activation or flow execution in preview
  - no persistence/edits in preview
  - OpenCode-first UX, CLI as debug/plumbing parity
- `agent.bitsentry.permission.edit=deny`

### Phase 7.3 Security Report Markdown Contract (OpenCode-only)

- Hardened security flow documentation contracts (no runtime execution) for:
  - `assets/skills/security/security-findings/SKILL.md`
  - `assets/skills/security/security-report/SKILL.md`
- Enforced exact report section contract and minimum finding token contract with strict static tests.
- Enforced required value enums:
  - Severity: `Critical | High | Medium | Low | Informational`
  - Confidence: `High | Medium | Low`
- Added explicit handoff contract `security-findings -> security-report` checks.
- Kept flow stability: no flow ID change, no stage ID/order changes, no runtime flow execution.
- Preserved guardrails: read-only first, no `.env`/secrets, no exploit/live target testing, no destructive actions, no MCP credential mutation, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing only, `agent.bitsentry.permission.edit=deny`.

### Phase 7.4A Security Findings Taxonomy & Calibration (OpenCode-only)

- Defined official findings taxonomy enum for source security review normalization:
  - Authentication, Authorization, Session Management, Input Validation, Injection, Cross-Site Scripting, Server-Side Request Forgery, File Upload, Secrets Exposure, Cryptography, Dependency Risk, GraphQL Security, Business Logic, Configuration, Logging / Monitoring, Error Handling, Data Exposure, Supply Chain, Informational.
- Added explicit Severity calibration anchors using **Impact × Likelihood** and Confidence calibration anchors by evidence quality.
- Added explicit skill-to-category mapping contract (primary/secondary) for specialized security skills.
- Added formal rules for deduplication, evidence grouping, and assumptions/limitations.
- Hardened findings→report handoff so `security-report` must consume taxonomy in both Risk Summary and Findings.
- Added strict static anchor/token tests for taxonomy, calibrations, mapping, dedup/evidence/assumptions rules, and report-consumption contract.
- Kept flow stability: ID `source-security-review` unchanged, same stage IDs/order, no runtime flow execution.
- Preserved all guardrails (`read-only first`, `no .env access`, `no secrets`, `no exploit execution`, `no external target testing`, `no destructive actions`, `no MCP credential mutation`, `no autonomous mode`, `no edits by default`, `OpenCode-first`, `CLI debug/plumbing only`, `agent.bitsentry.permission.edit=deny`).

### Phase 7.4B Security Findings/Report Fixtures & Golden Contracts (OpenCode-only)

- Added safe synthetic docs structure for source-security-review contracts:
  - `assets/docs/security/examples/findings-example.md`
  - `assets/docs/security/examples/report-example.md`
  - `assets/docs/security/fixtures/findings-golden.md`
  - `assets/docs/security/fixtures/report-golden.md`
  - `assets/docs/security/README.md`
- Added static contract tests validating:
  - fixture/example presence,
  - minimum finding token coverage in docs fixtures,
  - severity coverage in `findings-golden` (Critical/High/Medium/Low/Informational),
  - confidence enum presence (High/Medium/Low),
  - taxonomy usage examples,
  - deduplication/evidence-grouping/assumptions-limitations anchors,
  - required report section order in `report-golden`.
- Preserved flow/runtime stability: no changes to `source-security-review` flow identity, stages, runtime execution, route decision, schema, or registry.
- Preserved all guardrails for docs/examples (`read-only first`, no `.env`/secrets, no exploit/live-target behavior, no destructive actions, no MCP credential mutation, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing only, `agent.bitsentry.permission.edit=deny`).

### Available Flows

| Flow | Purpose | Skills | Status |
|------|---------|--------|--------|
| **SDD** (Spec Driven Development) | Structured software/product change | init → explore → propose → spec → design → tasks → apply → verify → archive | Declarative |
| **SDR** (Structured Discovery Research) | Research, notes, blog content, idea validation | capture → research → synthesis → questions → structure → validate → archive | Declarative |
| **Support** | Helper utilities (registry, review, testing, issues, PRs) | skill-registry, judgment-day, go-testing, skill-creator, issue-creation, branch-pr | Declarative |

---

## 2. Installation

### OpenCode-first Quickstart (Recommended)

1. Run `./install.sh`.
2. Launch TUI with `bitsentry-ai`.
3. Open **Install / Setup** and choose **Install Everything**.
4. In review/done screens, verify:
   - MCP Readiness summary
   - MCP Config Preview (**PREVIEW ONLY**, no-write)
5. In OpenCode, use a route preview prompt (example):
   - `@bitsentry Quiero mejorar el wizard del TUI`
6. Confirm Route+Capability Preview appears before deep planning/editing.

> CLI commands remain available for debugging/plumbing parity, but OpenCode + TUI is the primary UX.

### Quick Install

```bash
./install.sh
```

### Installation Options

```bash
./install.sh --dry-run      # Preview without modifying anything
./install.sh --skip-doctor  # Skip environment checks
./install.sh --prefix ~/.local/bin  # Custom installation path
```

### Verify Installation

```bash
bitsentry-ai doctor         # Check environment (OS, arch, shell, dependencies)
bitsentry-ai version        # Show version
```

### Install from Source

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
./bitsentry-ai version
go test ./...              # Run tests
```

### OpenCode Dogfooding (Local Smoke Test)

```bash
./scripts/install-opencode-local.sh
```

This script:
1. Checks `go` and `gofmt`
2. Formats Go files and runs `go test ./...`
3. Builds `./bin/bitsentry-ai`
4. Configures `bitsentry-dev` profile
5. Runs OpenCode export preview, dry-run, and real export
6. Verifies `OPENCODE_USAGE.md` and `skill-registry.md` in the export root

---

## 3. TUI Reference

### Launch TUI

```bash
bitsentry-ai
```

### Navigation Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Enter` | Select |
| `Esc` / `Backspace` | Back to home |
| `q` / `Ctrl+C` | Quit |

### Available Screens

| Screen | Description |
|--------|-------------|
| [Home Menu](#home-menu) | Main navigation hub |
| [Install / Setup](#install--setup) | Wizard for OpenCode capability installation |
| [System Check](#system-check) | OS, arch, shell, dependencies |
| [Detect AI Agents](#detect-ai-agents) | Agent detection status |
| [Components Status](#components-status) | Engram, Context7, MCPs, Skills |
| [Capabilities](#capabilities) | Skill/flow/MCP selector |
| [Profiles](#profiles) | Profile management |
| [Workflows](#workflows) | Workflow registry (stubs) |
| [Settings](#settings) | Config paths and preferences |

---

### Home Menu

The landing screen after launching `bitsentry-ai`. Shows:

- Active profile
- Available screens (9 total)
- Navigation hints

---

### Install / Setup

6-step wizard reframed as an OpenCode installer/control panel.

**Step 1 — Confirm OpenCode target**
- OpenCode detection (binary, config root, pack status)

**Step 2 — Choose install mode**
- `Install Everything` (main path)
- `Install Bitsentry Pack`
- `Update/Reinstall Bitsentry Pack` (main path when already installed)

**Step 3 — Install intent summary**
- Clear what installer will do vs will NOT do
- Explicit non-execution boundaries (no flow execution, no skill execution, no autonomous runtime)

**Step 4 — Review install plan**
- Detected state + selected mode + concrete operations
- Manual action notes when needed

**Step 5 — Install / Update**
- Execute selected install mode with backup safeguards

**Step 6 — Done / Control panel summary**
- Final status and manual notes
- OpenCode detection/config roots
- Pack/native integration status and backup paths
- OpenCode test prompt

| Key | Action |
|-----|--------|
| `Space` | Toggle/select option |
| `Enter` | Continue to next step |
| `r` | Refresh detection |
| `Esc` | Back to previous step |

---

### System Check

Displays system information:

```
- OS: darwin
- Arch: arm64
- Shell: /bin/zsh
- Package manager: brew (/opt/homebrew/bin/brew)
- Dependencies:
  - go: found [/usr/local/go/bin/go]
  - gofmt: found
  - git: found
  - ...
```

---

### Detect AI Agents

Shows detected AI agents:

```
- OpenCode (opencode): detected
  path: /usr/local/bin/opencode
  version: 1.x.x
```

---

### Components Status

Runtime-detected component status:

| Component | Status | Details |
|-----------|--------|---------|
| **Engram** | Runtime detected | Binary path, version, data dir, config status |
| **Context7** | Runtime detected | Command, path, config status |
| **MCPs** | Metadata registry | Modeled: engram, context7, postgres, filesystem, git, github, browser, firecrawl |
| **Skills** | Metadata registry | Core families + focused skills (declarative) |

---

### Capabilities

MVP selector for OpenCode target with toggles.

| Key | Toggle |
|-----|--------|
| `[` / `]` | Cycle presets |
| `m` | Engram MCP |
| `c` | Context7 MCP |
| `p` | PostgreSQL MCP (modeled) |
| `1` | SDD flow |
| `2` | SDR flow |
| `3` | notes flow |
| `4` | redteam flow |
| `z` | bitsentry-sdd skill |
| `x` | research-init skill |
| `n` | bugbounty-notes skill |

| Key | Action |
|-----|--------|
| `s` | Save draft |
| `v` | Validate |
| `l` | Plan preview |
| `d` | Apply dry-run |
| `a` then `y` | Real apply |

---

### Profiles

Select and activate profiles.

| Key | Action |
|-----|--------|
| `↑` / `↓` or `j` / `k` | Navigate |
| `Enter` | Activate selected profile |

---

### Workflows

Registry of available workflows (currently declarative stubs).

---

### Settings

Configuration information:

```
- Config directory: ~/.bitsentry-ai/
- Config file: ~/.bitsentry-ai/config.yaml
- Active profile: bitsentry-dev
- Dry-run: off
```

---

## 4. Integrations

### 4.1 Agents

#### OpenCode

OpenCode integration with full lifecycle management:

```bash
# Status and inspection
bitsentry-ai agents opencode status
bitsentry-ai agents opencode inspect-config

# Export lifecycle
bitsentry-ai agents opencode export-preview
bitsentry-ai agents opencode export

# Apply lifecycle
bitsentry-ai agents opencode apply-plan
bitsentry-ai agents opencode patch-plan
bitsentry-ai agents opencode apply

# Backup and restore
bitsentry-ai agents opencode backups
bitsentry-ai agents opencode restore
```

**Capabilities:**
- JSON validation with top-level key preservation (`$schema`, `agent`, `mcp`, `permission`, `provider`)
- Unknown MCPs preserved (aurea-core, aurea-documents, aurea-fx, obsidian)
- Managed MCPs: context7, engram
- Dry-run safety (no file modification, no backup creation)
- Pre-restore snapshots

---

### 4.2 MCPs (Model Context Protocol)

| MCP | Status | Description |
|-----|--------|-------------|
| **Engram** | ✅ Runtime detected | Persistent memory system |
| **Context7** | ✅ Runtime detected | Documentation lookup |
| **PostgreSQL** | 🔲 Modeled | Database integration |
| **Filesystem** | 🔲 Modeled | File operations |
| **Git** | 🔲 Modeled | Git operations |
| **GitHub** | 🔲 Modeled | GitHub integration |
| **Browser** | 🔲 Modeled | Web browsing |
| **Firecrawl** | 🔲 Modeled | Web scraping |

---

### 4.3 Skills

Declarative skill packs organized by family:

#### Core Skills Pack

| Family | Skills |
|--------|--------|
| **SDD** (8 skills) | sdd-init, sdd-explore, sdd-propose, sdd-spec, sdd-design, sdd-tasks, sdd-apply, sdd-verify, sdd-archive |
| **SDR** (7 skills) | sdr-capture, sdr-research, sdr-synthesis, sdr-questions, sdr-structure, sdr-validate, sdr-archive |
| **Support** (6 skills) | skill-registry, judgment-day, go-testing, skill-creator, issue-creation, branch-pr |

#### Skill Loading

Skills are loaded via the `skill` tool when task context matches skill description. Available skills:

- `bitsentry-sdd-*` — SDD phase skills
- `bitsentry-sdd-orchestrator` — SDD flow orchestrator
- `bitsentry-sdr-*` — SDR phase skills
- `bitsentry-sdr-orchestrator` — SDR flow orchestrator
- `bitsentry-support-*` — Support utilities
- `branch-pr` — PR creation workflow
- `go-testing` — Go testing patterns
- `issue-creation` — Issue creation workflow
- `judgment-day` — Adversarial review
- `research-bitsentry-*` — BitSentry research utilities
- `skill-creator` — Skill creation
- `skill-registry` — Registry management

---

### 4.4 Flows

#### SDD (Spec Driven Development)

```
init → explore → propose → spec → design → tasks → apply → verify → archive
```

**Gates:**
- `proposal-approval` — Before spec/design
- `pre-implementation` — Before apply
- `final-review` — Before archive

**Handoffs:**
- → SDR (deeper research)
- → judgment-day (adversarial review)
- → go-testing (Go verification)
- → issue-creation (follow-up)
- → branch-pr (branch/PR plan)

---

#### SDR (Structured Discovery Research)

```
capture → research → synthesis → questions → structure → validate → archive
```

**Gates:**
- `continuation-review` — Before structure
- `quality-review` — Before archive

**Handoffs:**
- → SDD (implementation needed)
- → issue-creation (follow-up)
- → judgment-day (controversial topics)

---

#### Support

Independent utility invocations (not a mandatory sequential pipeline):

- `skill-registry` — Scan and maintain capability map
- `judgment-day` — Adversarial review with PASS/PASS WITH NOTES/PARTIAL/BLOCKED verdict
- `go-testing` — Hermetic Go testing guidance
- `skill-creator` — Create/improve skills
- `issue-creation` — Issue draft from findings
- `branch-pr` — Safe branch/PR plan

---

## 5. Intent-Aware Orchestration

### Intent Registry (8 declared intents)

| Intent | Default Flow | Complexity | Requires Confirmation |
|--------|--------------|------------|----------------------|
| `architecture-change` | SDD | medium | yes |
| `frontend-ux-change` | SDD | medium | yes |
| `bug-investigation` | Support → SDD | low | no |
| `research-analysis` | SDR | medium | no |
| `security-review` | Source Security Review | high | yes |
| `documentation-change` | SDD | low | no |
| `direct-answer` | none | — | no |

### Route Decision Preview

The primary UX for `@bitsentry` chat. Before non-trivial work:

```
Route Decision Preview
- matched_intent: <intent>
- decision: <decision>
- matched_signals: [<signals>]
- reason: <reason>
- requires_bounded_discovery: true/false
- requires_confirmation: true/false
- gates: [no_edits_in_preview, no_persistence_in_preview, no_flow_execution_in_preview]
```

### Available Specialist Roles (14)

| Role | Category | Use When |
|------|----------|----------|
| `codebase-onboarding` | engineering | New codebase context |
| `software-architect` | engineering | Architecture, integration, contracts |
| `frontend-engineer` | engineering | UI/TUI/wizard/layout |
| `backend-engineer` | engineering | API, services, data |
| `test-engineer` | engineering | Testing strategy |
| `security-reviewer` | security | Security review |
| `appsec-reviewer` | security | AppSec analysis |
| `threat-modeler` | security | Threat modeling |
| `product-analyst` | product | Research, content, notes |
| `ux-flow-designer` | product | UX flow design |
| `technical-writer` | product | Documentation |
| `bug-triage-engineer` | support | Bug investigation |
| `code-reviewer` | support | Code review |
| `incident-analyst` | support | Incident response |

### Decision Matrix

| Intent signal | Decision | Default route |
|--------------|----------|---------------|
| spec, design, feature, change | `use_flow_sdd` | SDD |
| security, appsec, threat, risk, source review | `use_flow_source-security-review` | Source Security Review |

### Source Security Review (Phase 7.1 MVP)

- Declarative/read-only-first flow: `source-security-review`
- Scope: source code, non-secret config, dependencies, risky patterns
- Out of scope: pentest runtime, exploit execution, external target testing
- Guardrails preserved: no `.env` access, no secrets, no destructive actions, no MCP credential mutation, no runtime flow execution, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing parity only, `agent.bitsentry.permission.edit=deny`

### Security Skill Pack MVP (Phase 7.2)

- Added 8 support security skills (checklists/references, not flow stages):
  - `security/auth-security-review`
  - `security/jwt-review`
  - `security/graphql-security-review`
  - `security/xss-review`
  - `security/file-upload-review`
  - `security/ssrf-review`
  - `security/secrets-review`
  - `security/dependency-risk-review`
- Kept flow contract stable: same ID (`source-security-review`) and same stage chain (`security-init` → `security-scope` → `security-map` → `security-review` → `security-findings` → `security-report`).
- Guardrails remain explicit and unchanged: read-only first, no `.env` access, no secrets handling, no exploit execution, no external target testing, no destructive actions, no MCP credential mutation, no runtime flow execution, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing only, `agent.bitsentry.permission.edit=deny`.
| support, bug, help, troubleshoot, error | `use_flow_support` | Support |
| direct/simple explanation | `direct_answer` | none |

---

## 6. Architecture

### Package Map

```
cmd/bitsentry-ai/
  └── Binary entrypoint

internal/
  ├── cli/            # Cobra command tree (version, doctor, agents, profiles, components, capabilities, route)
  ├── app/            # Application wiring and shared runtime services
  ├── config/         # Config path resolution and persistence (~/.bitsentry-ai/config.yaml)
  ├── system/         # OS/arch/shell detection and dependency checks
  ├── agents/         # Agent detector contracts and OpenCode detection
  ├── profiles/       # Default profile catalog and profile selection/persistence
  ├── components/     # Component registry (Engram, Context7, MCPs, Skills)
  ├── workflows/      # Workflow registry (stubs)
  ├── capabilities/   # Capability service (catalog, validation, planning, export)
  ├── orchestrator/   # Intent routing (Router, Planner, Session Store)
  ├── tui/            # Bubble Tea based interactive UI (model, screens, wizard, styles)
  └── logs/            # Logger initialization and log path handling
```

### Intent Routing Flow

```
                    ┌────────────────────┐
                    │  Route Decision    │
                    │  (route_decision.go)│
                    │  - 8 intents       │
                    │  - Confidence      │
                    │  - Decision contract│
                    └─────────┬──────────┘
                              │
                              ▼
                    ┌────────────────────┐
                    │  Router            │
                    │  (router.go)       │
                    │  RouteIntentToFlow │
                    │  Heuristics        │
                    └─────────┬──────────┘
                              │
                              ▼
                    ┌────────────────────┐
                    │  Planner           │
                    │  (planner.go)       │
                    │  BuildExecutionPlan│
                    └─────────┬──────────┘
                              │
                              ▼
                    ┌────────────────────┐
                    │  Session Store     │
                    │  (store.go)        │
                    │  SaveSession/      │
                    │  LoadSession       │
                    └────────────────────┘
```

### Capability Service

```
SelectionDraft → Validate → Plan → ApplyDryRun/Apply
     │              │         │
     ▼              ▼         ▼
  SaveConfig    Check      Project
  LoadConfig    catalog    OpenCode
```

### TUI Architecture

```
model.go (state management)
    │
    ├── screens.go (render functions)
    ├── install_wizard.go (install flow)
    └── styles.go (Lipgloss styling)
```

---

## 7. CLI Reference

### Full Command Tree

```text
bitsentry-ai
├── version
├── doctor
├── agents
│   └── opencode
│       ├── status
│       ├── inspect-config
│       ├── export-preview
│       ├── export
│       ├── apply-plan
│       ├── patch-plan
│       ├── apply
│       ├── backups
│       └── restore
├── profiles
├── profile use <name>
├── components
│   ├── <component> status
│   ├── mcps list
│   ├── skills list
│   ├── <component> configure [--dry-run]
│   ├── engram status
│   ├── context7 status
│   ├── mcps status
│   └── skills status
├── capabilities
│   ├── status
│   ├── validate
│   ├── plan
│   ├── export-preview
│   ├── export
│   ├── configure
│   ├── apply
│   ├── inspect <capability>
│   └── report latest
├── route
│   ├── inspect
│   ├── preview
│   ├── decide [prompt]
│   ├── start [intent] [--flow-hint <flow>]
│   ├── report --session <id>
│   ├── list [--flow <flow>] [--status <status>] [--archived]
│   ├── archive --session <id>
│   ├── restore --session <id>
│   ├── status --session <id>
│   ├── handoff --session <id> [--output <path>]
│   ├── resume --session <id>
│   ├── next --session <id>
│   ├── progress --session <id>
│   ├── mark-current --session <id> --stage <id>
│   ├── mark-done --session <id> --stage <id>
│   ├── validate --session <id>
│   ├── migrate --session <id> (--dry-run | --apply --confirm)
│   ├── repair --session <id> (--dry-run | --apply --confirm)
│   ├── audit [--include-archived]
│   └── cleanup [--older-than <duration>] [--flow <flow>] [--status <status>] [--completed-only] (--dry-run | --apply --confirm)
├── config path
└── --dry-run
```

### Route Commands Detail

| Command | Description |
|---------|-------------|
| `route inspect` | List available flows and stages |
| `route preview [intent]` | Preview route decision without side effects |
| `route decide [prompt]` | Show route decision envelope (debug parity) |
| `route start [intent] --flow-hint <flow>` | Create planned session |
| `route list` | List persisted sessions |
| `route status --session <id>` | Show session status |
| `route progress --session <id>` | Show declarative progress |
| `route resume --session <id>` | Resume session |
| `route next --session <id>` | Recommend next step |
| `route validate --session <id>` | Validate session read-only |
| `route cleanup` | Archive completed sessions |

### Capabilities Commands Detail

```bash
# Configure from preset
bitsentry-ai capabilities configure --target-agent opencode --preset bitsentry-dev

# Configure custom selection
bitsentry-ai capabilities configure --target-agent opencode --mcp engram --mcp context7 --flow sdd --flow support

# Clear selections
bitsentry-ai capabilities configure --clear-mcps
bitsentry-ai capabilities configure --clear-skills --clear-flows
bitsentry-ai capabilities configure --reset-all
bitsentry-ai capabilities configure --clear-all

# Plan from saved config
bitsentry-ai capabilities plan

# Validate and apply
bitsentry-ai capabilities validate
bitsentry-ai --dry-run capabilities apply  # Dry-run
bitsentry-ai capabilities apply            # Real apply

# Inspect
bitsentry-ai capabilities inspect engram
bitsentry-ai capabilities report latest
```

---

## 8. Profiles

### Available Profiles (8)

| Profile | Description |
|---------|-------------|
| `default` | Default configuration |
| `minimal` | Minimal setup |
| `development` | Development-focused |
| `research` | Research and discovery |
| `blog` | Blog content creation |
| `oscp` | OSCP-style security |
| `bug-bounty` | Bug bounty hunting |
| `redteam` | Red team operations |

### Activate a Profile

```bash
bitsentry-ai profile use <name>
```

Example:
```bash
bitsentry-ai profile use development
```

---

## 9. Development

### Build

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
```

### Test

```bash
go test ./...
```

### Local Smoke Test

```bash
./scripts/install-opencode-local.sh
```

### Run TUI

```bash
bitsentry-ai
```

### Run Doctor

```bash
bitsentry-ai doctor
```

---

## Links

- [Installation guide](docs/install.md)
- [Architecture overview](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
