# Roadmap

## Phase 2 — Components MVP (completed)
- Real component lifecycle (install/check/status)
- Stronger component contracts and diagnostics
- Better install/setup flows per component
- Engram runtime detection
- Context7 metadata configuration
- MCP metadata registry
- Skills metadata registry (6 available)

## Phase 2.5 — OpenCode Integration (completed)
- OpenCode MCP export/apply workflow
- Config safety: JSON validation, key preservation
- Unknown MCP preservation
- Dry-run safety verification
- Backup/restore with pre-restore snapshots

## Phase 3 — SDR MVP (current)

### Phase 3.5 — Capability service extraction (completed)
- Extract shared capability service for selection load/save, validation, plan/projection, and apply/report operations
- Reduce TUI subprocess coupling by reusing service methods
- Keep existing CLI behavior stable while centralizing reusable logic

### Phase 3.6 — In-process preview/report core + release polish (in progress)
- Share preview summary/report read-write logic in `internal/capabilities`
- Reuse shared preview/report core from both CLI and TUI/service
- Keep real apply behavior on current safe path

### Phase 3.7 — Core Skills Pack (in progress)
- Stabilize default capability presets and skill packs for common workflows
- Improve capability ergonomics for day-to-day development/research modes
- Add declarative SDD/SDR/support skill families and shared contracts
- Fill skill contracts with phase-grade content and enforce headings via tests

### Phase 3.8A — Flow Manifest Normalization (completed)
- Normalize SDD/SDR/support flow manifests for dynamic discovery/routing readiness
- Enforce flow manifest contract checks in tests

### Phase 3.8B — Dynamic Asset Discovery API (completed)
- Add read-only discovery API for flows, skill packs, skills, shared contracts, and orchestrator docs
- Validate discovered skills against strict required heading contract
- Keep runtime orchestration/export/apply behavior unchanged

### Phase 3.8C — Selection-aware OpenCode Export Projection (completed)
- Add read-only projection from selected IDs to discovered assets for OpenCode managed area
- Generate skill registry preview content from projected assets
- Preserve existing OpenCode apply behavior and runtime boundaries

### Phase 3.9 — Selection-aware OpenCode Export/Apply (completed)
- Integrate projection into safe export pipeline for OpenCode managed bitsentry area
- Add generic capabilities commands: `export-preview` and `export` (opencode target)
- Add dry-run/export report persistence and managed path safety checks

### Phase 4.0 — Orchestrator MVP (next)
- Introduce minimal runtime flow routing with strict safety checkpoints

### Phase 4.0.1 — OpenCode Dogfooding Readiness Fixes (completed)
- Align capability presets with valid discovered/exportable flow-pack aliases
- Enforce strict validation for preset references and export selection aliases
- Improve export-preview/export visibility for selected flows and skills
- Generate `bitsentry/OPENCODE_USAGE.md` in exported managed area
- Clarify docs that OpenCode export is a Bitsentry capability pack, not runtime integration

### Phase 4.0.2 — Local OpenCode Install/Smoke Script (completed)
- Add `scripts/install-opencode-local.sh` for one-command local dogfooding validation
- Run format/test/build/capability-config/export preview/dry-run/real export sequence
- Verify exported managed files (`OPENCODE_USAGE.md`, `skill-registry.md`) under `~/.opencode/bitsentry`

### Phase 4.0.3 — TUI Install / Setup Wizard MVP (completed)
- Replace Install/Setup placeholder with guided wizard flow for OpenCode
- Detect OpenCode + config root + Bitsentry pack status and verification files
- Allow target selection, preset selection, MCP toggles (Engram/Context7), plan review, and confirmed install
- Export capability pack using existing safe export logic (no `opencode.json` mutation)
- Verify `OPENCODE_USAGE.md` and `skill-registry.md` and print next dogfooding prompt

### Phase 4.0.3A — TUI Wizard UX Fixes (completed)
- Move wizard to step-by-step navigation (one step visible at a time)
- Add cursor/toggle persistence and install readiness checks

### Phase 4.0.3B — Final Prompt Rendering Fix (completed)
- Render a complete multiline OpenCode dogfooding prompt in Step 6

### Phase 4.0.4 — Native OpenCode Bitsentry Integration (completed)
- Keep managed pack export under `bitsentry/`
- Register native `agent.bitsentry` and `instructions` entrypoint
- Install native `/bit-*` command entries and projected actionable native skills
- Merge/create `opencode.json` safely with backups and key preservation

### Phase 4.0.4C — OpenCode command/native schema bugfix (completed)
- Normalize command schema to top-level `command` (singular)
- Use `bit-*` keys (no slash) and `template` file references
- Add repair/migration for legacy Bitsentry command entries

### Phase 4.0.4D — agent.bitsentry schema normalization (completed)
- Normalize/repair `agent.bitsentry` to `mode: primary`, file prompt and ask permissions
- Remove legacy unsupported `name` field and preserve unrelated config

## Phase 4 — OpenCode Integration Foundation (final)
- Status: PASS (pending only optional/manual final validation in OpenCode)

## Phase 3 — SDR MVP (future)
- Implement real SDR workflow primitives
- Guided research flow and artifact conventions
- Validation gates for research quality

## Phase 4 — SDD MVP (future)
- Implement proposal/spec/design/tasks/apply/verify/archive workflow execution
- Improve traceability across artifacts
- Better workflow state visibility in CLI/TUI

## Phase 5 — Bitsentry Runtime Experience & Capability Optimization (next)

### Phase 5.1 — Agent Behavior Polish
- Make bitsentry behave as a true Bitsentry orchestrator (SDD/SDR/support first)
- Avoid generic coding-agent defaults and preserve explicit implementation consent

### Phase 5.2 — Commands QA & Optimization
- Validate and polish all `/bit-*` commands
- Keep structured outputs and reduce unnecessary token usage

### Phase 5.3 — Native Skills Quality Pass
- Compact and harden projected native skills (safe, actionable, bounded)

### Phase 5.4 — MCP Install Center MVP
- Controlled MCP registry/detection/configuration with safe preservation of user config

### Phase 5.5 — Token-aware Capability Loading
- Keep entrypoint/registry minimal and load only task-relevant context

### Phase 5.6 — OpenCode QA Matrix
- Validate startup, agent registration, commands, native skills, backups/restore and config preservation

## Phase 6 — Advanced profiles and model routing (future)
- Profile composition and inheritance
- Dynamic model/provider routing by task type
- Policy controls for cost, latency, and capability
- Normalize flow manifests for dynamic discovery/routing readiness
