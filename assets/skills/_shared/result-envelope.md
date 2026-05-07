---
name: result-envelope
description: >
  Mandatory communication protocol for all Bitsentry skills. 
  Ensures deterministic parsing and consistent state reporting.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: shared
  status: declarative
---

# Result Envelope Contract (v1.1)

## Purpose
Standardize the "Return Object" of every skill. This contract ensures the Orchestrator can programmatically decide the next step, track artifacts, and handle errors without human intervention.

## 1. Status Matrix
| Value | Condition |
| :--- | :--- |
| **success** | Goal met, all artifacts generated, no critical blockers. |
| **partial** | Primary goal met, but some non-critical items were deferred. |
| **blocked** | Goal not met due to missing info, technical error, or safety risk. |

## 2. Mandatory Structure (Markdown)

### 2.1 Status & Summary
- **Status**: `{success | partial | blocked}`
- **Executive Summary**: 1-3 sentences on what was achieved and why it matters.

### 2.2 Detailed Report
The core intellectual output of the skill (logic, analysis, blueprints, or validation).

### 2.3 Artifact Index
| Key/Path | Description | Format |
| :--- | :--- | :--- |
| `path/to/file` | Brief usage description | `md | yaml | json` |

### 2.4 Persistence Manifesto
*Must comply with `_shared/persistence-contract.md`.*
- **target**: `engram | opencode | local`
- **key/path**: Hierarchical identifier.
- **content summary**: One-line data description.

### 2.5 Strategic Planning
- **Next Recommended**: The specific skill or phase ID to execute next.
- **Handoffs**: Mandatory YAML block if switching flows (compliant with `_shared/handoff-contract.md`).
- **Risks / Gaps**: Explicit list of unresolved assumptions or missing evidence.

---

## 3. Quality Guardrails
1.  **Deterministic Parsing**: Use clear headers (`##`) and bullet points. Avoid flowery prose in sections 1, 4, 5, and 6.
2.  **Explicit Blockers**: If status is `blocked`, the first line of the Executive Summary must state the exact blocker.
3.  **Traceability**: Every result must reference at least one artifact path.
4.  **Zero-Hallucination**: If a value is unknown, mark it as `UNKNOWN` instead of guessing.

## 4. Example Output (SDR-Capture)
> **Status**: success  
> **Executive Summary**: Research session for `zfs-tuning` initialized. Hypothesis centered on ARC cache optimization for low-RAM environments.  
> 
> **Detailed Report**: [Structured hypothesis, sources, and scope...]  
> 
> **Artifacts**:  
> - `notes/sdr/zfs-tuning/state.yaml`: Master state file.  
> 
> **Persistence Actions**:  
> - target: local | key: `notes/sdr/zfs-tuning/state.yaml` | summary: Initial state and metadata.  
> 
> **Next Recommended**: `sdr-research`  
> 
> **Risks / Gaps**: Limited documentation found for ZFS on kernels < 5.15.