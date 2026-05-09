---
name: sdr-structure
description: >
  Creates the technical outline for the blog post using the mandatory 
  BitSentry visual styles (prefijos //, headings, logic blocks).
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: sdr
  phase: structure
  status: declarative
---

# Skill: sdr-structure

## Purpose
Design the technical skeleton of the blog post. This skill enforces your specific visual language (`01 //`, `02 //`) and ensures that code blocks and admonitions are placed strategically for maximum impact.

## Use When
- `sdr-questions` has given a "Proceed" verdict.
- You need a blueprint before writing the final Markdown prose.

## Workflow
1.  **Headline Mapping**: Design the H1 and H2s. Use the `01 // TITLE` format for technical headers.
2.  **Narrative Arc**: Organize the synthesis points into: `Introduction`, `The Concept`, `The Technical Stack`, `Mistakes/Lessons`, and `Conclusion`.
3.  **Admonition Placement**: Identify where to insert `[!NOTE]`, `[!IMPORTANT]`, or `[!WARNING]` blocks for critical security/safety info.
4.  **Code Block Strategy**: Define which snippets (bash, python, js, sql) are essential for the "Technical Subtitles".
5.  **Visual Assets**: List where images (`![ALT](path.png)`) or tables should go to break text density.

## Outputs
### Content Blueprint (Markdown)
Must follow the **BitSentry MD Specification**:
- **Proposed Outline**: Using `## 01 //`, `## 02 //` logic.
- **Block Map**: List of where quotes, notes, and code blocks will live.
- **Excerpt Finalization**: Final Polish of the excerpt for the Frontmatter.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/structure.md` | The blog skeleton.
- **target: engram** | `sdr/{slug}/blueprint` | Final structural metadata.

**Next Recommended**: `sdr-validate`

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
