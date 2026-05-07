# Partial Verify — Fase 2.5: OpenCode Detection/Export-Preview

**Change**: fase2-5-opencode
**Type**: verification — partial, focused on OpenCode detection & export-preview
**Date**: 2026-05-06
**Mode**: Standard (no TDD artifacts found)

---

## 1. Summary

Phase 2.5 Batch 1 OpenCode features verified. All build, CLI, safety, and regression checks **PASS**. No issues found.

---

## 2. Commands Executed

### Build & Test
| Command | Result |
|---------|--------|
| `go mod tidy` | ✅ Clean |
| `go test ./...` | ✅ No tests found (none required) |
| `go build ./cmd/bitsentry-ai` | ✅ Built successfully |

### CLI Commands (new)
| Command | Result |
|---------|--------|
| `go run ./cmd/bitsentry-ai agents` | ✅ Lists OpenCode detected with path+version |
| `go run ./cmd/bitsentry-ai agents opencode status` | ✅ Shows binary, version, config path candidate, can-export-safely |
| `go run ./cmd/bitsentry-ai agents opencode export-preview` | ✅ Shows in-memory plan, selected MCPs/skills, no file writes |

### CLI Commands (regression)
| Command | Result |
|---------|--------|
| `go run ./cmd/bitsentry-ai components` | ✅ Shows all 7 components with status |
| `go run ./cmd/bitsentry-ai doctor` | ✅ Shows OS, arch, shell, package manager, config dir, dependencies |
| `go run ./cmd/bitsentry-ai components engram status` | ✅ Runtime details + config consistency |
| `go run ./cmd/bitsentry-ai components context7 status` | ✅ Metadata-backed status |
| `go run ./cmd/bitsentry-ai components mcps status` | ✅ 7 modeled MCPs |
| `go run ./cmd/bitsentry-ai components skills status` | ✅ 7 modeled skills |
| `go run ./cmd/bitsentry-ai version` | ✅ `bitsentry-ai v0.1.0-alpha` |
| `go run ./cmd/bitsentry-ai profiles` | ✅ 8 profiles listed |
| `go run ./cmd/bitsentry-ai config path` | ✅ Shows config dir and file |

---

## 3. OpenCode Detection Checklist

| Check | Status | Evidence |
|-------|--------|----------|
| Binary detection via `exec.LookPath("opencode")` | ✅ | `agents.go` → `agents.OpenCodeDetector.Detect()` uses `exec.LookPath` |
| Version extraction via `opencode --version` | ✅ | `opencode.go` L30-34: runs `opencode --version`, trims output |
| Detection result includes path + version | ✅ | `AgentDetectionResult` struct fields: `Path`, `Version` |
| Hint shown when binary not found | ✅ | `opencode.go` L23: sets `res.Hint` on not-found |
| `agents opencode status` shows detection result | ✅ | CLI output confirms: "binary detected: yes", path, version |
| `agents opencode export-preview` shows binary status | ✅ | CLI output confirms: "opencode binary detected: yes" |
| OpenCode binary gates export-safety flag | ✅ | `agents.go` L231-238: `CanExportSafely = true` only when `res.Found` |

---

## 4. Export-Preview Safety Checklist

Code review of all paths in `agents.go` and `opencode.go`:

| Check | Status | Evidence |
|-------|--------|----------|
| No file writes in `export-preview` | ✅ | `agents.go` L97-193: only `fmt.Fprintln/Fprintf` to stdout, no `os.WriteFile`, `ioutil.WriteFile`, or `os.Create` |
| No directory creation | ✅ | No `os.Mkdir*` calls in any new code |
| No OpenCode config modification | ✅ | `opencode.go` only has `Detect()`, `OpenCodeConfigPathCandidates()`, `ExistingOpenCodeConfigPaths()` — no write operations |
| No MCP server execution | ✅ | `export-preview` reads config metadata and runtime status only; does not exec any MCP binary |
| No skill file copying | ✅ | `export-preview` only builds `OpenCodeExportPlan` struct from config; no file I/O |
| Actions list explicitly says "preview only" | ✅ | `agents.go` L130-134: actions say "Prepare in-memory only", "Do not write", "Keep TODO" |
| Explicit "No OpenCode files were modified" footer | ✅ | `agents.go` L191: `fmt.Fprintln(out, "Preview only. No OpenCode files were modified.")` |

**Behavior confirmed**: `agents opencode export-preview` is read-only. Output is in-memory plan displayed to stdout. Zero filesystem side-effects.

---

## 5. Config Path Behavior Checklist

Code review of `OpenCodeConfigPathCandidates()` and `ExistingOpenCodeConfigPaths()`:

| Check | Status | Evidence |
|-------|--------|----------|
| Checks `~/.config/opencode` | ✅ | `opencode.go` L46: `filepath.Join(homeDir, ".config", "opencode")` |
| Checks `~/.opencode` | ✅ | `opencode.go` L47: `filepath.Join(homeDir, ".opencode")` |
| Checks project-local `./.opencode` only if present | ✅ | `opencode.go` L50-55: `os.Stat(projectLocal)` — only added if `statErr == nil && st.IsDir()` |
| Checks existing path from detector if available | ✅ | `agents.go` L224-228: prefers `found[0]` (existing) over `candidates[0]` (fallback candidate) |
| Does NOT scan entire filesystem | ✅ | Only calls `os.UserHomeDir()` once, `os.Stat()` on ≤3 fixed paths, no `filepath.Walk`, no `os.ReadDir` recursive |
| Deduplicates and sorts candidates | ✅ | `opencode.go` L57-70: dedup map + `sort.Strings()` before return |
| Handles empty/blank paths | ✅ | `opencode.go` L60-63: `strings.TrimSpace(p) != ""` guard |
| `status` command shows all found config paths | ✅ | `agents.go` L76-82: iterates `ConfigPathsFound` and prints each |
| `status` shows target candidate with existing-first priority | ✅ | `agents.go` L225: `found[0]` preferred; `candidates[0]` fallback with warning |

**Behavior confirmed**: Config path detection is scoped to 3 fixed candidate paths. No filesystem scanning beyond `os.Stat` on those paths.

---

## 6. Regression Checklist

| Command | Expected | Result |
|---------|----------|--------|
| `components` — all 7 components listed | Yes | ✅ |
| `components engram status` — runtime + config | Yes | ✅ |
| `components context7 status` — metadata-backed | Yes | ✅ |
| `components mcps status` — registry table | Yes | ✅ |
| `components skills status` — registry table | Yes | ✅ |
| `doctor` — system info + deps | Yes | ✅ |
| `version` — version string | Yes | ✅ |
| `profiles` — 8 profiles | Yes | ✅ |
| `config path` — config dir/file | Yes | ✅ |
| `agents` — OpenCode in agent list | Yes | ✅ |

**All regression checks: PASS**

---

## 7. Bugs Found

**None.**

---

## 8. Fixes Applied

**None required.**

---

## 9. Verdict

**✅ PASS**

All verification checks passed. OpenCode detection and export-preview behave correctly, safely, and without side effects. No regressions detected.

### Files Involved

| File | Role |
|------|------|
| `internal/agents/opencode.go` | OpenCode detector, config path candidates |
| `internal/agents/opencode_export_plan.go` | Export plan data model |
| `internal/cli/agents.go` | CLI commands: `agents`, `agents opencode status`, `agents opencode export-preview` |

### Evidence

- Build: ✅ `go build ./cmd/bitsentry-ai` passes
- Tests: ✅ `go test ./...` — no tests (none required for this change)
- `agents opencode status` output confirms detection + config path
- `agents opencode export-preview` output confirms in-memory plan with no file writes
- `OpenCodeConfigPathCandidates()` code review confirms 3-path limit, no filesystem scan
- `export-preview` code review confirms zero filesystem side-effects