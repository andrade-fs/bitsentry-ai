---
name: sdr-questions
description: >
  Evaluates the research value and the "BitSentry Angle". 
  The primary quality gate before content creation.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: sdr
  phase: questions
  status: declarative
---

# Skill: sdr-questions

## Purpose
Apply a "Value Stress Test" to the synthesis. This skill ensures that the research isn't just another technical cliché, but a useful, actionable, and unique piece of content for the BitSentry blog.

## Use When
- `sdr-synthesis` is complete.
- You need to decide the final format (Deep Dive, Quick Note, or Discard).

## Workflow
1.  **Originality Check**: Is this common knowledge or does it provide a unique perspective?
2.  **Actionability**: Can a reader take this and apply it to their HomeLab/Workplace immediately?
3.  **Cliché Audit**: Detect and remove "AI-generic" filler or repetitive technical advice.
4.  **Format Selection**: Based on depth, choose: `Tutorial`, `Opinion`, `Case Study`, or `Quick Tip`.
5.  **Target Audience**: Confirm if it's for `Junior`, `Senior`, or `Decision Maker`.

## Outputs
### Evaluation Report (Markdown)
Must answer:
- **Why does this matter?**
- **Is it reproducible?**
- **What is the unique BitSentry angle?**
- **Decision**: `Proceed to Structure` | `Pivot` | `Discard`.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/evaluation.md` | Critical assessment report.
- **target: engram** | `sdr/{slug}/quality_gate` | Reasoning for publication.

**Next Recommended**: `sdr-structure`

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Boundaries
This section is required by contract. Keep behavior declarative and avoid runtime orchestration changes.

## Result Envelope
Return status, executive summary, artifacts, next recommended, risks, and skill resolution when applicable.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.

## Quality Checklist
- [ ] Required heading present for contract compliance.
- [ ] Guidance remains declarative.
- [ ] No runtime behavior changes introduced.
