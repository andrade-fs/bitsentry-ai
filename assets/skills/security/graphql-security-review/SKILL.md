---
name: graphql-security-review
description: Source-level checklist for GraphQL authz, query safety, and schema exposure risks.
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

# Skill: graphql-security-review

## Purpose
Guide read-only GraphQL security triage to uncover authorization, query-complexity, and data exposure issues.

## Use When
- Repository includes GraphQL schema/resolvers/gateway logic.
- `security-map` highlights GraphQL as a trust boundary.

## Inputs
- Schema definitions, resolver implementations, and GraphQL server config.
- Existing limits (depth, complexity, persisted queries) and introspection policy.

## Workflow
1. Locate schema roots, resolvers, auth middleware, and batching layers.
2. Quick triage checklist:
   - Field-level authz in sensitive resolvers.
   - Query depth/complexity limits configured.
   - Introspection and playground exposure policy controlled.
   - Error messages avoid sensitive internal leakage.
3. Risk patterns:
   - Authz only at UI layer, not resolver layer.
   - N+1 or batch loaders exposing cross-tenant data.
   - Missing limits enabling denial-of-wallet/resource abuse.
4. Capture safe evidence from schema/resolver snippets and config values.
5. Produce finding candidates with impact + remediation.

## Outputs
- GraphQL-specific actionable candidates for `security-findings` normalization.

## Boundaries
- Read-only first; no default edits.
- No runtime query firing, no exploit execution, no external target testing.
- No `.env` access, no secrets handling, no destructive operations.
- No MCP credential mutation, no autonomous mode, no runtime flow execution.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/review.md`
  **action**: append
  **summary**: GraphQL security candidate findings with bounded evidence.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
GraphQL surface was triaged for authz, exposure, and abuse controls with safe evidence capture.

**Next Recommended**: `security-findings`

## Handoffs
- To `security/security-findings`: include severity orientation (High for broken authz/data leakage, Medium for missing complexity controls), false-positive notes, and remediations.
- To `security/security-report`: include residual architecture risks around schema governance.

## Quality Checklist
- [ ] Resolver-level authz checks evaluated.
- [ ] Complexity/depth/rate controls reviewed.
- [ ] False positives noted (limits enforced at gateway layer).
- [ ] Output mapped to actionable finding template.
