# Install Guide

## Local install (recommended)

From repository root:

```bash
chmod +x install.sh
./install.sh
```

The installer will:
1. Detect OS/arch (macOS/Linux/WSL where detectable)
2. Verify required commands (`git`, `curl`, `go`)
3. Build local binary via `go build -o bitsentry-ai ./cmd/bitsentry-ai`
4. Create `~/.local/bin` and `~/.bitsentry-ai` if missing
5. Copy binary to `~/.local/bin/bitsentry-ai`
6. Run `bitsentry-ai version` and `bitsentry-ai doctor` (unless skipped)

## Options

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

Persist it by adding the line to your shell profile:
- `~/.zshrc` (zsh)
- `~/.bashrc` or `~/.bash_profile` (bash)
- `~/.config/fish/config.fish` (fish, syntax differs)

## Uninstall

```bash
rm -f "$HOME/.local/bin/bitsentry-ai"
rm -rf "$HOME/.bitsentry-ai"
```

> Note: deleting `~/.bitsentry-ai` removes local config/logs for this tool.

## Development build

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
./bitsentry-ai version
./bitsentry-ai doctor
go test ./...
```

## Installed binary is killed under ~/.local/bin

### Symptom

`~/.local/bin/bitsentry-ai version` exits with signal 9 / RC -9, or appears to be killed immediately.

### Explanation

This can happen in restricted host environments or path-policy setups where binaries executed from `~/.local/bin` are blocked or killed.

### How to confirm

Build and run from project directory:

```bash
go build -o bitsentry-ai ./cmd/bitsentry-ai
./bitsentry-ai version
```

Build and run from `/tmp`:

```bash
go build -o /tmp/bitsentry-ai-debug ./cmd/bitsentry-ai
/tmp/bitsentry-ai-debug version
```

### Workaround

Install to an alternate prefix:

```bash
./install.sh --prefix "$HOME/bin" --skip-doctor
```

Then add it to PATH:

```bash
export PATH="$HOME/bin:$PATH"
```

### Note

If the binary works from `./bitsentry-ai` or `/tmp` but not from `~/.local/bin`, the issue is likely environment/path policy, not `bitsentry-ai` itself.
