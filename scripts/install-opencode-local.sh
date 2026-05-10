#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"
BIN="${BIN_DIR}/bitsentry-ai"
EXPORT_CANDIDATES=(
  "${HOME}/.opencode/bitsentry"
  "${HOME}/.config/opencode/bitsentry"
)

log() {
  printf "\n==> %s\n" "$1"
}

fail() {
  printf "\n[FAIL] %s\n" "$1" >&2
  exit 1
}

log "Resolving repository root"
printf "ROOT_DIR=%s\n" "${ROOT_DIR}"

log "Checking dependencies"
command -v go >/dev/null 2>&1 || fail "go is required but was not found in PATH"
command -v gofmt >/dev/null 2>&1 || fail "gofmt is required but was not found in PATH"

log "Formatting Go files"
GOFMT_INPUTS=()
while IFS= read -r file; do
  [[ -n "${file}" ]] || continue
  GOFMT_INPUTS+=("${ROOT_DIR}/${file}")
done < <(git -C "${ROOT_DIR}" ls-files '*.go')

if [[ ${#GOFMT_INPUTS[@]} -gt 0 ]]; then
  gofmt -w "${GOFMT_INPUTS[@]}"
else
  printf "No Go files found.\n"
fi

log "Running tests"
(cd "${ROOT_DIR}" && go test ./...)

log "Building bitsentry-ai binary"
mkdir -p "${BIN_DIR}"
(cd "${ROOT_DIR}" && go build -o "${BIN}" ./cmd/bitsentry-ai)

log "Ensuring OpenCode home directory exists"
mkdir -p "${HOME}/.opencode"

log "Capabilities status"
(cd "${ROOT_DIR}" && "${BIN}" capabilities status)

log "Configuring dogfooding preset"
(cd "${ROOT_DIR}" && "${BIN}" capabilities configure --preset bitsentry-dev)

log "Running OpenCode export preview"
(cd "${ROOT_DIR}" && "${BIN}" capabilities export-preview --target-agent opencode)

log "Running OpenCode export dry-run"
(cd "${ROOT_DIR}" && "${BIN}" capabilities export --target-agent opencode --dry-run)

log "Running OpenCode export"
(cd "${ROOT_DIR}" && "${BIN}" capabilities export --target-agent opencode)

log "Resolving exported managed root"
ACTUAL_EXPORT_ROOT=""
CHECKED_PATHS=()
for candidate in "${EXPORT_CANDIDATES[@]}"; do
  usage_file="${candidate}/OPENCODE_USAGE.md"
  registry_file="${candidate}/skill-registry.md"
  CHECKED_PATHS+=("${candidate}")
  if [[ -f "${usage_file}" && -f "${registry_file}" ]]; then
    ACTUAL_EXPORT_ROOT="${candidate}"
    break
  fi
done

if [[ -z "${ACTUAL_EXPORT_ROOT}" ]]; then
  fail "Could not resolve OpenCode managed export root. Checked: ${CHECKED_PATHS[*]}"
fi

if [[ "${ACTUAL_EXPORT_ROOT}" != "${HOME}/.opencode/bitsentry" ]]; then
  printf "\nNOTE: OpenCode export root resolved to: %s\n" "${ACTUAL_EXPORT_ROOT}"
fi

log "Verifying exported files"
[[ -f "${ACTUAL_EXPORT_ROOT}/OPENCODE_USAGE.md" ]] || fail "Missing ${ACTUAL_EXPORT_ROOT}/OPENCODE_USAGE.md"
[[ -f "${ACTUAL_EXPORT_ROOT}/skill-registry.md" ]] || fail "Missing ${ACTUAL_EXPORT_ROOT}/skill-registry.md"

printf "\n[PASS] OpenCode capability pack exported successfully.\n"
printf "Export path: %s\n" "${ACTUAL_EXPORT_ROOT}"
printf "\nNext recommended OpenCode dogfooding prompt:\n"
printf "\"Use the exported Bitsentry capability pack at %s and run an SDD-style change proposal flow for this repository, following %s and %s.\"\n" "${ACTUAL_EXPORT_ROOT}" "${ACTUAL_EXPORT_ROOT}/OPENCODE_USAGE.md" "${ACTUAL_EXPORT_ROOT}/skill-registry.md"
