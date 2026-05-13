---
name: security-findings
description: Normalize findings with severity, confidence, and mitigation guidance.
license: MIT
metadata:
  author: Bitsentry
  version: "1.0"
  family: security
  phase: findings
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/handoff-contract.md
---

# Skill: security-findings

## Purpose
Convert review candidates into actionable security findings with clear severity and remediation direction.

## Findings Contract (Minimum Required Tokens)
Every finding entry MUST include these exact field labels:

- ID
- Title
- Severity
- Confidence
- Category
- Affected files
- Affected component
- Evidence
- Impact
- Likelihood
- Remediation
- Verification
- References / Notes

## Official Taxonomy (Category Enum — Exact)
- Authentication
- Authorization
- Session Management
- Input Validation
- Injection
- Cross-Site Scripting
- Server-Side Request Forgery
- File Upload
- Secrets Exposure
- Cryptography
- Dependency Risk
- GraphQL Security
- Business Logic
- Configuration
- Logging / Monitoring
- Error Handling
- Data Exposure
- Supply Chain
- Informational

### Allowed Values (Exact)
Severity:
- Critical
- High
- Medium
- Low
- Informational

Confidence:
- High
- Medium
- Low

## Severity Calibration (Impact × Likelihood)
- Critical: impacto muy alto + likelihood alta o exposición clara con riesgo sistémico.
- High: impacto alto con likelihood razonable, o impacto crítico con evidencia parcial.
- Medium: impacto moderado o explotación condicionada.
- Low: impacto limitado, alcance reducido o mitigaciones claras.
- Informational: observación útil sin impacto explotable confirmado.

## Confidence Calibration (Evidence Quality)
- High: evidencia directa en código/config no secreta, flujo claro y baja ambigüedad.
- Medium: patrón razonable con algunas asunciones.
- Low: señal débil, contexto incompleto o requiere verificación manual.

## Skill → Category Mapping Contract
Use this mapping in `security-map` and in final findings normalization.

- auth-security-review -> primary: Authentication | secondary: Authorization, Session Management
- jwt-review -> primary: Session Management | secondary: Authentication, Cryptography
- graphql-security-review -> primary: GraphQL Security | secondary: Authorization, Injection, Data Exposure
- xss-review -> primary: Cross-Site Scripting | secondary: Input Validation
- file-upload-review -> primary: File Upload | secondary: Input Validation, Configuration
- ssrf-review -> primary: Server-Side Request Forgery | secondary: Input Validation, Configuration
- secrets-review -> primary: Secrets Exposure | secondary: Configuration, Supply Chain
- dependency-risk-review -> primary: Dependency Risk | secondary: Supply Chain, Configuration

## Deduplication Rules
- Deduplicate by semantic key: `Category + affected component + sink/source path + root cause pattern`.
- Merge duplicate candidates into one canonical finding when evidence points to the same vulnerability condition.
- Preserve all supporting evidence references under the canonical finding (do not drop evidence lines).
- Do not deduplicate distinct exploit paths that require different remediation.

## Evidence Grouping Rules
- Group evidence by finding ID and by affected component.
- Keep direct code/config excerpts (non-secret) separate from inferred reasoning notes.
- Tag each evidence item as `direct` or `inferred` to support Confidence calibration.
- Include stable pointers (file path + line/range or config key path) for each evidence item.

## Assumptions / Limitations Rules
- Record explicit assumptions used to score Likelihood or exploitability.
- Record limitations caused by missing runtime context, missing infra context, or partial repository visibility.
- If Confidence is `Low` or `Medium`, include a concrete manual verification step.

## Use When
- `security-review` has candidate findings.
- A formal findings gate is required before final reporting.

## Inputs
- Candidate findings with evidence.
- Scope and risk rubric from previous stages.

## Workflow
1. Deduplicate and normalize finding statements.
2. Assign severity and confidence using declared rubric.
3. Add impact, likelihood, and exploit preconditions (conceptual, no execution).
4. Add mitigation guidance and verification hints.
5. Validate all minimum finding tokens and allowed value enums.
6. Output findings manifest for final report.

## Outputs
- `findings.md` with normalized findings register.

## Boundaries
- read-only first.
- OpenCode-first.
- no runtime flow execution.
- no autonomous mode.
- no edits by default.
- agent.bitsentry.permission.edit = deny.
- no .env access.
- no secrets.
- no exploit execution.
- no external target testing.
- no destructive actions.
- no MCP credential mutation.
- CLI debug/plumbing only.
- NO vulnerability proof-of-concept execution.
- NO runtime mutation or environment manipulation.
- NO credential or secret operations.

## Persistence Actions
- **target**: local
  **key/path**: `security-review/{slug}/findings.md`
  **action**: upsert
  **summary**: Normalized findings with severity and mitigations.

## Result Envelope
**Status**: `success | blocked`

**Executive Summary**:
Findings gate completed with severity-ranked, actionable entries.

**Next Recommended**: `security-report`

## Handoffs
- explicit handoff: `security-findings -> security-report` for final synthesis.
- to `support/issue-creation` when follow-up tracking is required.

### Handoff Payload Requirements (`security-findings -> security-report`)
- Include full taxonomy category per finding (from Official Taxonomy enum).
- Include Severity + Confidence calibrated via declared anchors.
- Include deduplicated canonical IDs with grouped evidence references.
- Include assumptions and limitations per finding when present.

## Quality Checklist
- [ ] Every finding has severity and evidence.
- [ ] Every finding contains all required contract tokens (exact labels).
- [ ] Severity value is one of: Critical, High, Medium, Low, Informational.
- [ ] Confidence value is one of: High, Medium, Low.
- [ ] Mitigations are actionable and bounded.
- [ ] No secrets/runtime/pentest behavior introduced.
