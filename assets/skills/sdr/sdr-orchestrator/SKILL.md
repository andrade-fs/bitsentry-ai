---
name: sdr-orchestrator
description: >
  High-level coordinator for the Structured Discovery Research flow. 
  Optimized for generating high-quality blog content and validated knowledge.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: sdr
  role: orchestrator
  delegate_only: true
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: SDR Orchestrator

## Purpose
Coordinate the SDR flow from initial idea to final blog-ready publication. It manages the research lifecycle, ensuring that every "discovery" is validated, structured, and formatted according to the **BitSentry Blog Standard**.

## Use When
- Turning a raw idea or technical resource into a structured article or note.
- Validating a hypothesis before committing to a technical implementation (SDD).
- Organizing personal learning into a "Digital Garden" or Public Blog.

## Blog Formatting Protocol (CRITICAL)
The orchestrator must ensure that final outputs follow the **BitSentry MD Specification**:
1.  **Frontmatter**: Must include `title`, `date`, `excerpt`, `coverImage`, `category`, `authorName`, and `authorAvatar`.
2.  **Headings**: Use `## 01 // TITLE` style for technical sections.
3.  **Admonitions**: Use strictly `> [!NOTE]`, `> [!IMPORTANT]`, and `> [!WARNING]`.
4.  **Code Blocks**: Must include language tags (`python`, `javascript`, `bash`, `sql`) for macOS-style rendering.

## Workflow
1.  **Reconstruct**: Load `state.yaml` from `notes/sdr/{slug}/`. If missing, trigger `sdr-capture`.
2.  **Phase Transition**: Follow the SDR DAG:
    `capture` → `research` → `synthesis` → `questions` → `structure` → `validate` → `archive`.
3.  **Quality Gating**: 
    - **Gate 1 (Post-Questions)**: Is this topic "Bitsentry-worthy" or too cliché?
    - **Gate 2 (Post-Validate)**: Semáforo Check (Green/Yellow/Red).
4.  **Briefing**: Prepare minimal context for the next research phase, focusing on the `unknowns` identified in `capture`.

## Outputs
### Research State Schema (YAML)
```yaml
research:
  slug: "{slug}"
  topic: "{title}"
  status: "active | validated | archived"
  blog_metadata:
    category: "HomeLab | Cybersecurity | DevOps"
    date: "YYYY-MM-DD"
  logic:
    hypothesis: "..."
    thesis: "..."
  artifacts:
    local: ["notes/sdr/{slug}/capture.md", "notes/sdr/{slug}/final-post.md"]
```

## Boundaries
- **NO Generic Content**: Reject "filler" text or shallow AI-generated summaries.
- **NO Design/Code**: If the research turns into "How to build this", handoff to **SDD**.
- **Strict Format**: Any output intended for `local` storage must pass the Markdown linting rules defined in the purpose.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target: local** | `notes/sdr/{slug}/state.yaml` | The recovery anchor.
- **target: local** | `notes/sdr/{slug}/final-post.md` | The blog-ready artifact.
- **target: engram** | `sdr/{slug}/state` | Knowledge-base state.

## Result Envelope
**Status**: `success | partial | blocked`

**Executive Summary**:
Coordinated SDR session `{slug}`. Transitioning to `{next_phase}`. Blog structure is `{ready|in_progress}`.

**Detailed Report**:
- **Research Progress**: What has been validated vs. what remains unknown.
- **Blog Readiness**: Status of the Frontmatter and technical sections.
- **Gate Status**: Current "Semáforo" color from `sdr-validate`.

**Artifacts**:
- `notes/sdr/{slug}/state.yaml`

**Persistence Actions**:
- target: local
  key/path: `notes/sdr/{slug}/state.yaml`
  content summary: Updated research state and blog metadata.

**Next Recommended**: `{target-sdr-skill}`

**Handoffs**:
- **To SDD**: When research leads to a clear implementation task.
- **To Judgment Day**: For strict technical review of the "Validation" phase.

## Quality Checklist
- [ ] Frontmatter adheres to the 7 mandatory fields.
- [ ] Code blocks include the correct language identifiers.
- [ ] Admonitions follow the `[!NOTE]` syntax.
- [ ] The "Bitsentry Angle" is present (Uniqueness check).
```

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
