---
name: security-scope
description: Define source review scope and non-goals for AppSec analysis.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: scope
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: security-scope

## Purpose
Set precise review boundaries to avoid pentest drift and keep analysis source-centric.

## Use When
- Initialization has confirmed source-security-review route.
- The repository surface is broad and needs explicit in/out scope.

## Inputs
- Repository modules, dependency manifests, and non-secret configs.
- User constraints and risk priorities.

## Workflow
1. Enumerate in-scope source/config/dependency surfaces.
2. Record non-goals: exploit execution, runtime probing, external targets.
3. Define risk categories and severity rubric for findings.
4. List evidence expectations for each finding.
5. Output scoped plan for mapping stage.

## Outputs
- `scope.md` with in-scope areas, non-goals, and evidence rubric.

## Boundaries
- NO secrets extraction or handling.
- NO `.env` reads.
- NO destructive or mutating actions.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/scope.md`
  **action**: upsert
  **summary**: Source security scope definition.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Review scope defined with explicit non-goals and risk rubric.

**Next Recommended**: `security-map`

## Handoffs
- to `security/security-map` for threat-surface mapping.

## Quality Checklist
- [ ] Scope excludes live pentest/runtime activity.
- [ ] Includes dependency and config review criteria.
- [ ] Non-goals are explicit and testable.
