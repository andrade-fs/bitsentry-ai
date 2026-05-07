---
name: sdr-synthesis
description: >
  Synthesizes research findings into a core thesis and main angles.
  The "Meaning-Making" phase of SDR.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: sdr
  phase: synthesis
  status: declarative
---

# Skill: sdr-synthesis

## Purpose
Convert the raw "Evidence Log" into a coherent narrative. This skill defines the **Thesis** of the blog post and filters out everything that doesn't add value to the reader.

## Use When
- Research is sufficient to form a clear opinion or explanation.
- You need to fill the `excerpt` and `title` (final) of the blog metadata.

## Workflow: The "Distillation Protocol"
1.  **Thesis Formation**: Define the "Long Story Short" (1 sentence).
2.  **Point Filtering**: Select the 3-5 most impactful points from the research.
3.  **Angle Refinement**: How is this different from generic tutorials? (The "BitSentry touch").
4.  **Metadata Completion**: 
    - Write the **Excerpt** (max 160 chars).
    - Refine the **Title** to be catchy but technical.
5.  **Structure Preparation**: Group evidence into logical clusters for the next phase.

## Outputs
### Synthesis Manifesto (Markdown)
- **Final Thesis**: The core message.
- **Key Takeaways**: Bullet points for the reader.
- **Draft Excerpt**: For the blog Frontmatter.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/synthesis.md` | The narrative core.
- **target: local** | `notes/sdr/{slug}/state.yaml` | Update `blog_frontmatter.excerpt` and `blog_frontmatter.title`.
- **target: engram** | `sdr/{slug}/thesis` | Compressed knowledge for future sessions.

**Next Recommended**: `sdr-questions`