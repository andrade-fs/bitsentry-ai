---
name: engram-convention
description: >
  This document defines the **Source of Truth** for indexing and retrieving SDD/SDR state via the Engram MCP.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: shared
  status: declarative
---

## Overview
This document defines the **Source of Truth** for indexing and retrieving SDD/SDR state via the Engram MCP. Sub-agents must treat these keys as **Deterministic Addressable Memory**.

---

## 1. Global Namespace Rules
All keys follow a strict hierarchical URI-like structure to prevent collision and ensure fast retrieval:
`{flow}/{slug}/{artifact-type}`

* **Flow**: `sdd` | `sdr` | `support`
* **Slug**: The unique session ID created in `init` (kebab-case, e.g., `refactor-auth-layer`).
* **Artifact Type**: The specific output of a phase (e.g., `spec`, `design`).

### Metadata Contract
| Field | Value | Reason |
| :--- | :--- | :--- |
| **title** | `{flow}/{slug}/{type}` | Deterministic identification. |
| **topic_key** | `{flow}/{slug}/{type}` | **CRITICAL**: Enables UPSERT (update existing) instead of creating duplicates. |
| **type** | `architecture` | Constant category for technical documentation. |
| **project** | `{project_name}` | Scopes memory to the current repository. |

---

## 2. Artifact Mapping Table

| Flow | Phase | Type Key | Description |
| :--- | :--- | :--- | :--- |
| **SDD** | `init` | `sdd/{slug}/init` | Identity, Goal, and Scope. |
| **SDD** | `explore` | `sdd/{slug}/explore` | Impact map and dependency analysis. |
| **SDD** | `propose` | `sdd/{slug}/proposal` | High-level strategy (Gate 1). |
| **SDD** | `spec` | `sdd/{slug}/spec` | Functional requirements and ACs. |
| **SDD** | `design` | `sdd/{slug}/design` | Technical blueprint and API contracts. |
| **SDD** | `tasks` | `sdd/{slug}/tasks` | Atomic implementation checklist. |
| **SDD** | `verify` | `sdd/{slug}/verify` | Quality gate report and Verdict. |
| **SDD** | `archive` | `sdd/{slug}/archive` | Final manifesto and ID lineage. |
| **Any** | `orchestrator` | `{flow}/{slug}/state` | **Master State Object** (The recovery anchor). |

---

## 3. The State Object (Recovery Anchor)
The orchestrator MUST save a state object at every transition. This is the **Primary Entry Point** for any agent resuming a session.

**Action: `mem_save`**
```yaml
title: "sdd/{slug}/state"
topic_key: "sdd/{slug}/state"
content: |
  session_id: {slug}
  current_phase: {phase}
  status: active | blocked | partial
  gate_approval: {true|false}
  artifacts:
    init: "obs_001"
    explore: "obs_002"
    # ... IDs for fast retrieval
  last_updated: 2026-05-07T19:56:00Z
```

---

## 4. Operational Protocol (Deterministic AI Steps)

### A. Initialization & Updates (The UPSERT Rule)
To avoid polluting Engram with multiple versions of the same file, always use the same `topic_key`. 
- **New/Update**: Use `mem_save` with the specific `topic_key`. Engram will automatically handle the versioning or overwrite based on that key.

### B. Recovery (The "Search-then-Get" Dance)
`mem_search` previews are **truncated**. To get the full context without data loss, follow this protocol:

1.  **Step 1: Search IDs**
    `mem_search(query: "sdd/{slug}/", project: "{project}")`
2.  **Step 2: Full Content Retrieval**
    `mem_get_observation(id: "obs_ID_from_step_1")`
    *Always perform Step 2 before making architectural decisions.*

---

## 5. Traceability & Lineage
The `sdd-archive` phase MUST gather all Observation IDs from the current session and store them in the final archive report. This creates a **Permanent Lineage** for future audits or automated rollbacks.

---

## 6. Why This Convention Works
* **Zero Ambiguity**: The `slug` + `topic_key` ensures a 1:1 mapping between a file and its memory slot.
* **Fast Recovery**: A new agent can reconstruct the entire project state with just two tool calls.
* **Tool Safety**: Explicitly prevents the IA from acting on truncated search previews.

