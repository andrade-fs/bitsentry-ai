---
name: secrets-review
description: Read-only checklist to detect secret exposure patterns and unsafe secret-handling practices.
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

# Skill: secrets-review

## Purpose
Find source-level secret exposure risks while strictly avoiding secret retrieval or disclosure.

## Use When
- Config/bootstrap/logging code is in scope.
- Repo history or files may contain credential-like patterns.

## Inputs
- Source/config templates, CI/CD config, logging/error handlers, and sample env docs.
- Secret management integration code paths (vault/KMS wrappers).

## Workflow
1. Review secret ingestion and usage boundaries in application code.
2. Quick triage checklist:
   - No hardcoded credentials/tokens/private keys in source.
   - Logs/errors do not print secrets.
   - Config templates avoid real secret values.
   - Rotation/revocation hooks exist for compromised credentials.
3. Risk patterns:
   - Secrets committed in fixtures/examples.
   - Debug logging of auth headers/tokens.
   - Credentials embedded in URLs.
4. Capture only safe evidence: locations + masked placeholders, never raw secret value.
5. Produce actionable candidate finding.

## Outputs
- Secret-handling findings with sanitized evidence and remediation steps.

## Boundaries
- Read-only first; no edits by default.
- Explicitly NO `.env` access and NO secret extraction/reveal.
- No exploit execution, no external target testing, no destructive operations.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: Secret-handling candidate findings with sanitized evidence.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Secret exposure risk was assessed with strict non-disclosure evidence handling.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High when active credential exposure likely, Medium for process/control gaps), false-positive notes, and response actions.
- To `security/security-report`: include residual risk requiring key rotation program updates.

## Quality Checklist
- [ ] No raw secret values captured.
- [ ] Hardcoded/logging/template checks completed.
- [ ] False positives documented (dummy/test placeholders).
- [ ] Finding output includes immediate containment guidance.
