---
name: sdd-explore
description: >
  Systematic exploration of existing code, architecture, and context. 
  Identifies affected areas, dependencies, and risks before proposing changes.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdd
  phase: explore
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdd-explore

## Purpose
Investigate the current state of the system to resolve the `unknowns` identified in `sdd-init`. This skill maps the "technical surface" of the change, ensuring the subsequent proposal is grounded in reality.

## Use When
- `sdd-init` is complete and a `change-slug` exists.
- You need to identify which files, modules, or APIs will be impacted.
- You need to validate assumptions made during initialization.

## Inputs
### Required
- **Session State**: `sdd/{slug}/state.yaml` (from init).
- **Initialization Brief**: `sdd/{slug}/init.md`.
- **Repo Access**: Ability to list files, read signatures, or search symbols.

### Optional
- Relevant documentation or existing specs.
- Specific files suggested by the user.

## Workflow
1.  **Unknowns Triage**: Prioritize the `unknowns` list from the session state.
2.  **Surface Mapping**: 
    - Identify core files/modules affected.
    - Trace dependencies (what calls this? what does this call?).
    - Locate relevant tests and documentation.
3.  **Architecture Review**: Observe existing patterns (naming conventions, error handling, DI patterns) to ensure the change remains idiomatic.
4.  **Risk Discovery**: Detect potential side effects (e.g., "Changing this shared utility will impact 15 other modules").
5.  **Evidence Collection**: Record specific line ranges or symbols that act as "anchor points" for the change.

## Outputs
### Exploration Report (Markdown)
Must include:
- **Impact Map**: List of files/modules with "High/Medium/Low" impact ratings.
- **Resolved Unknowns**: Answers to the questions posed in `sdd-init`.
- **New Risks**: Technical debt or complexities discovered during recon.
- **Pattern Matching**: Description of existing code style/patterns to follow.

## Boundaries
- **NO Implementation**: Do not write new code or refactor existing code.
- **NO Final Design**: Do not decide *how* to fix it yet; only document *where* and *what* is there.
- **NO File Edits**: This is a read-only phase.
- **Minimal Context**: Do not read unrelated modules; stay within the "blast radius" of the intent.

## Persistence Actions
Compliant with _shared/persistence-contract.md and _shared/engram-convention.md

- **target: local** | `sdd/{slug}/explore.md` | Detailed reconnaissance report.
- **target: local** | `sdd/{slug}/state.yaml` | Update `scope.in` and `scope.unknown` based on findings.
- **target: engram** | `sdd/{slug}/impact_map` | Structural map of affected areas.

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Exploration of `{slug}` complete. Identified `{N}` affected files and resolved `{M}` unknowns. High-impact areas: `{list}`.

**Detailed Report**:
- **Findings**: Summary of the system's current behavior regarding the goal.
- **Impact List**: Structured list of target files/symbols.
- **Assumptions Validated**: Which hypotheses from `init` were confirmed/refuted.

**Artifacts**:
- `sdd/{slug}/explore.md`

**Persistence Actions**:
- target: local
  key/path: `sdd/{slug}/state.yaml`
  content summary: Moving items from `unknown` to `in-scope` or `out-of-scope`.

**Next Recommended**: `sdd-propose`

**Handoffs**:
- **To SDR**: If the codebase is so complex/obfuscated that it requires a separate research flow to understand.
- **To Blocked**: If the target area is missing, inaccessible, or completely different from the `init` description.

**Risks / Gaps**:
- List any "New Unknowns" discovered that require a specific proposal to solve.

## Quality Checklist
- [ ] Every `unknown` from `sdd-init` was addressed.
- [ ] Impacted files are listed with reasons for inclusion.
- [ ] Side effects on external dependencies are identified.
- [ ] Existing architectural patterns are documented.
- [ ] No code was written or modified.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
