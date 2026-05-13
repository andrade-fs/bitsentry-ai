---
name: web-assessment-map
description: Map authorized web target surface declaratively without runtime interaction.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: map
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: web-assessment-map

## Purpose
Produce a structured map of in-scope web surface, trust boundaries, and priority areas from declared target context.

## Assessment Session Context
- Intent and Scope remain the source of truth for Mapping decisions.
- Discovery and Mapping are declarative: no crawler/scanner real and no live requests.
- Surface Ranking must be explicit and tied to in-scope assets only.
- Execution mode respected: planning_only / dry_run / execute_approved / retest.

## Tooling Policy / Command Safety
This stage uses only authorized evidence and follows canonical safety in `security/web-assessment-requests`.
- no execution by default
- authorized target required
- exact scope required
- allowed tools required
- prohibited actions required
- no out-of-scope scanning

Semantics:
- web-assessment-map: mapea desde evidencia autorizada, no escanea fuera de scope.

## Use When
- Recon plan exists and remains within approved scope.
- The team needs prioritization before defining tests.

## Inputs
- Recon plan outputs and scope contract.
- Approved environments and timing constraints.
- Explicitly declared in-scope subdomains/paths/assets.

## Workflow
1. Discovery: enumerate in-scope assets and trust boundaries from authorized scope.
2. Mapping: classify surface components and data-flow touchpoints.
3. Surface Ranking: flag high-priority zones by impact/exposure.
4. Risk Hypotheses: document what should be validated later.
5. Record assumptions and ask technical clarifying questions when detail is missing.
6. Hand off prioritized map to test-plan stage.

## Outputs
- `web-assessment-map.md` with asset map, trust boundaries, priorities, and assumptions.

## Boundaries
- Authorization required before live target interaction.
- Exact scope required.
- Explicit permission before requests.
- No external target testing without authorization.
- No exploit execution by default.
- No destructive actions.
- No DoS/load testing.
- No credential attacks.
- No secrets exposure.
- No MCP credential mutation.
- No runtime flow execution.
- No autonomous mode.
- OpenCode-first.
- CLI debug/plumbing only.
- `agent.bitsentry.permission.edit = deny`.

## Persistence Actions
- **target**: local
  **key/path**: `web-assessment/{slug}/map.md`
  **action**: upsert
  **summary**: Declarative target map and prioritization.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Web surface map prepared with prioritized zones and explicit assumptions.

**Next Recommended**: `web-assessment-test-plan`

## Handoffs
- to `security/web-assessment-test-plan` with prioritized map and assumptions.
- Command safety canonical source: `security/web-assessment-requests`.

## Quality Checklist
- [ ] In-scope assets are explicit and bounded.
- [ ] Priorities are justified by impact/exposure.
- [ ] No runtime requests/tooling are executed.
