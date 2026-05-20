# Install Guide (Public MVP)

## Fast path (TUI-first, recommended)

From repository root:

```bash
chmod +x install.sh
./install.sh
bitsentry-ai
```

Then in TUI:
1. Open **Install / Setup**.
2. Complete the guided wizard.
3. Validate readiness verdict (`PASS`, `PASS WITH NOTES`, `FAIL`).

> MVP posture: guided/manual setup first. No live web execution, no `.env`/secrets handling.

## What installer does

1. Detects OS/arch (macOS/Linux/WSL where detectable).
2. Verifies required commands (`git`, `curl`, `go`).
3. Builds local binary: `go build -o bitsentry-ai ./cmd/bitsentry-ai`.
4. Creates `~/.local/bin` and `~/.bitsentry-ai` if missing.
5. Copies binary to `~/.local/bin/bitsentry-ai`.
6. Runs `bitsentry-ai version` and `bitsentry-ai doctor` (unless skipped).

## CLI support/debug path

CLI is for support/debug and diagnostics:

```bash
bitsentry-ai version
bitsentry-ai doctor
```

Useful install options:

```bash
./install.sh --dry-run
./install.sh --skip-doctor
./install.sh --prefix "$HOME/.local/bin"
```

## PATH troubleshooting

If installer warns that `~/.local/bin` is not in PATH, add it:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

Persist in your shell profile:
- `~/.zshrc` (zsh)
- `~/.bashrc` or `~/.bash_profile` (bash)
- `~/.config/fish/config.fish` (fish; syntax differs)

## Installed binary is killed under ~/.local/bin

### Symptom
`~/.local/bin/bitsentry-ai version` exits with signal 9 / RC -9, or is killed immediately.

### Likely cause
Restricted host/path-policy environments may block execution from `~/.local/bin`.

### Confirm quickly

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
./bitsentry-ai version
go build -o /tmp/bitsentry-ai-debug ./cmd/bitsentry-ai
/tmp/bitsentry-ai-debug version
```

### Workaround

```bash
./install.sh --prefix "$HOME/bin" --skip-doctor
export PATH="$HOME/bin:$PATH"
```

If it runs from `./bitsentry-ai` or `/tmp` but not `~/.local/bin`, the issue is environment policy, not BitsentryAI.

## Uninstall

```bash
rm -f "$HOME/.local/bin/bitsentry-ai"
rm -rf "$HOME/.bitsentry-ai"
```

Deleting `~/.bitsentry-ai` removes local config/logs.
