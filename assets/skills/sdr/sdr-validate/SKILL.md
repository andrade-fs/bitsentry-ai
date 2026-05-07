---
name: sdr-validate
description: >
  Final quality gate for research content. Enforces the BitSentry Semáforo 
  Check and ensures visual consistency with the blog UI.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdr
  phase: validate
  status: declarative
---

# Skill: sdr-validate

## Purpose
Acting as the final editorial gate. This skill performs a cold, adversarial review of the content structure and technical accuracy. It ensures the post is not only "correct" but also visually compatible with the macOS-styled blog frontend.

## Use When
- `sdr-structure` is complete.
- You have a near-final draft of the research and its technical snippets.

## Workflow: The "Semáforo Check"
1.  **BitSentry Semáforo Protocol**:
    - **GREEN (Pass)**: Content is original, technical blocks are correct, and visual style is consistent.
    - **YELLOW (Review)**: Missing a technical subtitle (0x //), weak edge cases, or generic excerpt.
    - **RED (Block)**: Content is too cliché, lacks the BitSentry "angle", or fails Markdown rigor.
2.  **MD Rigor Audit**:
    - Verify `## 0X // TITLE` pattern.
    - Check for `[!NOTE]`, `[!IMPORTANT]`, or `[!WARNING]` placement.
    - Ensure code blocks have the correct language tags (`python`, `javascript`, `bash`, `sql`).
3.  **Density & Balance**: Check if there's too much text without images or code blocks. Enforce `text-pretty` logic (no orphans/widows in reasoning).
4.  **Security/Sanity Check**: Ensure no hardcoded API keys exist in code blocks and that security alerts are marked with `[!WARNING]`.

## Outputs
### Validation Report (Markdown)
Must include:
- **Verdict**: `GREEN | YELLOW | RED`.
- **Styling Score**: Check of macOS-style consistency (buttons, tags, copy-button readiness).
- **Required Fixes**: List of adjustments to headers or admonitions.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/validation.md` | The final review manifesto.
- **target: engram** | `sdr/{slug}/verdict` | Quality historical data.

**Next Recommended**: `sdr-archive`