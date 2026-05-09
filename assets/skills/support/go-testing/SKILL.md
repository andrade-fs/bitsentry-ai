---
name: support-go-testing
description: >
  Executes and manages deterministic test suites for Go CLI/TUI projects. 
  Focuses on table-driven tests and hermetic execution.
license: MIT
metadata:
  author: Bitsentry
  version: "1.1"
  family: support
  status: declarative
  requires:
    - _shared/result-envelope.md
    - _shared/persistence-contract.md
    - _shared/engram-convention.md
---

# Skill: support-go-testing

## Purpose
Ensure the technical integrity of Go-based components. This skill automates the formatting, execution, and coverage analysis of the test suite, prioritizing deterministic results and avoiding side effects in the host system.

## Use When
- `sdd-apply` has modified `.go` files.
- You need a pre-flight check before `sdd-verify`.
- Validating TUI/CLI logic or complex algorithms.

## Workflow
1.  **Static Analysis**: 
    - Execute `gofmt -s -l` to check formatting.
    - Run `go vet ./...` to detect common pragmatism issues.
2.  **Test Execution**:
    - Execute `go test -v -race ./...` (Race detector enabled).
    - Use `t.Parallel()` in table-driven tests to maximize speed.
3.  **Environment Isolation**:
    - Enforce the use of `t.TempDir()` for all filesystem-related tests.
    - Ensure `os.Environ()` is mocked or cleaned up after execution.
4.  **Coverage Analysis**:
    - Generate a coverage profile: `go test -coverprofile=coverage.out ./...`.
    - Identify "Cold Zones" (untested critical logic).
5.  **TUI/CLI Mocking**:
    - For TUI components (Bubbletea/Cobra), ensure `io.Writer` and `io.Reader` are injected and tested via `bytes.Buffer`.

## Outputs
### Test Execution Report (Markdown)
Must include:
- **Test Summary**: `Total Passed | Failed | Skipped`.
- **Race Condition Report**: Any issues detected by `-race`.
- **Coverage Map**: % per modified package.
- **Recommendations**: Suggested tests for edge cases or error paths.

## Boundaries
- **No Network**: Strictly no external API calls (use `httptest.NewServer` if needed).
- **No Side Effects**: Tests must not leave artifacts outside of `t.TempDir()`.
- **No External Services**: Use mocks/interfaces for DB or Cache layers.

## Persistence Actions
*Compliant with _shared/persistence-contract.md*
- **target**: local
  **key/path**: `test-reports/{slug}-go-testing.md`
  **action**: upsert
  **summary**: Go test execution results and coverage report.
- **target**: engram
  **key**: `support/go-testing/{slug}`
  **action**: append
  **summary**: Historical test data for the current change.

## Result Envelope
**Status**: `success | partial | failed`

**Executive Summary**:
Go testing for `{slug}` completed. Coverage: `{XX}%`. Race detector: `{CLEAN|ISSUES}`.

**Next Recommended**: `sdd-verify`

**Handoffs**:
- **to**: `sdd-verify`
  **reason**: Test evidence ready for final validation.
- **to**: `sdd-apply`
  **reason**: Fix required for failing tests or race conditions.

## Quality Checklist
- [ ] Formatting is clean (`gofmt`).
- [ ] No race conditions detected.
- [ ] Table-driven tests used for complex logic.
- [ ] All filesystem tests use `t.TempDir()`.

## Inputs
- If not otherwise specified above, use the invoking user request and current project context.

## Handoffs
- If downstream phases/skills are needed, hand off according to the flow manifest and contracts.
