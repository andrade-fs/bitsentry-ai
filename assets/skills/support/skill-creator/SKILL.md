---
name: support-skill-creator
description: >
  Generates new Bitsentry skills following Phase 3.7.x standards. 
  Ensures architectural compliance with shared contracts and result envelopes.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: support-skill-creator

## Purpose
Codify new AI capabilities into a structured, deterministic format. This skill acts as a factory for `SKILL.md` files, ensuring that new skills are immediately compatible with existing orchestrators and persistence protocols.

## Use When
- A recurring pattern is identified that isn't covered by SDD/SDR.
- You need to add a specialized support tool (e.g., a "Security Scanner" or "Database Migrator").
- You want to expand the Bitsentry ecosystem with a new `family` of skills.

## Workflow
1.  **Intent Distillation**: Define the single responsibility of the new skill.
2.  **Metadata Alignment**: 
    - Assign to a `family` (sdd, sdr, support, etc.).
    - Define `phase` and `version`.
    - List mandatory `requires` from the `_shared/` directory.
3.  **Anatomy Construction**:
    - **Purpose/Use When**: Clear triggers.
    - **Workflow**: Step-by-step logic (The "Protocol").
    - **Boundaries**: Strictly define what the skill **cannot** do.
4.  **Contract Integration**:
    - **Persistence**: Define `engram` keys and `local` paths using the slug-based hierarchy.
    - **Handoffs**: Identify valid entry/exit points to other flows.
5.  **Validation**: Run a mental "Judgment Day" to ensure the skill is declarative and lacks implementation drift.

## Outputs
### New Skill Draft (Markdown)
A complete `SKILL.md` following the Bitsentry standard, ready for deployment.

## Boundaries
- **NO Orchestration**: Do not define how the skill is called (that's for `flow.yaml`).
- **NO Logic Execution**: Only drafts the specification; does not run the code.
- **Contract Strictness**: Cannot create skills that bypass `_shared` persistence rules.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: "assets/skills/{family}/{skill-name}/SKILL.md"
  **action**: upsert
  **summary**: New skill specification file.
- **target**: engram
  **key**: "support/skill-creator/{new-skill-slug}"
  **action**: upsert
  **summary**: Traceability metadata for the new capability.

## Result Envelope
**Status**: `success | partial`

**Executive Summary**:
Skill `{skill-name}` drafted for family `{family}`. Compliant with all Bitsentry core contracts.

**Persistence Actions**:
- target: local
  key/path: "assets/skills/{family}/{skill-name}/SKILL.md"
  content summary: Complete declarative specification.

**Next Recommended**: `support/judgment-day` (To audit the new skill).

**Handoffs**:
- **to**: `support/skill-registry`
  **reason**: To register the new capability in the global index.

**Quality Checklist**:
- [ ] Frontmatter includes references to `_shared/` contracts.
- [ ] `Boundaries` section is explicit and restrictive.
- [ ] Persistence keys follow the `flow/slug/type` convention.
- [ ] `Result Envelope` format is strictly adhered to.

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.

## Quality Checklist
- [ ] Required heading present for contract compliance.
- [ ] Guidance remains declarative.
- [ ] No runtime behavior changes introduced.
