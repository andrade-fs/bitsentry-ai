---
name: skill-loading-contract
description: >
  Defines the deterministic sequence and strategy for injecting 
  skill guidance into AI contexts to minimize overhead.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: shared
  status: declarative
---

# Skill Loading Contract (v1.1)

## Purpose
Optimize token usage and maintain high fidelity in sub-agent execution. This contract ensures that any orchestrator (human or AI) provides the necessary "Contractual Guardrails" before a skill is executed, preventing "context drift."

## 1. Deterministic Load Order
To ensure the AI understands the hierarchy of rules, they MUST be injected in this specific sequence:

1.  **`result-envelope.md`**: The communication protocol (The "How to talk back").
2.  **`persistence-contract.md`**: The storage rules (The "How to save").
3.  **`handoff-contract.md`**: The transition rules (The "How to pass the ball").
4.  **`engram-convention.md`**: The memory keys (The "How to remember").
5.  **`opencode-convention.md`**: The artifact rules (The "How to handle code").
6.  **`SKILL.md`**: The specific task logic (The "What to do now").

## 2. Injection Strategy: "Lean Context"
Orchestrators must not dump raw files. They must inject a **Compact Guidance Block** that includes:

- **Constraint Header**: "You are operating under BitSentry Phase 3.7.x Contracts."
- **Active Envelope**: Explicitly remind the skill of the `Result Envelope` structure.
- **Persistence Anchor**: Provide the current session `slug` and the expected `target`.
- **Handoff Expectations**: Define which skills are valid next steps.

## 3. Minimal Injected Block (The "Bootstrap")
Every sub-skill execution must start with this context injection:

> **CONTRACTUAL OVERRIDE**:  
> - **Format**: Respond strictly using the `Result Envelope`.  
> - **Identity**: All artifacts must use slug `{slug}`.  
> - **Persistence**: Targets are `{local | engram | opencode}`.  
> - **Naming**: Use `{flow}/{slug}/{type}` hierarchy.

## 4. Anti-Patterns (The "Stop" List)
- **Bloated Injection**: Do not inject the documentation of `sdd-init` while running `sdd-verify`.
- **Implicit Contracts**: Never assume a skill "knows" the persistence rules. Always re-inject the compact manifesto.
- **Envelope Skipping**: Accepting a response that doesn't follow the `Result Envelope` header structure.
- **Global Context Leak**: Passing unrelated session data (e.g., SDR context into an SDD code fix).

## 5. Result Envelope Compliance
The orchestrator MUST validate the output against the `result-envelope.md` immediately upon receipt. If the envelope is missing or malformed, the orchestrator MUST trigger a **Retry/Correction** cycle before persisting any data.