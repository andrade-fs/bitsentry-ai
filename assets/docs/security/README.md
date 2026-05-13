# Security Docs Fixtures and Examples

This directory contains **safe, synthetic** examples and golden fixtures for the `source-security-review` documentation contracts.

## Structure

- `assets/docs/security/examples/`
  - `findings-example.md`
  - `report-example.md`
- `assets/docs/security/fixtures/`
  - `findings-golden.md`
  - `report-golden.md`
  - `web-assessment-report-golden.md`
  - `web-assessment-evidence-golden.md`

## Purpose

- Provide deterministic reference artifacts for static contract tests.
- Document expected finding/report structure without executing flows.
- Preserve contract clarity for section order, enum usage, and taxonomy consumption.

## Safety and Guardrails (Mandatory)

- read-only first
- no .env access
- no secrets
- no exploit execution
- no external target testing
- no destructive actions
- no MCP credential mutation
- no runtime flow execution
- no autonomous mode
- no edits by default
- OpenCode-first
- CLI debug/plumbing only
- `agent.bitsentry.permission.edit = deny`

## Notes

- All scenarios are synthetic (example component/demo config/sample handler).
- No sensitive data, live targets, or offensive payload instructions are included.

## web-assessment contracts

- These fixtures/examples are for the `web-assessment-*` namespace and are distinct from `source-security-review` fixtures.
- `source-security-review` fixtures model repository/read-only findings and report contracts.
- `web-assessment-*` fixtures model authorization/scope/request-evidence traceability for authorized web assessment reporting contracts.

Golden fixtures in this directory:
- `assets/docs/security/fixtures/web-assessment-report-golden.md`
- `assets/docs/security/fixtures/web-assessment-evidence-golden.md`

Alignment rules (skills ↔ fixtures):
- Keep anchor names identical between skills and fixtures.
- Update static tests whenever contract anchors change.
- Keep deterministic section order for report golden fixtures.
- Do not rely on long paragraph matching; validate exact tokens/anchors.

Mandatory restrictions for web-assessment fixtures/examples:
- no runtime
- no flow execution
- no tooling operativo
- no target testing vivo
- no exploits
- no DoS/load testing
- no credential attacks
- no secrets exposure
- evidence must redact sensitive data
- no MCP credential mutation
- no autonomous mode
- OpenCode-first
- CLI debug/plumbing only
- agent.bitsentry.permission.edit = deny
