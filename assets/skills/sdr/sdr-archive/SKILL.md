---
name: sdr-archive
description: >
  Generates the final, blog-ready Markdown file with strict Frontmatter 
  and BitSentry technical styling.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdr
  phase: archive
  status: declarative
---

# Skill: sdr-archive

## Purpose
The final assembly line. This skill produces the definitive `.md` file that will be rendered by the blog. It must merge the research, synthesis, and structure into a single document that respects every line of your CSS and React components.

## Use When
- `sdr-validate` has issued a GREEN or YELLOW verdict.
- You are ready to close the research session and publish.

## Workflow: The "Final Assembly Protocol"
1.  **Frontmatter Construction**: Generate the YAML block exactly as specified:
    - `title`, `date`, `excerpt`, `coverImage`, `category`, `authorName`, `authorAvatar`.
2.  **Narrative Assembly**:
    - Start with a `> Blockquote` for the intro hook.
    - Use `## 01 // TITLE` for every major technical section.
    - Insert `![ALT](image.png)` with correct relative paths.
3.  **Code Block Hardening**:
    - Ensure all blocks use three backticks + language.
    - Verify macOS-style consistency: no extra line breaks at the end of blocks.
4.  **Admonition Styling**:
    - Convert notes to `> [!NOTE]`, `> [!IMPORTANT]`, or `> [!WARNING]` blocks.
5.  **Final Cleanup**: Remove all internal research notes and metadata; output only the pure blog post.

## Outputs
### Final Blog Post (Markdown)
Must be a **Single File** following this exact structure:
```markdown
---
title: "..."
date: "YYYY-MM-DD"
excerpt: "..."
category: "..."
# ... other FM fields
---

> Intro quote...

## 01 // TECHNICAL HEADER
Content...

> [!IMPORTANT]
> Critical info.

```bash
# code...

## Boundaries
- **NO Content Changes**: Do not invent new research at this stage.
- **Strict MD ONLY**: No HTML tags unless explicitly required for edge cases.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/final-post.md` | **The Master Artifact**.
- **target: local** | `notes/sdr/{slug}/state.yaml` | Mark as `status: archived`.
- **target: engram** | `sdr/{slug}/final_archive` | Complete lineage with observation IDs.

**Next Recommended**: `none`

**Handoffs**:
- **To Issue Creation**: If the post identifies a bug or feature to be implemented in the lab.