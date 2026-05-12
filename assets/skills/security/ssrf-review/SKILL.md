---
name: ssrf-review
description: Bounded source checklist for SSRF exposure in outbound request and fetch pipelines.
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

# Skill: ssrf-review

## Purpose
Identify SSRF-prone source patterns and controls in outbound network call paths, without runtime probing.

## Use When
- Application fetches URLs/resources influenced by user input.
- Integrations/webhooks/crawlers/preview services are in scope.

## Inputs
- HTTP client wrappers, URL parsers/validators, proxy settings, and allowlist logic.
- Any metadata/internal-service protections in code/config.

## Workflow
1. Trace user-controlled URL fields to outbound request sinks.
2. Quick triage checklist:
   - Strong URL parsing and canonicalization.
   - Destination allowlist/denylist enforcement.
   - Private/internal IP and metadata endpoint blocks.
   - Redirect handling restrictions.
3. Risk patterns:
   - Regex-only URL validation.
   - Blind proxying/fetching arbitrary URLs.
   - DNS rebinding-unaware validation flow.
4. Capture safe evidence references (code/config lines), no network execution.
5. Build actionable candidate finding.

## Outputs
- SSRF-oriented findings with preconditions, impact boundary, evidence, and mitigations.

## Boundaries
- Read-only first; no default edits.
- No live probing/exploit execution/external testing.
- No `.env` reads, no secrets handling, no destructive actions.
- No MCP credential mutation, no runtime flow execution, no autonomous mode.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: SSRF candidate findings and safe evidence.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Outbound request paths were reviewed for SSRF controls and misuse patterns with no runtime probes.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High when internal network/metadata exposure is plausible, Medium for constrained misuse), false positives, and mitigations.
- To `security/security-report`: include residual SSRF risk where network controls are external dependencies.

## Quality Checklist
- [ ] Source-to-outbound sink trace documented.
- [ ] Allowlist/internal-network/redirect controls checked.
- [ ] False positives captured (egress proxy enforces destination policy).
- [ ] Finding output remains actionable and bounded.
