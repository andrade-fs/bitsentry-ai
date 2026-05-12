---
name: security-map
description: Map source-level attack surface and trust boundaries.
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

# Skill: security-map

## Purpose
Build a source-centric threat surface map from code structure, dependencies, and configuration.

## Use When
- Scope is defined and review can start structured mapping.
- Risk hotspots are unknown and need prioritization.

## Inputs
- Scoped modules and paths.
- Dependency manifests and lockfiles.
- Non-secret config files.

## Workflow
1. Identify trust boundaries and untrusted inputs.
2. Map auth, authorization, data handling, and integration touchpoints.
3. Highlight high-risk dependency and supply-chain zones.
4. Produce hotspot map with review priority ordering.
5. Prepare map for deep review stage.

## Outputs
- `map.md` with threat surface map and prioritized hotspots.

## Boundaries
- NO exploit attempts.
- NO runtime scans.
- NO external endpoint interaction.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/map.md`
  **action**: upsert
  **summary**: Threat surface and hotspot map.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Threat surface mapped and review priorities defined.

**Next Recommended**: `security-review`

## Handoffs
- to `security/security-review` for bounded read-only analysis.

## Quality Checklist
- [ ] Trust boundaries are explicit.
- [ ] Priority hotspots are justified.
- [ ] Mapping remains source/code/config/dependency focused.
