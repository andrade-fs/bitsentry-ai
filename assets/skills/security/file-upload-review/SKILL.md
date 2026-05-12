---
name: file-upload-review
description: Read-only checklist for unsafe file-upload validation, storage, and processing flows.
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

# Skill: file-upload-review

## Purpose
Assess file upload attack surface through source analysis and produce bounded, actionable findings.

## Use When
- The system accepts user files/images/documents/media.
- Review includes ingestion, storage, scanning, or transformation pipeline.

## Inputs
- Upload handlers, MIME/type checks, storage adapters, and processing jobs.
- Access control and retrieval/download paths.

## Workflow
1. Map upload entrypoint to storage and consumption points.
2. Quick triage checklist:
   - Extension + MIME + magic-byte validation enforced.
   - Size/count limits and rate controls present.
   - Storage path isolation and non-executable serving configured.
   - Optional malware scanning/quarantine policy defined.
3. Risk patterns:
   - Trusting client-supplied content-type.
   - Path traversal in filename handling.
   - Public bucket/object exposure without auth controls.
4. Capture safe evidence from handlers/config only (no real file exploitation).
5. Emit actionable candidate findings.

## Outputs
- File-upload findings with evidence pointers, severity direction, and concrete mitigations.

## Boundaries
- Read-only first; no default edits.
- No exploit file creation/execution, no external target testing.
- No `.env` access, no secrets exposure, no destructive actions.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: File-upload security candidates and evidence references.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
File-upload pipeline was reviewed for validation, storage isolation, and abuse controls.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High for upload-to-RCE or public exposure vectors, Medium for weak validation), false-positive notes, and remediation plan.
- To `security/security-report`: include residual risk where infra/storage hardening is required.

## Quality Checklist
- [ ] End-to-end upload flow mapped.
- [ ] Validation + storage + access checks documented.
- [ ] False positives addressed (upstream gateway/blob policy mitigates risk).
- [ ] Output is actionable for engineering remediation.
