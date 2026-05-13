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

## Use When
- Recon plan exists and remains within approved scope.
- The team needs prioritization before defining tests.

## Inputs
- Recon plan outputs and scope contract.
- Approved environments and timing constraints.
- Explicitly declared in-scope subdomains/paths/assets.

## Workflow
1. Enumerate in-scope assets and trust boundaries from authorized scope.
2. Classify attack surface components and data-flow touchpoints.
3. Flag high-priority zones based on business impact and exposure.
4. Record assumptions and missing context requiring clarification.
5. Hand off prioritized map to test-plan stage.

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

## Quality Checklist
- [ ] In-scope assets are explicit and bounded.
- [ ] Priorities are justified by impact/exposure.
- [ ] No runtime requests/tooling are executed.
