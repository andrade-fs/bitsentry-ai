# bitsentry-ai

`bitsentry-ai` is the BitSentry local CLI/TUI bootstrap tool. Phase 2 delivers working Components MVP so you can inspect your environment, manage agent components (Engram, Context7, MCPs, Skills), detect local AI tooling, manage profiles, and prepare for future workflows.

## Phase 2 status (Components MVP)

Implemented now (Phase 2):
- All Phase 1 commands plus: `components`, `components <component> status`, `components mcps list`, `components skills list`, `components <component> configure --dry-run`
- Engram runtime detection (binary, version, data dir, config consistency)
- Context7 metadata configuration (no runtime validation)
- MCP metadata registry (modeled MCPs: engram, context7, postgres, filesystem, git, github, browser, firecrawl)
- Skills metadata registry (core families + focused skills; declarative)
- Component configure dry-run preview

## Phase 2.5 status (OpenCode Integration)

Implemented now (Phase 2.5):
- OpenCode MCP integration: `agents opencode status`, `agents opencode inspect-config`, `agents opencode export-preview`, `agents opencode export`, `agents opencode apply-plan`, `agents opencode patch-plan`, `agents opencode apply`, `agents opencode backups`, `agents opencode restore`
- Config safety: JSON validation, top-level key preservation ($schema, agent, mcp, permission, provider)
- Unknown MCPs preserved: aurea-core, aurea-documents, aurea-fx, obsidian
- Managed MCPs: context7, engram
- Dry-run safety: no file modification, no backup creation
- Backup/restore with pre-restore snapshots
- Capability installer foundation (registry + planner + OpenCode projection): `capabilities status`, `capabilities plan --target-agent opencode --preset <id>`
- Capability selection persistence + staged apply MVP: `capabilities configure ...`, `capabilities apply [--dry-run]`

Not implemented yet:
- Native OpenCode runtime integration for Bitsentry flows/skills (export produces a managed capability pack)
- Real SDR workflows
- Real SDD workflows
- Real Red Team / Bug Bounty workflow execution

## Phase 1 status (Bootstrap MVP)

Implemented now:
- CLI commands: `version`, `doctor`, `agents`, `profiles`, `profile use <name>`, `components`, `config path`
- Minimal TUI (launch with no subcommand)
- Environment doctor (OS/arch/shell/dependencies)
- OpenCode agent detection
- Profile system (8 default profiles)
- Components/workflows registries as **stubs**
- Local config path: `~/.bitsentry-ai/config.yaml`

Not implemented yet:
- Real SDR workflows
- Real SDD workflows
- Real Red Team / Bug Bounty workflow execution

## Installation

```bash
./install.sh
```

Useful options:

```bash
./install.sh --dry-run
./install.sh --skip-doctor
./install.sh --prefix "$HOME/.local/bin"
```

## Development usage

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
./bitsentry-ai version
./bitsentry-ai doctor
go test ./...
```

## Local OpenCode dogfooding smoke script (Phase 4.0.2)

Run the one-command local installer/smoke flow:

```bash
./scripts/install-opencode-local.sh
```

What it does:
- checks `go` and `gofmt`
- formats Go files and runs `go test ./...`
- builds `./bin/bitsentry-ai`
- configures `bitsentry-dev`
- runs OpenCode export preview, dry-run, and real export
- detects the actual managed OpenCode export root (`~/.opencode/bitsentry` or `~/.config/opencode/bitsentry`)
- verifies `OPENCODE_USAGE.md` and `skill-registry.md` in the detected export root
- prints the resolved export path in the final PASS summary

If needed, make it executable once:

```bash
chmod +x ./scripts/install-opencode-local.sh
```

## TUI Install / Setup Wizard (Phase 4.0.3)

Launch the TUI and open `Install / Setup`:

```bash
bitsentry-ai
```

Wizard MVP capabilities:
- Detect OpenCode status, config root, and Bitsentry pack status
- Select install target (OpenCode), preset (`bitsentry-dev` default), and MCP toggles (Engram/Context7)
- Review install plan and explicit non-actions (no `opencode.json` mutation, no runtime flow execution)
- Confirm install and export capability pack
- Verify `OPENCODE_USAGE.md` and `skill-registry.md`
- Show next OpenCode dogfooding prompt with detected real paths

## CLI command list

```text
bitsentry-ai
├── version
├── doctor
├── agents
├── agents opencode status
├── agents opencode inspect-config
├── agents opencode export-preview
├── agents opencode export
├── agents opencode apply-plan
├── agents opencode patch-plan
├── agents opencode apply
├── agents opencode backups
├── agents opencode restore
├── profiles
├── profile use <name>
├── components
├── components <component> status
├── components mcps list
├── components skills list
├── components <component> configure [--dry-run]
├── components engram status
├── components context7 status
├── components mcps status
├── components skills status
├── capabilities status
├── capabilities validate
├── capabilities plan
├── capabilities export-preview
├── capabilities export
├── capabilities configure
├── capabilities apply
└── config path
```

## Capability installer usage (Phase 3.x)

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

# Validate saved capability config
bitsentry-ai capabilities validate

# Safe preview apply
bitsentry-ai --dry-run capabilities apply

# Inspect one capability and latest apply report
bitsentry-ai capabilities inspect engram
bitsentry-ai capabilities report latest

# TUI capability selector (MVP)
bitsentry-ai
# Main Menu -> Capabilities
# Controls in screen:
# [ / ] preset, m/c/p MCP toggles, 1/2/3/4 flow toggles, z/x/n skill toggles
# s save draft, v validate, l plan preview, d apply dry-run, a then y real apply
```

Apply reports are persisted under:

`~/.bitsentry-ai/exports/capabilities/apply/`

Notes:
- Preview/validate/plan/report summary logic is shared in-process via `internal/capabilities`.
- Real apply remains on the existing safe OpenCode apply path (backup/merge safeguards preserved).

## Core Skills Pack (Phase 3.7, declarative)

The repository includes declarative skill assets under:

- `assets/skills/_shared/` (contracts)
- `assets/skills/sdd/` (Spec Driven Development family)
- `assets/skills/sdr/` (Structured Discovery/Research family)
- `assets/skills/support/` (helper skills)
- `assets/flows/` (flow manifests)

Current status:
- Contracts and skill definitions are available as assets.
- Runtime orchestration/execution is intentionally out of scope in this phase.

## Dynamic Asset Discovery API (Phase 3.8B)

Read-only discovery is now available in-process via `internal/capabilities`:

- `DiscoverAssets(root string) (AssetCatalog, error)`

Discovery scans:
- `assets/flows/*.yaml`
- `assets/skills/_shared/*.md`
- `assets/skills/<pack-id>/**/SKILL.md`
- `assets/orchestrators/*.md` (optional; missing directory is handled gracefully)

Catalog includes discovered flows, skill packs, skills, shared contracts, and orchestrator contracts with file metadata and validation status.

Notes:
- This phase is discovery only (no runtime orchestration, no autonomous execution).
- OpenCode export/apply behavior remains unchanged.

## Selection-aware OpenCode Export Projection (Phase 3.8C)

Read-only projection is now available in-process:

- `BuildOpenCodeExportProjection(catalog AssetCatalog, selectedIDs []string)`
- `GenerateSkillRegistry(projection OpenCodeExportProjection)`

Behavior:
- Resolves selected IDs to discovered flows/skill packs (including aliases like `bitsentry-sdd -> sdd`).
- Includes required assets for selected flows/packs:
  - `_shared` contracts
  - selected flow manifests
  - selected skill packs and discovered skills
  - support pack when selected flows require support skills
  - optional orchestrator contracts if present
- Generates a registry preview (`bitsentry/skill-registry.md`) from projection data.

Notes:
- Projection only (no file writes, no OpenCode config mutations).
- Existing apply/export runtime behavior is unchanged in this phase.

## OpenCode Skills Export (Phase 3.9)

Selection-aware export commands (generic capabilities surface):

```bash
bitsentry-ai capabilities export-preview --target-agent opencode [--select ...]
bitsentry-ai capabilities export --target-agent opencode --dry-run [--select ...]
bitsentry-ai capabilities export --target-agent opencode [--select ...]
```

Behavior:
- Uses dynamic discovery + projection to decide included assets.
- Dry-run previews written paths without modifying OpenCode files.
- Real export writes only under managed area: `<opencode-config-root>/bitsentry/`.
- Generated registry always included when assets are exported: `bitsentry/skill-registry.md`.
- OpenCode usage guide is generated on export: `bitsentry/OPENCODE_USAGE.md`.
- Export reports are persisted to `~/.bitsentry-ai/exports/opencode-skills/` with `latest.yaml`.

Important boundary:
- This export creates a **Bitsentry capability pack** for OpenCode-managed files.
- It does **not** provide native runtime orchestration/execution in OpenCode yet.
- It does **not** modify `opencode.json` automatically.

## Native OpenCode integration (Phase 4.0.4+)

When using the TUI `Install / Setup` wizard with native integration enabled, bitsentry-ai can also:
- register `agent.bitsentry`
- install `/bit-*` native command entries
- project selected actionable native skills under `<opencode-config-root>/skills/`
- merge/create `opencode.json` safely with backups while preserving unrelated user config keys

Safety notes:
- Managed pack export under `<opencode-config-root>/bitsentry/` remains unchanged.
- Runtime flow execution is still out of scope.
- MCP config mutation remains disabled unless explicitly supported by a future safe registry.


## Default profiles (8)

- default
- minimal
- development
- research
- blog
- oscp
- bug-bounty
- redteam

## Docs

- [Installation guide](docs/install.md)
- [Architecture overview](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
