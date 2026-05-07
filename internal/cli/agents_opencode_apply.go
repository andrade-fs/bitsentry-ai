package cli

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
	"github.com/spf13/cobra"
)

func newAgentsOpenCodeApplyCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply managed OpenCode MCP entries with backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			targetFile, _ := resolveOpenCodeJSONTarget(statusDetails)
			if strings.TrimSpace(targetFile) == "" {
				return fmt.Errorf("could not resolve target opencode.json; run 'bitsentry-ai agents opencode inspect-config' and 'bitsentry-ai agents opencode export' first")
			}

			rawOriginal, err := os.ReadFile(targetFile)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("target file missing: %s\nSuggestion: run 'bitsentry-ai agents opencode inspect-config' and export/patch-plan first", targetFile)
				}
				return fmt.Errorf("read target opencode.json: %w", err)
			}

			var root map[string]any
			if err := json.Unmarshal(rawOriginal, &root); err != nil {
				return fmt.Errorf("parse target opencode.json failed (no writes performed): %w", err)
			}

			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			_, mcpSummary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)

			topLevelKeysBefore := sortedKeys(root)
			existingMCP := map[string]any{}
			if existing, ok := root["mcp"]; ok {
				existingMap, ok := existing.(map[string]any)
				if !ok {
					return fmt.Errorf("target opencode.json has top-level 'mcp' that is not an object; refusing to write")
				}
				existingMCP = cloneMap(existingMap)
			}

			merged := cloneMap(existingMCP)
			managedValues, updatedIDs := buildManagedOpenCodeMCPValues(cfg, mcpSummary.Selected)
			for id, value := range managedValues {
				merged[id] = value
			}

			preserved := make([]string, 0)
			for id := range existingMCP {
				if _, managed := managedValues[id]; !managed {
					preserved = append(preserved, id)
				}
			}
			sort.Strings(preserved)

			patchedRoot := cloneMap(root)
			patchedRoot["mcp"] = merged

			patchedBytes, err := json.MarshalIndent(patchedRoot, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal patched opencode.json: %w", err)
			}
			patchedBytes = append(patchedBytes, '\n')

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] OpenCode MCP apply preview")
				fmt.Fprintf(out, "- target file: %s\n", targetFile)
				fmt.Fprintf(out, "- file checksum (before): %s\n", sha256Hex(rawOriginal))
				fmt.Fprintf(out, "- file mtime (before): %s\n", fileModTime(targetFile))
				fmt.Fprintf(out, "- top-level keys preserved: %s\n", fallback(strings.Join(topLevelKeysBefore, ", "), "none"))
				fmt.Fprintf(out, "- mcp entries to update: %s\n", strings.Join(updatedIDs, ", "))
				fmt.Fprintf(out, "- mcp entries preserved: %s\n", fallback(strings.Join(preserved, ", "), "none"))
				fmt.Fprintln(out, "- skills skipped: true")
				fmt.Fprintln(out, "No backup was written. Target file was not modified.")
				return nil
			}

			backupPath, err := backupOpenCodeJSON(rawOriginal)
			if err != nil {
				return fmt.Errorf("backup failed (no writes performed): %w", err)
			}

			if err := writeAtomicJSON(targetFile, patchedBytes); err != nil {
				return fmt.Errorf("write failed; backup remains at %s: %w", backupPath, err)
			}

			fmt.Fprintln(out, "OpenCode MCP apply")
			fmt.Fprintf(out, "- backup path: %s\n", backupPath)
			fmt.Fprintf(out, "- target file modified: %s\n", targetFile)
			fmt.Fprintf(out, "- mcp entries updated: %s\n", strings.Join(updatedIDs, ", "))
			fmt.Fprintf(out, "- mcp entries preserved: %s\n", fallback(strings.Join(preserved, ", "), "none"))
			fmt.Fprintln(out, "- skills skipped: true")
			fmt.Fprintln(out, "Apply complete. OpenCode MCP config was updated with backup.")
			return nil
		},
	}
}

func buildManagedOpenCodeMCPValues(cfg config.Config, selected []string) (map[string]any, []string) {
	selectedSet := map[string]bool{}
	for _, id := range selected {
		selectedSet[strings.TrimSpace(id)] = true
	}

	context7Command := strings.TrimSpace(cfg.Components.Context7.Command)
	if context7Command == "" {
		context7Command = "npx"
	}
	context7Package := strings.TrimSpace(cfg.Components.Context7.Package)
	if context7Package == "" {
		context7Package = "@upstash/context7-mcp"
	}

	context7Value := map[string]any{
		"selected": selectedSet["context7"],
		"enabled":  cfg.Components.Context7.Enabled,
		"command":  context7Command,
		"args":     []string{"-y", context7Package},
	}

	engramValue := map[string]any{
		"selected":   selectedSet["engram"],
		"enabled":    cfg.Components.Engram.Enabled,
		"configured": cfg.Components.Engram.Configured,
		"command":    "engram",
	}
	if p := strings.TrimSpace(cfg.Components.Engram.BinaryPath); p != "" {
		engramValue["binary_path"] = p
	}
	if p := strings.TrimSpace(cfg.Components.Engram.Project); p != "" {
		engramValue["project"] = p
	}

	return map[string]any{
		"context7": context7Value,
		"engram":   engramValue,
	}, []string{"context7", "engram"}
}

func writeAtomicJSON(targetFile string, payload []byte) error {
	dir := filepath.Dir(targetFile)
	tmpFile, err := os.CreateTemp(dir, "opencode.json.tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmpFile.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()

	if _, err := tmpFile.Write(payload); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, targetFile); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	cleanup = false
	return nil
}

func backupOpenCodeJSON(rawOriginal []byte) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	backupDir := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode", time.Now().UTC().Format("20060102T150405Z"))
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}

	backupFile := filepath.Join(backupDir, "opencode.json")
	if err := os.WriteFile(backupFile, rawOriginal, 0o644); err != nil {
		return "", fmt.Errorf("write backup file: %w", err)
	}

	return backupDir, nil
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func sha256Hex(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func fileModTime(path string) string {
	st, err := os.Stat(path)
	if err != nil {
		return "unknown"
	}
	return st.ModTime().UTC().Format(time.RFC3339Nano)
}
