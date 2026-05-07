package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type openCodeBackupEntry struct {
	Timestamp  string
	BackupPath string
	FilePath   string
	Exists     bool
	ValidJSON  bool
	SizeBytes  int64
	Err        error
}

func newAgentsOpenCodeBackupsCmd(rt *Runtime) *cobra.Command {
	_ = rt
	return &cobra.Command{
		Use:   "backups",
		Short: "List OpenCode backups and JSON validity",
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := listOpenCodeBackups()
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode backups")
			if len(entries) == 0 {
				fmt.Fprintln(out, "No OpenCode backups found yet under ~/.bitsentry-ai/backups/opencode/.")
				return nil
			}

			for _, e := range entries {
				fmt.Fprintf(out, "- timestamp: %s\n", e.Timestamp)
				fmt.Fprintf(out, "  backup path: %s\n", e.BackupPath)
				fmt.Fprintf(out, "  opencode.json exists: %s\n", yesNo(e.Exists))
				if e.Exists {
					fmt.Fprintf(out, "  size: %d bytes\n", e.SizeBytes)
				} else {
					fmt.Fprintln(out, "  size: n/a")
				}
				fmt.Fprintf(out, "  valid json: %s\n", yesNo(e.ValidJSON))
				if e.Err != nil {
					fmt.Fprintf(out, "  note: %s\n", e.Err.Error())
				}
			}

			return nil
		},
	}
}

func newAgentsOpenCodeRestoreCmd(rt *Runtime) *cobra.Command {
	var backupTS string

	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore OpenCode config from backup safely",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			targetFile, _ := resolveOpenCodeJSONTarget(statusDetails)
			if strings.TrimSpace(targetFile) == "" {
				return fmt.Errorf("could not resolve target opencode.json; run 'bitsentry-ai agents opencode inspect-config' first")
			}

			selected, err := selectBackupEntry(strings.TrimSpace(backupTS))
			if err != nil {
				return err
			}

			if !selected.Exists {
				return fmt.Errorf("selected backup has no opencode.json: %s", selected.BackupPath)
			}
			if !selected.ValidJSON {
				return fmt.Errorf("selected backup JSON is invalid: %s", selected.FilePath)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve user home directory: %w", err)
			}
			preRestorePath := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode", time.Now().UTC().Format("20060102T150405Z")+"-pre-restore")

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] OpenCode restore preview")
				fmt.Fprintf(out, "- selected backup: %s\n", selected.FilePath)
				fmt.Fprintf(out, "- target file: %s\n", targetFile)
				fmt.Fprintf(out, "- pre-restore backup path (would be created): %s\n", preRestorePath)
				fmt.Fprintln(out, "- actions:")
				fmt.Fprintln(out, "  - validate selected backup JSON")
				fmt.Fprintln(out, "  - backup current target file to pre-restore path")
				fmt.Fprintln(out, "  - atomically replace target with backup JSON")
				fmt.Fprintln(out, "No files were modified.")
				return nil
			}

			rawTarget, targetMode, err := readTargetForPreRestore(targetFile)
			if err != nil {
				return err
			}

			if err := writePreRestoreBackup(preRestorePath, rawTarget); err != nil {
				return fmt.Errorf("pre-restore backup failed; target was not modified: %w", err)
			}

			rawBackup, err := os.ReadFile(selected.FilePath)
			if err != nil {
				return fmt.Errorf("read selected backup failed; target was not modified: %w", err)
			}

			if err := writeAtomicJSONWithMode(targetFile, rawBackup, targetMode); err != nil {
				return fmt.Errorf("restore write failed; pre-restore backup remains at %s: %w", preRestorePath, err)
			}

			fmt.Fprintln(out, "OpenCode restore")
			fmt.Fprintf(out, "- backup restored from: %s\n", selected.FilePath)
			fmt.Fprintf(out, "- pre-restore backup path: %s\n", preRestorePath)
			fmt.Fprintf(out, "- target file restored: %s\n", targetFile)
			fmt.Fprintln(out, "Restore complete. OpenCode config was restored from backup.")
			return nil
		},
	}

	cmd.Flags().StringVar(&backupTS, "backup", "", "Specific backup timestamp to restore (e.g. 20260506T190441Z)")
	return cmd
}

func listOpenCodeBackups() ([]openCodeBackupEntry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	baseDir := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode")
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []openCodeBackupEntry{}, nil
		}
		return nil, fmt.Errorf("read backup directory: %w", err)
	}

	result := make([]openCodeBackupEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ts := entry.Name()
		backupPath := filepath.Join(baseDir, ts)
		filePath := filepath.Join(backupPath, "opencode.json")

		b := openCodeBackupEntry{Timestamp: ts, BackupPath: backupPath, FilePath: filePath}
		st, statErr := os.Stat(filePath)
		if statErr != nil {
			if !errors.Is(statErr, os.ErrNotExist) {
				b.Err = fmt.Errorf("cannot stat opencode.json: %v", statErr)
			}
			result = append(result, b)
			continue
		}

		b.Exists = true
		b.SizeBytes = st.Size()
		raw, readErr := os.ReadFile(filePath)
		if readErr != nil {
			b.Err = fmt.Errorf("cannot read opencode.json: %v", readErr)
			result = append(result, b)
			continue
		}

		var root any
		if err := json.Unmarshal(raw, &root); err != nil {
			b.Err = fmt.Errorf("invalid json: %v", err)
			b.ValidJSON = false
		} else {
			b.ValidJSON = true
		}
		result = append(result, b)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	return result, nil
}

func selectBackupEntry(requestedTS string) (openCodeBackupEntry, error) {
	entries, err := listOpenCodeBackups()
	if err != nil {
		return openCodeBackupEntry{}, err
	}
	if len(entries) == 0 {
		return openCodeBackupEntry{}, fmt.Errorf("no OpenCode backups found under ~/.bitsentry-ai/backups/opencode/")
	}

	if requestedTS != "" {
		for _, entry := range entries {
			if entry.Timestamp == requestedTS {
				return entry, nil
			}
		}
		return openCodeBackupEntry{}, fmt.Errorf("backup timestamp not found: %s", requestedTS)
	}

	for _, entry := range entries {
		if entry.Exists && entry.ValidJSON {
			return entry, nil
		}
	}

	return openCodeBackupEntry{}, fmt.Errorf("no valid backup found (all candidates missing opencode.json or invalid JSON)")
}

func readTargetForPreRestore(targetFile string) ([]byte, fs.FileMode, error) {
	raw, err := os.ReadFile(targetFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, 0, fmt.Errorf("target file missing: %s", targetFile)
		}
		return nil, 0, fmt.Errorf("read target file: %w", err)
	}

	st, err := os.Stat(targetFile)
	if err != nil {
		return nil, 0, fmt.Errorf("stat target file: %w", err)
	}

	return raw, st.Mode().Perm(), nil
}

func writePreRestoreBackup(preRestorePath string, rawTarget []byte) error {
	if err := os.MkdirAll(preRestorePath, 0o755); err != nil {
		return fmt.Errorf("create pre-restore backup directory: %w", err)
	}

	backupFile := filepath.Join(preRestorePath, "opencode.json")
	if err := os.WriteFile(backupFile, rawTarget, 0o644); err != nil {
		return fmt.Errorf("write pre-restore backup file: %w", err)
	}

	return nil
}

func writeAtomicJSONWithMode(targetFile string, payload []byte, mode fs.FileMode) error {
	if !json.Valid(payload) {
		return fmt.Errorf("restore payload is not valid JSON")
	}

	dir := filepath.Dir(targetFile)
	tmpFile, err := os.CreateTemp(dir, "opencode.json.restore.tmp-*")
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

	if mode != 0 {
		if err := os.Chmod(tmpName, mode); err != nil {
			return fmt.Errorf("set temp file mode: %w", err)
		}
	}

	if err := os.Rename(tmpName, targetFile); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	cleanup = false
	return nil
}
