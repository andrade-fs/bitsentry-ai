#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=false
SKIP_DOCTOR=false
PREFIX="${HOME}/.local/bin"

log() {
  printf "%s\n" "$*"
}

warn_env_exec_issue() {
  warn "Installed binary was terminated by the OS from ${TARGET_BIN_PATH} (signal 9)."
  warn "This is usually an environment policy/path issue (not a build issue)."
  log "Try a different --prefix (e.g. /tmp/bin), check endpoint security, or run from repo binary."
}

ok() {
  printf "✅ %s\n" "$*"
}

warn() {
  printf "⚠️  %s\n" "$*"
}

fail() {
  printf "❌ %s\n" "$*" >&2
  exit 1
}

run() {
  if [ "$DRY_RUN" = "true" ]; then
    printf "[dry-run] %s\n" "$*"
    return 0
  fi
  "$@"
}

usage() {
  cat <<EOF
bitsentry-ai installer

Usage:
  ./install.sh [--dry-run] [--skip-doctor] [--prefix <path>] [--help]

Options:
  --dry-run      Show actions without executing them
  --skip-doctor  Skip post-install doctor command
  --prefix PATH  Install binary to PATH (default: ~/.local/bin)
  --help         Show this message
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      DRY_RUN=true
      ;;
    --skip-doctor)
      SKIP_DOCTOR=true
      ;;
    --prefix)
      shift
      [ "${1:-}" != "" ] || fail "--prefix requires a path"
      PREFIX="$1"
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "Unknown option: $1"
      ;;
  esac
  shift
done

OS="$(uname -s 2>/dev/null || printf "unknown")"
ARCH="$(uname -m 2>/dev/null || printf "unknown")"
OS_LABEL="unknown"

case "$OS" in
  Darwin) OS_LABEL="macOS" ;;
  Linux)
    if grep -qi microsoft /proc/version 2>/dev/null; then
      OS_LABEL="WSL"
    else
      OS_LABEL="Linux"
    fi
    ;;
esac

ARCH_LABEL="unknown"
case "$ARCH" in
  arm64|aarch64) ARCH_LABEL="arm64" ;;
  x86_64|amd64) ARCH_LABEL="amd64" ;;
esac

log "➡️  Installing bitsentry-ai"
ok "Detected OS: ${OS_LABEL}"
ok "Detected arch: ${ARCH_LABEL}"

for cmd in git curl go; do
  if command -v "$cmd" >/dev/null 2>&1; then
    ok "Dependency found: $cmd"
  else
    if [ "$cmd" = "go" ]; then
      fail "Go is required but was not found in PATH. Install Go first: https://go.dev/doc/install"
    fi
    fail "Required dependency missing: $cmd"
  fi
done

BIN_NAME="bitsentry-ai"
TARGET_BIN_DIR="$PREFIX"
TARGET_BIN_PATH="${TARGET_BIN_DIR}/${BIN_NAME}"
APP_DIR="${HOME}/.bitsentry-ai"
LOCAL_BUILD_BIN="./${BIN_NAME}"

run mkdir -p "$TARGET_BIN_DIR"
ok "Ensured binary directory: $TARGET_BIN_DIR"

run mkdir -p "$APP_DIR"
ok "Ensured app directory: $APP_DIR"

log "➡️  Building local binary"
if [ "$DRY_RUN" = "true" ]; then
  log "[dry-run] go build -o ${LOCAL_BUILD_BIN} ./cmd/bitsentry-ai"
else
  go build -o "$LOCAL_BUILD_BIN" ./cmd/bitsentry-ai
fi
ok "Build completed"

if [ -f "$TARGET_BIN_PATH" ]; then
  warn "Existing binary found at ${TARGET_BIN_PATH}; it will be replaced"
fi

run cp "$LOCAL_BUILD_BIN" "$TARGET_BIN_PATH"
run chmod +x "$TARGET_BIN_PATH"
ok "Installed binary to ${TARGET_BIN_PATH}"

case ":$PATH:" in
  *":${TARGET_BIN_DIR}:"*)
    ok "${TARGET_BIN_DIR} is already in PATH"
    ;;
  *)
    warn "${TARGET_BIN_DIR} is not in PATH"
    log "Add it with: export PATH=\"${TARGET_BIN_DIR}:\$PATH\""
    ;;
esac

if [ "$DRY_RUN" = "true" ]; then
  ok "Dry-run finished"
  exit 0
fi

log "➡️  Verifying installation"
set +e
"$TARGET_BIN_PATH" version
VERIFY_RC=$?
set -e

if [ "$VERIFY_RC" -eq 0 ]; then
  :
elif [ "$VERIFY_RC" -eq 137 ] || [ "$VERIFY_RC" -eq 9 ]; then
  warn_env_exec_issue
  if [ "$SKIP_DOCTOR" = "true" ]; then
    warn "Continuing because --skip-doctor was requested"
    ok "bitsentry-ai installation complete (with environment warning)"
    exit 0
  fi
  fail "Cannot verify binary execution from install path due to environment policy"
else
  fail "Post-install version check failed (exit ${VERIFY_RC})"
fi

if [ "$SKIP_DOCTOR" = "true" ]; then
  warn "Skipping doctor by request"
else
  "$TARGET_BIN_PATH" doctor
fi

log "➡️  Configuring bitsentry-dev preset"
"$TARGET_BIN_PATH" capabilities configure --preset bitsentry-dev

log "➡️  Exporting capability pack to OpenCode"
"$TARGET_BIN_PATH" capabilities export --target-agent opencode

ok "bitsentry-ai installation complete"
ok "OpenCode capability pack exported to:"
ok "  ~/.config/opencode/bitsentry/"
ok "  (or ~/.opencode/bitsentry/ if .config/opencode does not exist)"
log "Run 'bitsentry-ai' (no arguments) for TUI menu with full capability management."
