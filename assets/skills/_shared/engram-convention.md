# Engram Convention (Contract)

This phase defines conventions only; Engram runtime is not required.

Rules:
- stable topic keys by flow + phase
- upsert same topic for evolving artifacts
- keep executive summary + full details separable
- avoid storing secrets

Key format:
- `sdd/{change-name}/{phase}`
- `sdr/{topic-id}/{phase}`
