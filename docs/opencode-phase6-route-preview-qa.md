# OpenCode Phase 6 Route Preview QA

Status: **PASS WITH NOTES**

## Phase 6.4 MCP Readiness & Validation QA

Status: **PASS**

- Confirmed MCP readiness model exists and is testable with fields:
  - `status`
  - `detected_evidence`
  - `blockers`
  - `manual_hints`
  - `safe_usable`
- Confirmed Engram/Context7 readiness now differentiates configured/detected/missing/manual-step-needed paths.
- Confirmed modeled-only and unsupported MCP statuses are explicit (`modeled_only`, `not_implemented`).
- Confirmed TUI Install/Setup shows MCP readiness summary in review and done/control panel steps.
- Confirmed guardrails preserved:
  - no `.env`/secret access
  - no MCP credential mutation
  - no autonomous runtime activation
  - no flow execution
  - no skill execution
  - no change to `agent.bitsentry.permission.edit=deny`

## Phase 6.3 TUI Install Everything Mode QA

Status: **PASS**

- Confirmed Install/Setup is now positioned as OpenCode installer/control panel (not CLI-first UX).
- Confirmed primary install path is simplified to:
  1) Install Everything
  2) Install Bitsentry Pack
  3) Update/Reinstall Bitsentry Pack (default when already installed)
- Confirmed normal installation no longer forces granular flow/skill/preset selection in main path.
- Confirmed review/final screens clearly state:
  - what it will do
  - what it will not do
  - detected state
  - manual action notes
- Confirmed final summary includes:
  - OpenCode detected/not detected
  - OpenCode config root
  - Bitsentry pack root
  - pack status
  - native integration status
  - backup path(s)
  - manual steps/notes
  - OpenCode test prompt
- Confirmed guardrails preserved:
  - OpenCode-only target
  - no MCP credential mutation
  - no autonomous runtime activation
  - no flow execution
  - no skill execution
  - no change to `agent.bitsentry.permission.edit=deny`

## Phase 6.2 Capability Selection Preview QA

Status: **PASS**

- Added route preview capability fields: `primary_skills`, `secondary_skills`, `deferred_skills`, `primary_roles`, `secondary_roles`, `capability_reason`, `capability_gates`.
- Confirmed OpenCode-first contract remains primary UX; CLI `route decide` remains debug/plumbing parity only.
- Confirmed preview remains non-mutating and non-executing:
  - no skill execution
  - no flow execution
  - no Engram/OpenSpec persistence
  - no code edits in preview (`agent.bitsentry.permission.edit=deny` boundary preserved)
- Confirmed mandatory preview gates from 6.1 remain intact:
  - `no_edits_in_preview`
  - `no_persistence_in_preview`
  - `no_flow_execution_in_preview`

## Critical Re-test after 6.1E.1

Status: **FAIL**

- Prompt used: "En la landing quiero evitar que parezca que bitsentry-ai es una CLI para usar directamente"
- Observed behavior (unexpected): no visible Route Decision Preview before discovery, discovery executed first, `Thinking:` surfaced, and direct edit attempt on `src/components/Hero.astro` without confirmation.
- Conclusion: prompt-only enforcement is insufficient for the current model/runtime behavior.
- Corrective action (Phase 6.1F): enforce non-mutation at permission layer (`agent.bitsentry.permission.edit=deny`) and keep bitsentry as orchestrator-only.

## Follow-up QA learning before 6.1G

- The agent correctly adjusted routing criteria and accepted that the landing messaging request can be compact SDD (not direct_answer) when product narrative/perception is affected.
- Remaining gap: enforce the formal Route Decision Preview envelope as the first visible response before discovery/analysis.
- Corrective action (Phase 6.1G): frontload non-negotiable first-response contract at the top of the prompt.

## Phase 6.1E.1 corrective note

- Root cause identified: conflicting prompt wording allowed interpreting discovery before visible route preview.
- Fix applied: prompt now requires visible Route Decision Preview BEFORE non-trivial discovery, and allows bounded discovery only AFTER preview when `requires_bounded_discovery=true`.
- Additional guardrail: prompt explicitly forbids exposing `Thinking:`/hidden reasoning and limits output to preview/findings/envelopes/user-facing decisions.

## Goal

Validate that the native OpenCode `bitsentry` agent treats route decision preview as an OpenCode-first capability, not as a direct CLI workflow.

## Latest Landing QA Result

- Scenario: landing repo requested less CLI-centric positioning.
- Observed: agent performed safe bounded read-only discovery, understood CLI-centric issue, did not edit files, did not persist memory, did not execute flows, and asked confirmation before mutations.
- Note: in that run, the visible Route Decision Preview was not shown before discovery.
- Corrective action (Phase 6.1E): enforce visible Route Decision Preview before non-trivial discovery/analysis/planning requests.

## Global expectations

- The agent must not tell the user to run `bitsentry-ai route decide` manually.
- The agent must show a route decision before activating SDD, SDR, or Support.
- The agent must show `matched_signals` when available.
- The agent must not edit files during preview.
- The agent must not persist memory during preview.
- The agent must not execute flows during preview.
- The agent must ask for confirmation when the selected path requires it.

## Case 1 — Frontend / TUI change

Prompt:

```text
@bitsentry Quiero mejorar el wizard del TUI
```

Expected:

- `matched_intent: frontend-ux-change`
- `decision: use_flow_sdd`
- `recommended_flow: sdd`
- `matched_signals`: present if available
- `requires_confirmation: true`
- No edits before confirmation.

## Case 2 — Bug/security overlap

Prompt:

```text
@bitsentry Hay un bug de seguridad en export-preview
```

Expected:

- `matched_intent: security-review`
- Security wins over bug.
- `matched_signals` includes `seguridad`/`security` if available.
- No edits before confirmation.

## Case 3 — Direct answer

Prompt:

```text
@bitsentry Qué es un embedding en RAG?
```

Expected:

- `matched_intent: direct-answer`
- `decision: direct_answer`
- No SDD.
- No Support.
- No persistence.
- No confirmation unless needed.

## Case 4 — CLI boundary

Prompt:

```text
@bitsentry Cómo decido si esto va por SDD o Support?
```

Expected:

- The agent explains route decision preview.
- The agent does not tell the user to run `bitsentry-ai route decide`.
- The agent may mention the CLI only as internal/debug plumbing if necessary.
