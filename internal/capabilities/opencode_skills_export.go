package capabilities

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type SkillsExportResult struct {
	DryRun         bool
	Status         string
	TargetRoot     string
	SelectedIDs    []string
	IncludedFlows  []string
	IncludedPacks  []string
	IncludedSkills int
	GeneratedFiles []string
	WrittenFiles   []string
	Warnings       []string
	Skipped        []string
	BackupPath     string
}

func ExecuteOpenCodeSkillsExport(projection OpenCodeExportProjection, targetConfigRoot string, dryRun bool) (SkillsExportResult, error) {
	managedRoot := strings.TrimSpace(targetConfigRoot)
	root := filepath.Join(managedRoot, "bitsentry")
	if strings.TrimSpace(targetConfigRoot) == "" {
		return SkillsExportResult{}, fmt.Errorf("target config root is required")
	}
	result := SkillsExportResult{
		DryRun:         dryRun,
		Status:         "preview",
		TargetRoot:     root,
		SelectedIDs:    append([]string{}, projection.SelectedIDs...),
		IncludedFlows:  idsFromProjectedFlows(projection.IncludedFlows),
		IncludedPacks:  idsFromProjectedPacks(projection.IncludedSkillPacks),
		IncludedSkills: len(projection.IncludedSkills),
		Warnings:       append([]string{}, projection.Warnings...),
		Skipped:        append([]string{}, projection.Skipped...),
	}

	copyPairs := make([][2]string, 0)
	for _, s := range projection.IncludedSharedContracts {
		copyPairs = append(copyPairs, [2]string{s.SourcePath, s.TargetPath})
	}
	for _, f := range projection.IncludedFlows {
		copyPairs = append(copyPairs, [2]string{f.SourcePath, f.TargetPath})
	}
	for _, s := range projection.IncludedSkills {
		if s.Status != "valid" {
			return result, fmt.Errorf("projection contains invalid skill %q; refusing export", s.ID)
		}
		copyPairs = append(copyPairs, [2]string{s.SourcePath, s.TargetPath})
	}
	for _, o := range projection.IncludedOrchestrators {
		copyPairs = append(copyPairs, [2]string{o.SourcePath, o.TargetPath})
	}
	for _, g := range projection.GeneratedFiles {
		result.GeneratedFiles = append(result.GeneratedFiles, g.Path)
	}

	if dryRun {
		for _, p := range copyPairs {
			if err := validateTargetRelativePath(p[1]); err != nil {
				return result, err
			}
			result.WrittenFiles = append(result.WrittenFiles, p[1])
		}
		for _, g := range projection.GeneratedFiles {
			if err := validateTargetRelativePath(g.Path); err != nil {
				return result, err
			}
			result.WrittenFiles = append(result.WrittenFiles, g.Path)
		}
		sort.Strings(result.WrittenFiles)
		return result, nil
	}

	backupPath, err := backupOpenCodeBitsentryDir(root)
	if err != nil {
		result.Status = "failed"
		return result, err
	}
	result.BackupPath = backupPath

	for _, pair := range copyPairs {
		if err := copyProjectedFile(managedRoot, pair[0], pair[1]); err != nil {
			result.Status = "partial"
			return result, err
		}
		result.WrittenFiles = append(result.WrittenFiles, pair[1])
	}
	for _, g := range projection.GeneratedFiles {
		if err := writeGeneratedFile(managedRoot, g.Path, g.Content); err != nil {
			result.Status = "partial"
			return result, err
		}
		result.WrittenFiles = append(result.WrittenFiles, g.Path)
	}
	sort.Strings(result.WrittenFiles)
	result.Status = "exported"
	return result, nil
}

func copyProjectedFile(targetRoot string, sourcePath string, targetRel string) error {
	if err := validateTargetRelativePath(targetRel); err != nil {
		return err
	}
	raw, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read source %s: %w", sourcePath, err)
	}
	target := filepath.Join(targetRoot, filepath.FromSlash(targetRel))
	if err := ensurePathUnderRoot(targetRoot, target); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("mkdir target: %w", err)
	}
	if err := os.WriteFile(target, raw, 0o644); err != nil {
		return fmt.Errorf("write target %s: %w", target, err)
	}
	return nil
}

func writeGeneratedFile(targetRoot string, targetRel string, content string) error {
	if err := validateTargetRelativePath(targetRel); err != nil {
		return err
	}
	target := filepath.Join(targetRoot, filepath.FromSlash(targetRel))
	if err := ensurePathUnderRoot(targetRoot, target); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("mkdir generated target: %w", err)
	}
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write generated file %s: %w", target, err)
	}
	return nil
}

func validateTargetRelativePath(targetRel string) error {
	t := filepath.ToSlash(filepath.Clean(strings.TrimSpace(targetRel)))
	if strings.HasPrefix(t, "../") || strings.Contains(t, "/../") || strings.HasPrefix(t, "/") {
		return fmt.Errorf("unsafe target relative path: %s", targetRel)
	}
	if !strings.HasPrefix(t, "bitsentry/") {
		return fmt.Errorf("target path must stay under bitsentry/: %s", targetRel)
	}
	return nil
}

func ensurePathUnderRoot(root, target string) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	if absTarget != absRoot && !strings.HasPrefix(absTarget, absRoot+string(filepath.Separator)) {
		return fmt.Errorf("target escapes managed root: %s", target)
	}
	return nil
}

func backupOpenCodeBitsentryDir(bitsentryRoot string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	backupDir := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode-skills", time.Now().UTC().Format("20060102T150405Z"))
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup directory: %w", err)
	}
	st, err := os.Stat(bitsentryRoot)
	if err == nil && st.IsDir() {
		if err := copyDir(bitsentryRoot, filepath.Join(backupDir, "bitsentry")); err != nil {
			return "", fmt.Errorf("backup managed bitsentry dir: %w", err)
		}
	}
	return backupDir, nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}

func OpenCodeSkillsReportDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".bitsentry-ai", "exports", "opencode-skills"), nil
}

func WriteOpenCodeSkillsExportReport(result SkillsExportResult) (string, error) {
	dir, err := OpenCodeSkillsReportDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create report directory: %w", err)
	}
	ts := time.Now().UTC().Format("20060102T150405Z")
	path := filepath.Join(dir, fmt.Sprintf("%s-export.yaml", ts))
	payload := map[string]any{
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
		"dry_run":         result.DryRun,
		"status":          result.Status,
		"selected_ids":    result.SelectedIDs,
		"target_root":     result.TargetRoot,
		"included_flows":  result.IncludedFlows,
		"included_packs":  result.IncludedPacks,
		"included_skills": result.IncludedSkills,
		"generated_files": result.GeneratedFiles,
		"written_files":   result.WrittenFiles,
		"warnings":        result.Warnings,
		"skipped":         result.Skipped,
		"backup_path":     result.BackupPath,
	}
	raw, err := yaml.Marshal(payload)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", err
	}
	latest := filepath.Join(dir, "latest.yaml")
	if err := os.WriteFile(latest, raw, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func idsFromProjectedFlows(flows []ProjectedFlow) []string {
	out := make([]string, 0, len(flows))
	for _, f := range flows {
		out = append(out, f.ID)
	}
	sort.Strings(out)
	return out
}

func idsFromProjectedPacks(packs []ProjectedSkillPack) []string {
	out := make([]string, 0, len(packs))
	for _, p := range packs {
		out = append(out, p.ID)
	}
	sort.Strings(out)
	return out
}
