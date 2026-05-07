---
name: persistence-contract
description: >
  Mandatory protocol for declaring state and artifact persistence. 
  Ensures consistent storage across Engram, OpenCode, and local files.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: shared
  status: declarative
---

# Persistence Contract (v1.1)

## Purpose
In the Bitsentry ecosystem, skills are declarative. They do not write to disk directly; instead, they produce a **Persistence Manifesto**. This contract ensures that any orchestrator can execute these storage actions consistently, maintaining a "Source of Truth" across sessions.

## 1. The Persistence Manifesto (Output Block)
Every skill MUST include a `Persistence Actions` block in its Result Envelope using the following schema:

```yaml
persistence_actions:
  - target: "engram | opencode | local"
    priority: "critical | standard"
    key_path: "{flow}/{slug}/{artifact_type}"
    format: "markdown | yaml | json"
    content_summary: "Brief 1-line description of the data."
    action: "upsert | append"
```

---

## 2. Deterministic Key Strategy
To ensure recoverability, keys MUST follow the slug-based hierarchy defined in `engram-convention.md`.

### SDD Hierarchy (Software Development)
- **Engram Key**: `sdd/{slug}/{phase}`
- **Local Path**: `./sdd/{slug}/{phase}.md`
- **State File**: `./sdd/{slug}/state.yaml` (The anchor)

### SDR Hierarchy (Research & Discovery)
- **Engram Key**: `sdr/{slug}/{phase}`
- **Local Path**: `./notes/sdr/{slug}/{phase}.md`
- **State File**: `./notes/sdr/{slug}/state.yaml`

---

## 3. Storage Target Responsibilities

| Target | Purpose | Retention |
| :--- | :--- | :--- |
| **Engram** | Long-term memory and cross-session retrieval. | Permanent (Managed via MCP) |
| **OpenCode** | References for the coding agent (artifacts). | Session-based |
| **Local** | Human-readable logs and project-specific state. | Persistent in Repo |

---

## 4. Session State Update Contract
Every skill MUST return the following state-delta to the orchestrator to ensure the flow remains deterministic:

```yaml
state_update:
  session_id: "{slug}"
  phase: "{current_phase}"
  status: "success | partial | blocked"
  last_updated: "{ISO_8601_TIMESTAMP}"
  next_recommended: "{next_skill_id}"
  artifact_registry:
    - type: "{phase_name}"
      uri: "local://sdd/{slug}/{phase}.md"
```

---

## 5. Persistence Boundaries & Safety
* **No Secrets**: Strictly forbid persisting API keys, tokens, or credentials to Engram or local files.
* **Concise Summaries**: Engram search relies on summaries. Keep `content_summary` technical and keyword-rich.
* **Atomic Updates**: Declare one action per artifact. Do not bundle multiple phases into a single persistence key.
* **Idempotency**: Use `upsert` logic (topic_key in Engram) to ensure that re-running a skill doesn't create duplicate "ghost" artifacts.

---

## 6. Implementation Example (from sdd-propose)

**Persistence Actions**:
- **target**: `engram`
  **key**: `sdd/auth-refactor/proposal`
  **action**: `upsert`
  **summary**: Strategic approach for OAuth2 migration and risk mitigation.
- **target**: `local`
  **path**: `sdd/auth-refactor/proposal.md`
  **action**: `upsert`
  **summary**: Human-readable proposal for developer review.
```

