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

## Phase 3 — SDR MVP (future)
- Implement real SDR workflow primitives
- Guided research flow and artifact conventions
- Validation gates for research quality

## Phase 4 — SDD MVP (future)
- Implement proposal/spec/design/tasks/apply/verify/archive workflow execution
- Improve traceability across artifacts
- Better workflow state visibility in CLI/TUI

## Phase 5 — Red Team / Bug Bounty workflows (future)
- Structured, authorized workflow templates
- Reporting pipelines and evidence tracking
- Safer operational guardrails in execution

## Phase 6 — Advanced profiles and model routing (future)
- Profile composition and inheritance
- Dynamic model/provider routing by task type
- Policy controls for cost, latency, and capability
- Normalize flow manifests for dynamic discovery/routing readiness
