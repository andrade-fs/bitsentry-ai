---
name: handoff-contract
description: >
  To eliminate context drift and "hallucination gaps" during transitions between flows (SDD/SDR) or helper skills.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: shared
  status: declarative
---

## Purpose
To eliminate context drift and "hallucination gaps" during transitions between flows (SDD/SDR) or helper skills. This contract enforces a **zero-loss transfer** of intent, artifacts, and unresolved risks.

## 1. Valid Transition Matrix

| Source Flow | Target Flow | Trigger Condition |
| :--- | :--- | :--- |
| **SDD** | **SDR** | Technical unknown, library benchmarking, or architectural feasibility study needed. |
| **SDR** | **SDD** | Research validates a hypothesis that now requires implementation/prototyping. |
| **Any** | **Support** | Triggering specialized helpers (Judgment Day, Issue Creation, Go-Testing). |
| **Any** | **Archive** | Finalizing session and persisting lineage. |

---

## 2. The Mandatory Handoff Payload
Every handoff MUST be serialized as a YAML/JSON object within the `Result Envelope`.

```yaml
handoff:
  metadata:
    from_id: "{skill_id}"
    to_id: "{flow_id | skill_id}"
    session_slug: "{slug}"
  context:
    reason: "Clear, single-sentence justification."
    entry_point: "{phase_id}" # Where the target should start
  assets:
    artifacts: ["path/to/essential/file.md", "engram:obs_id"]
    state_ref: "engram:sdd/{slug}/state"
  instruction:
    open_questions:
      - "Specific question 1?"
      - "Specific question 2?"
    constraints:
      - "Preserve existing X pattern"
      - "Do not exceed Y scope"
  risk_register:
    - "Unresolved assumption about Z"
```

---

## 3. The "Protocol of Acceptance" (Reasoning)
When a skill receives a handoff, it MUST:
1.  **Validate**: Confirm all `assets.artifacts` are accessible.
2.  **Acknowledge**: Explicitly mention the `reason` for the handoff in its first output.
3.  **Address**: Prioritize the `instruction.open_questions` before proceeding to its core workflow.

---

## 4. Quality Guardrails (The "No-Fly" Zone)
Handoffs will be considered **Malformed** if:
* **Generic Intent**: Reasons like "Continue work" or "Next step" are prohibited.
* **Missing Lineage**: No reference to the `session_slug` or `state_ref`.
* **Orphan Handoff**: Providing instructions without the artifacts necessary to execute them.
* **Scope Leak**: Asking the target flow to do something explicitly marked as a `non-goal` in the origin flow.

---

## 5. Typical Handoff Scenarios

### A. SDD → SDR (The "I'm Stuck" Handoff)
* **Reason**: "Architecture of the legacy `auth` module is undocumented; need discovery before proposing refactor."
* **Open Questions**: "¿How does the current session store handle expiration?"

### B. SDR → SDD (The "Build It" Handoff)
* **Reason**: "Research confirms `lib-xyz` is compatible; ready to implement the adapter."
* **Open Questions**: "¿Can we implement the `Provider` interface without breaking changes?"

---

### ¿Qué hemos mejorado?
1.  **Estructura YAML**: Ahora es un esquema de datos, no solo una lista de texto. Esto permite que tu código (el runtime) pueda validar el handoff automáticamente.
2.  **Protocolo de Aceptación**: Obligamos al agente receptor a decir: "He recibido esto por X razón y voy a responder Y".
3.  **Handoffs Tipados**: Separamos `metadata`, `assets` e `instructions` para que la IA sepa qué es un dato y qué es una orden.

