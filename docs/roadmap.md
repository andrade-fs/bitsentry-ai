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

## Phase 5 — OpenCode Agent Orchestration Polish (current)

- Focus: polish native OpenCode `bitsentry` behavior for SDD/SDR/support/tool guidance.
- Boundary: BitsentryAI remains installer/projector/validator/pack manager, not a standalone runtime executor.
- Boundary: no new agent targets in this phase (OpenCode only).
- Boundary: no MCP credential mutation by default.

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

## Phase 6 — Intent-Aware Orchestration & Role Packs MVP (completed)
- Status: **PASS WITH NOTES**
- Added Intent Decision Contract in native `bitsentry` prompt (direct answer vs skills vs roles vs flows)
- Added declarative intent registry (`assets/intents/*.yaml`)
- Added specialist role pack (`assets/roles/*.md`)
- Projected roles into OpenCode as subagents while keeping `bitsentry` as primary
- Preserved safety boundaries: OpenCode-only target, no autonomous runtime, no MCP credential mutation

Non-blocking note:
- `export-preview` / report YAML summary does not yet expose intents/roles explicitly.

### Phase 6.1 — Route Decision Preview MVP (completed)
- Added lightweight read-only intent classifier and route decision preview envelope.
- Kept behavior non-mutating and OpenCode-only.

### Phase 6.1B — OpenCode Agent Route Preview Wiring (completed)
- Updated native `bitsentry` prompt contract to frame route preview as OpenCode-native agent behavior.
- Added `matched_signals` to route preview output contract in prompt guidance.
- Kept `bitsentry-ai route decide` as debug/plumbing parity, not primary end-user UX.

### Phase 6.1C — OpenCode Prompt Contract Snapshot QA (completed)
- Added prompt anti-regression tests for OpenCode-first route preview semantics.
- Added guardrails so prompt does not present CLI `route decide` as primary UX.
- Added manual QA checklist for OpenCode route preview contract validation.

### Phase 6.1E — Visible Route Preview Enforcement (completed)
- Enforced visible Route Decision Preview before non-trivial discovery/analysis/planning/recommendation requests.
- Enforced required preview gates (`no_edits_in_preview`, `no_persistence_in_preview`, `no_flow_execution_in_preview`) in prompt contract.
- Added anti-regression tests and documented landing QA outcome as PASS WITH NOTES.

### Phase 6.1E.1 — Route Preview Prompt Conflict Cleanup (completed)
- Removed contradictory prompt wording that allowed interpreting discovery-before-preview.
- Enforced explicit ordering: preview first, bounded discovery second (only when required), update decision only when findings change.
- Reinforced no-`Thinking:` exposure rule and added anti-regression tests.

### Phase 6.1F — Bitsentry Orchestrator No-Edit Guardrail (completed)
- Enforced `agent.bitsentry.permission.edit = deny` in native OpenCode integration/repair path.
- Hardened prompt contract: bitsentry is orchestrator-only and non-mutating by default.
- Captured critical landing re-test FAIL and applied permission-layer mitigation.

### Phase 6.1G — Prompt Priority / Frontload Contract (completed)
- Reordered prompt to frontload non-negotiable first-response Route Decision Preview contract before secondary matrices/rules.
- Added formal first-response preview template and mandatory gates.
- Captured QA routing learning: small copy/product-messaging/UX changes may still route to compact SDD when core narrative/perception is affected.

### Phase 6.2 — Capability Selection Preview (completed)
- Extended route decision envelope with capability preview fields (`primary_*`, `secondary_*`, `deferred_skills`, `capability_reason`, `capability_gates`).
- Resolved capability preview from declarative assets (`assets/intents`, `assets/roles`, `assets/skills`, `assets/flows`) through existing discovery.
- Updated OpenCode native prompt contract to require visible Route + Capability Decision Preview before deep discovery/planning/proposals.
- Preserved 6.1 preview gates and non-mutating orchestrator boundaries (`edit=deny`, no flow/skill execution, no persistence in preview).

### Phase 6.3 — TUI Install Everything Mode (completed)
- Reframed Install/Setup as OpenCode installer/control panel with a simple main path:
  1) Install Everything
  2) Install Bitsentry Pack
  3) Update/Reinstall Bitsentry Pack (default when already installed)
- Removed granular preset/skill/flow/MCP toggles from primary path and relegated that granularity to advanced/debug pathways.
- Improved factual copy for detected state, actions, guardrails, manual steps, and final outcome summary.
- Expanded final screen with OpenCode detection/config roots, pack/native integration status, backup paths, and test prompt.
- Preserved core safety boundaries: OpenCode-only target, no flow/skill execution, no autonomous runtime, no MCP credential mutation, no change to `agent.bitsentry.permission.edit=deny` contract.

Non-blocking Phase 6.4 note:
- Consider finer-grained capability ranking heuristics and confidence weighting per flow stage.

### Phase 6.4 — MCP Readiness & Validation (completed)
- Added a normalized MCP readiness model (`status`, evidence, blockers, manual hints, `safe_usable`).
- Enriched Engram + Context7 runtime summaries with readiness and explicit manual-step states.
- Added readiness state distinctions: configured/detected/missing/modeled-only/not-implemented/manual-step-needed.
- Integrated readiness summaries in TUI install review and final control panel screens.
- Added regression tests for readiness model, MCP registry statuses, TUI rendering, and no-mutation guardrail defaults.

Minimal 6.5 roadmap note:
- Phase 6.5 should focus on deeper per-MCP validation adapters while preserving no-secret-mutation boundaries.

### Phase 6.5 — OpenCode MCP Config Preview (completed)
- Added a preview-only `opencode.json` MCP config model (no write path) with explicit safety contract fields.
- Modeled scenarios: missing/invalid/readable without MCP/readable with MCP plus readiness summary.
- Preserved existing top-level and MCP keys in preview output; no overwrite/replace behavior in preview phase.
- Added manual-step guidance for sensitive setup gaps without writing credentials/tokens.
- Integrated preview contract into TUI review + done/control-panel screens with **PREVIEW ONLY** marker.
- Added tests for no-write preview behavior, confirmation+backup gating contract, config-shape scenarios, and TUI rendering.

### Phase 6.6 — Release Stabilization (completed)
- Consolidated Phase 6 closure artifacts in a single release stabilization document.
- Aligned README + roadmap + QA docs around OpenCode-first positioning and safety boundaries.
- Finalized release demo flow: TUI Install Everything → MCP Readiness → MCP Config Preview → OpenCode Route+Capability Preview.
- Preserved all guardrails (no secrets/no `.env`/no MCP credential mutation/no autonomous runtime/no flow/skill execution in preview path).

### Phase 6 Status Snapshot (6.1 → 6.6)
- 6.1 Route Decision Preview: completed
- 6.2 Capability Selection Preview: completed
- 6.3 TUI Install Everything Mode: completed
- 6.4 MCP Readiness & Validation: completed
- 6.5 OpenCode MCP Config Preview: completed
- 6.6 Release Stabilization: completed

### Next Step
- Prepare and validate release candidate tag for Phase 6 closure (no new major feature scope).

## Phase 7 — Advanced profiles and model routing (current)

### Phase 7.0 — Compact Direct Mode & Agent Token Efficiency (completed)
- Added prompt-level contract to use Compact Direct Mode for direct/atomic/concrete/already-intentional requests.
- Kept Full Route Decision Preview mandatory for open/ambiguous/exploratory/planning/architecture/security/multi-file/sensitive requests.
- Added explicit compact+confirmation behavior for direct but sensitive/remote/destructive/ambiguous requests.
- Added anti-regression tests for Compact Direct Mode vs Full Preview triggers.
- Preserved non-mutation and security guardrails (no secrets/.env access, no MCP credential mutation, no runtime/flow execution in preview, no preview persistence/edits, OpenCode-first, CLI debug/plumbing only).

### Phase 7.1 — Source Security Review Flow MVP (completed)
- Added new declarative flow manifest: `assets/flows/source-security-review.yaml`.
- Added security skill pack with 6 structural skills: init, scope, map, review, findings, report.
- Wired `security-review` intent to recommend `source-security-review`.
- Updated OpenCode bitsentry prompt matrix to recognize source-security-review instead of the prior "no pentest flow yet" copy.
- Preserved guardrails: read-only first, no `.env`/secrets, no exploit execution, no external target testing, no destructive actions, no MCP credential mutation, no runtime flow execution, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing parity, `agent.bitsentry.permission.edit=deny`.
- Updated discovery/manifest/skills/route/prompt/workflow visibility tests for the new flow and security pack.

### Phase 7.2 — Security Skill Pack MVP (completed)
- Added 8 specialized support security skills under `assets/skills/security/*/SKILL.md`:
  - auth-security-review, jwt-review, graphql-security-review, xss-review
  - file-upload-review, ssrf-review, secrets-review, dependency-risk-review
- Kept flow contract stable: no flow ID change and no stage changes/additions.
- Linked the 8 skills as support references/checklists in `assets/flows/source-security-review.yaml` (`requires.support_skills`) without turning them into stages.
- Updated docs and manifest-contract tests to enforce flow stability and support-skill linkage.
- Preserved all security guardrails (read-only first, no `.env`/secrets, no exploit/live target testing, no destructive actions, no MCP credential mutation, no runtime flow execution, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing only, `agent.bitsentry.permission.edit=deny`).

### Phase 7.3 — Security Report Markdown (completed)
- Defined strict Markdown report contract in `security-report` with exact required sections:
  - Title, Executive Summary, Scope, Methodology, Repository / Application Context, Risk Summary, Findings, Evidence, Remediation Plan, Verification Steps, Assumptions and Limitations, Next Steps, Appendix.
- Defined strict minimum finding contract in `security-findings` with exact required tokens and enums:
  - fields: ID, Title, Severity, Confidence, Category, Affected files, Affected component, Evidence, Impact, Likelihood, Remediation, Verification, References / Notes.
  - Severity: Critical, High, Medium, Low, Informational.
  - Confidence: High, Medium, Low.
- Added static anti-regression tests for:
  - mandatory finding fields,
  - severity/confidence enums,
  - mandatory report sections,
  - explicit handoff `security-findings -> security-report`.
- Kept `source-security-review` stable (same flow ID and stage IDs/order).

### Phase 7.4A — Findings Taxonomy + Calibration Contract (completed)
- Added official findings taxonomy enum to `security-findings` for source review normalization:
  - Authentication, Authorization, Session Management, Input Validation, Injection, Cross-Site Scripting, Server-Side Request Forgery, File Upload, Secrets Exposure, Cryptography, Dependency Risk, GraphQL Security, Business Logic, Configuration, Logging / Monitoring, Error Handling, Data Exposure, Supply Chain, Informational.
- Added explicit Severity calibration anchors via **Impact × Likelihood** (Critical/High/Medium/Low/Informational).
- Added explicit Confidence calibration anchors by evidence quality (High/Medium/Low).
- Added formal skill→category mapping contract (primary/secondary) for specialized security review skills.
- Added explicit deduplication rules, evidence grouping rules, and assumptions/limitations rules.
- Hardened handoff `security-findings -> security-report` with required payload semantics.
- Updated `security-report` to explicitly consume taxonomy and calibrated fields in **Risk Summary** and **Findings**.
- Added strict anchor/token tests for official categories, enums, Impact × Likelihood token, dedup/evidence/assumptions rules, mapping contract, handoff, and report taxonomy consumption.
- Kept flow stability: `source-security-review` ID unchanged, no stage ID/order changes, no runtime flow execution.

### Phase 7.4B — Findings/Report Examples + Golden Fixtures (completed)
- Added safe synthetic documentation contract artifacts under `assets/docs/security/`:
  - examples: `findings-example.md`, `report-example.md`
  - fixtures: `findings-golden.md`, `report-golden.md`
  - reference guide: `assets/docs/security/README.md`
- Extended `internal/capabilities/security_contracts_test.go` with static checks for:
  - docs artifact existence,
  - minimum finding token coverage,
  - severity coverage in findings golden fixture (Critical/High/Medium/Low/Informational),
  - confidence enum presence (High/Medium/Low),
  - taxonomy usage and dedup/evidence/assumptions anchors,
  - required report section order in `report-golden`.
- Preserved flow and runtime boundaries: no `source-security-review` flow ID/stage changes, no runtime flow execution, no route/schema/registry mutations.
- Preserved security guardrails in fixtures/examples: read-only-first, no `.env`/secrets, no exploit or live-target testing, no destructive actions, no MCP credential mutation, no autonomous mode, no edits by default, OpenCode-first, CLI debug/plumbing only, `agent.bitsentry.permission.edit=deny`.

### Phase 7.4C — Source Security Review QA/Manual Smoke Protocol (completed)
- Added `docs/opencode-phase7-security-source-review-qa.md` as the formal QA protocol for OpenCode source-security-review.
- Defined 7 mandatory smoke cases with strict per-case contract: prompt, expected behavior, forbidden behavior, minimum evidence, PASS/FAIL criteria.
- Added explicit hard fail conditions (`.env` access, secrets exposure, offensive tooling, external target testing, default code mutation, skipping sensitive confirmations, fabricated findings, missing assumptions/limitations).
- Added evidence capture rules, guardrails matrix, known acceptable PASS WITH NOTES cases, and final go/no-go checklist before 7.5.
- Kept strict scope and stability boundaries: docs-only changes; no flow/runtime/schema/registry/agent/target mutations.

- Next Phase 7 tracks:
- 7.5 Web Assessment Flow (next, gated by 7.4C go/no-go checklist)
- Profile composition and inheritance
- Dynamic model/provider routing by task type
- Policy controls for cost, latency, and capability
- Normalize flow manifests for dynamic discovery/routing readiness
