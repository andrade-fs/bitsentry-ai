---
name: sdr-research
description: >
  Deepens topic understanding by extracting evidence and validating claims.
  The "Fact-Finding" phase of SDR.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: sdr
  phase: research
  status: declarative
---

# Skill: sdr-research

## Purpose
Investigate the `unknowns` from the capture phase. This skill gathers raw technical data, verifies facts, and identifies the core concepts that will form the backbone of the blog post.

## Use When
- `sdr-capture` is successful.
- You need to dive into documentation, logs, or external sources to find evidence.

## Workflow: The "Evidence Protocol"
1.  **Unknown Resolution**: Address each question from the `state.yaml`.
2.  **Fact Extraction**: Identify key technical data (versions, commands, configurations).
3.  **Cross-Reference**: Compare sources to find contradictions or unique "BitSentry" angles.
4.  **Claim Validation**: Mark findings as `verified`, `hypothetical`, or `debunked`.
5.  **Technical Snippet Gathering**: Collect high-quality code blocks (`python`, `bash`, etc.) for the future post.

## Outputs
### Research Report (Markdown)
- **Findings**: Categorized by the `unknowns` resolved.
- **Evidence Log**: Links, quotes, and technical specs.
- **Risks**: Areas where information is still shallow or cliché.

## Persistence Actions
- **target: local** | `notes/sdr/{slug}/research.md` | Full research log.
- **target: engram** | `sdr/{slug}/evidence` | Structured technical facts.

**Next Recommended**: `sdr-synthesis`