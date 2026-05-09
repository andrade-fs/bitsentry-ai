---
name: support-skill-registry
description: >
  Generates a compact, project-local index of all available skills, 
  contracts, and flows. Serves as the primary manifest for orchestrators.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
---

# Skill: support-skill-registry

## Purpose
Synthesize all distributed `SKILL.md` and `flow.yaml` files into a single, high-density reference document. This allows orchestrators to load the entire capability map without scanning the full filesystem in every step.

## Use When
- A new skill has been created via `support-skill-creator`.
- The folder structure or family definitions have changed.
- The orchestrator needs a fresh "Compact Map" to resolve dependencies or handoffs.

## Workflow
1.  **System Scan**: Identify all directories under `assets/skills/` and `assets/flows/`.
2.  **Metadata Extraction**: 
    - Parse `family`, `name`, `phase`, and `status` from frontmatter.
    - Extract `Purpose`, `Use When`, and `Handoffs` from the body.
3.  **Dependency Mapping**: List which shared contracts (`_shared/`) are required by each family.
4.  **Compact Synthesis**: Format the data into a hierarchical table or list designed for high-token-efficiency (compactness).
5.  **Stale Check**: Identify any skills that are missing mandatory sections or fail result-envelope compliance.

## Outputs
### Compact Skill Registry (Markdown)
A single document (`skill-registry.md`) containing:
- **Registry Version & Timestamp**.
- **Shared Contracts Index**: Quick links to `_shared/` protocols.
- **Skill Map by Family**: `{ID} | {Purpose} | {Handoffs}`.
- **Flow Inventory**: List of available multi-stage flows (SDD, SDR, etc.).

## Boundaries
- **NO Logic Execution**: Only catalogs the definitions; does not validate the internal logic of the skills.
- **No Orchestration**: Does not decide *which* skill to run, only lists what is *available*.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: ".bitsentry-ai/skill-registry.md"
  **action**: upsert
  **summary**: Comprehensive index of Bitsentry AI capabilities.
- **target**: engram
  **key**: "support/skill-registry/latest"
  **action**: upsert
  **summary**: Global snapshot of the current skill ecosystem.

## Result Envelope
**Status**: `success | partial`

**Executive Summary**:
Registry refreshed. `{Total_Skills}` skills and `{Total_Flows}` flows indexed across `{N}` families.

**Persistence Actions**:
- target: local
  key/path: ".bitsentry-ai/skill-registry.md"
  content summary: Compact map of IDs, purposes, and handoffs.

**Next Recommended**: `none` (The orchestrator will consume this on next boot).

**Handoffs**:
- **to**: `main-orchestrator`
  **reason**: Context update; new capabilities are now "visible" to the system.

**Quality Checklist**:
- [ ] Every skill has a unique ID and Family.
- [ ] Handoff targets exist in the registry (no broken links).
- [ ] Output is optimized for minimal token consumption.

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.

## Quality Checklist
- [ ] Required heading present for contract compliance.
- [ ] Guidance remains declarative.
- [ ] No runtime behavior changes introduced.
