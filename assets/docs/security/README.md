# Security Docs Fixtures and Examples

This directory contains **safe, synthetic** examples and golden fixtures for the `source-security-review` documentation contracts.

## Structure

- `assets/docs/security/examples/`
  - `findings-example.md`
  - `report-example.md`
- `assets/docs/security/fixtures/`
  - `findings-golden.md`
  - `report-golden.md`

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
