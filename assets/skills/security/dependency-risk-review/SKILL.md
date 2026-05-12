---
name: dependency-risk-review
description: Source-based dependency and supply-chain risk checklist for security review flow.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: review-support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: dependency-risk-review

## Purpose
Provide an operational checklist for dependency and supply-chain risk signals using read-only project artifacts.

## Use When
- Dependency manifests/lockfiles are within review scope.
- User requests AppSec/supply-chain hardening posture checks.

## Inputs
- Manifest and lock files (`go.mod`, `package-lock`, `poetry.lock`, etc.).
- Build pipeline config and dependency update policy docs.

## Workflow
1. Identify direct/transitive dependency governance points.
2. Quick triage checklist:
   - Lockfiles committed and reproducible builds favored.
   - High-risk/unmaintained packages identified.
   - Update cadence and vulnerability triage process exists.
   - Integrity/signature/provenance controls documented where available.
3. Risk patterns:
   - Wildcard/unpinned critical dependencies.
   - Abandoned packages in security-sensitive paths.
   - Build scripts pulling unaudited remote artifacts.
4. Capture safe evidence: manifest lines, policy gaps, CI config references.
5. Emit actionable candidate findings.

## Outputs
- Dependency-risk candidates with severity cues, impact surface, and mitigation suggestions.

## Boundaries
- Read-only first; no default edits.
- No package install/runtime execution for exploitation.
- No `.env`/secret handling, no destructive actions, no external target testing.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: Dependency risk candidates and evidence pointers.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Dependency and supply-chain risks were reviewed from source artifacts using bounded checklists.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High for known exploitable critical components in reachable path, Medium for governance/control gaps), false-positive notes, and upgrade strategy.
- To `security/security-report`: include residual risk when modernization requires phased rollout.

## Quality Checklist
- [ ] Manifests/lockfiles and policy controls reviewed.
- [ ] Risk patterns mapped to impacted components.
- [ ] False positives considered (unused/isolated dependencies).
- [ ] Output maps each finding to concrete remediation action.
