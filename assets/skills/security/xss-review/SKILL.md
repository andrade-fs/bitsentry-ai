---
name: xss-review
description: Practical source-review checklist for reflected/stored/DOM XSS controls and gaps.
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

# Skill: xss-review

## Purpose
Perform bounded XSS-focused code review and produce actionable findings backed by safe evidence.

## Use When
- User-supplied content is rendered in templates/UI/HTML generation.
- Frontend/server rendering paths are part of scoped review.

## Inputs
- Rendering components/templates, sanitizer usage, CSP headers/config.
- Input validation/encoding utilities and rich-text/editor flows.

## Workflow
1. Identify untrusted input sources and output sinks.
2. Quick triage checklist:
   - Context-appropriate output encoding present.
   - Dangerous HTML APIs guarded (`innerHTML`, equivalent).
   - Sanitizer policy is explicit and centrally reused.
   - CSP exists and is not overly permissive.
3. Risk patterns:
   - Bypassing framework auto-escaping APIs.
   - Markdown/rich text rendered without sanitization.
   - Client-side only sanitization assumptions.
4. Safe evidence capture: file/function/sink references, no payload execution.
5. Draft candidate with severity and mitigation.

## Outputs
- XSS finding candidates with attack preconditions, evidence pointers, and fix guidance.

## Boundaries
- Read-only first and no edits by default.
- No exploit execution or live target testing.
- No `.env`/secrets handling, no destructive actions.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: XSS candidate findings with safe source evidence.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
XSS controls were reviewed through source paths and sink analysis without runtime exploitation.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High for stored/privileged sinks, Medium for constrained reflected cases), common false positives, and remediation steps.
- To `security/security-report`: include residual risk if CSP/sanitization architecture is incomplete.

## Quality Checklist
- [ ] Source-to-sink path documented.
- [ ] Encoding/sanitization/CSP checks captured.
- [ ] False positives addressed (framework auto-escape active).
- [ ] Finding output is clear and actionable.
