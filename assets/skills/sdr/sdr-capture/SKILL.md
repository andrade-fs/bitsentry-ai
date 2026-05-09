---
name: sdr-capture
description: >
  Initial stage of SDR. Captures raw ideas, links, or documents and 
  normalizes them into a Research Session with Blog-aligned metadata.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: sdr
  phase: capture
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
    - _shared/engram-convention.md
---

# Skill: sdr-capture

## Purpose
Convert a raw research intent into a deterministic SDR session. It establishes the "Seed" of the blog post by defining the slug, initial hypothesis, and the metadata that will eventually populate the MD Frontmatter.

## Use When
- You have a link, a technical concept, or a personal experience to document.
- You need to decide if a topic is a "quick note" or a "deep research" piece.
- You want to start a knowledge-gathering session that ends in a BitSentry blog post.

## Inputs
- **Raw Input**: Topic, URL, snippet, or idea.
- **Category (Optional)**: `HomeLab`, `Cybersecurity`, `DevOps`, etc.
- **Initial Context**: Why this topic is relevant now.

## Workflow
1.  **Slug Generation**: Create a deterministic kebab-case `slug` based on the topic (e.g., `proxmox-zfs-best-practices`).
2.  **Metadata Pre-filling**: Initialize the Blog Frontmatter fields:
    - `title`: A working title.
    - `date`: Current date in `YYYY-MM-DD`.
    - `category`: Map to allowed blog categories.
    - `authorName`: Default to "BitSentry".
3.  **Hypothesis Definition**: Formulate a "Thesis Statement" (e.g., "ZFS RAID 10 is superior to RAID-Z for HomeLabs due to IOPS and rebuild safety").
4.  **Source Catalog**: List any provided URLs or local files as primary sources.
5.  **Scope Guardrails**: Define `In-Scope` questions and `Non-Goals` to prevent research sprawl.
6.  **State Creation**: Initialize the `state.yaml` following the shared persistence contract.

## Outputs
### Initial Research State (YAML)
```yaml
research:
  slug: "{slug}"
  topic: "{working_title}"
  hypothesis: "{thesis}"
  status: "captured"
  blog_frontmatter:
    title: "{working_title}"
    date: "{current_date}"
    excerpt: "" # To be filled in synthesis
    category: "{category}"
    authorName: "BitSentry"
    coverImage: "/assets/blog/{slug}/cover.png"
  unknowns: ["Q1", "Q2", "Q3"]
```

## Boundaries
- **NO Searching**: Do not perform the actual research; only define the plan.
- **NO Content Writing**: Do not generate the blog body yet.
- **Format Integrity**: Must strictly use the `slug` for all directory and engram naming.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: `notes/sdr/{slug}/state.yaml`
  **action**: upsert
  **summary**: Initial research state and blog metadata.
- **target**: engram
  **key**: `sdr/{slug}/capture`
  **action**: upsert
  **summary**: Topic hypothesis and initial unknown questions.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Captured topic `{slug}`. Identity and blog metadata initialized. Ready for `sdr-research`.

**Persistence Actions**:
- [Manifesto as defined above]

**Next Recommended**: `sdr-research`

**Handoffs**:
- **to**: `sdd-init`
  **condition**: If the input is actually a request to change code in an existing project.

**Risks / Gaps**:
- Identified if the initial sources are insufficient or if the topic is too broad.

## Quality Checklist
- [ ] `slug` is deterministic and kebab-case.
- [ ] Blog Frontmatter structure is present in the state.
- [ ] At least 3 specific `unknowns` are listed.
- [ ] Non-goals are explicit to keep research focused.
```

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
