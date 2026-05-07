package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/components"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newAgentsCmd(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "List detected AI agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := rt.App.AgentRegistry.List(context.Background())
			if err != nil {
				return fmt.Errorf("detect agents: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Agents")
			for _, r := range results {
				status := "not detected"
				if r.Found {
					status = "detected"
				}
				fmt.Fprintf(out, "- %s (%s): %s\n", r.Name, r.ID, status)
				if r.Path != "" {
					fmt.Fprintf(out, "  path: %s\n", r.Path)
				}
				if r.Version != "" {
					fmt.Fprintf(out, "  version: %s\n", r.Version)
				}
				if r.Hint != "" {
					fmt.Fprintf(out, "  hint: %s\n", r.Hint)
				}
			}

			return nil
		},
	}

	opencodeCmd := &cobra.Command{Use: "opencode", Short: "Inspect OpenCode integration status and export preview"}
	opencodeCmd.AddCommand(
		newAgentsOpenCodeStatusCmd(rt),
		newAgentsOpenCodeExportPreviewCmd(rt),
		newAgentsOpenCodeExportCmd(rt),
		newAgentsOpenCodeApplyPlanCmd(rt),
		newAgentsOpenCodeInspectConfigCmd(rt),
		newAgentsOpenCodePatchPlanCmd(rt),
		newAgentsOpenCodeBackupsCmd(rt),
		newAgentsOpenCodeRestoreCmd(rt),
		newAgentsOpenCodeApplyCmd(rt),
	)

	cmd.AddCommand(opencodeCmd)
	return cmd
}

func newAgentsOpenCodeExportCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Generate OpenCode export files in bitsentry-ai export directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, mcpSummary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, skillsSummary := components.SkillsRegistry(cfg)

			plan := agents.OpenCodeExportPlan{
				TargetAgent:      "opencode",
				TargetConfigPath: statusDetails.TargetConfigCandidate,
				Memory: agents.OpenCodeMemoryPlan{
					EngramEnabled:    cfg.Components.Engram.Enabled,
					EngramConfigured: cfg.Components.Engram.Configured,
					EngramBinaryPath: cfg.Components.Engram.BinaryPath,
					EngramProject:    cfg.Components.Engram.Project,
					Context7Enabled:  cfg.Components.Context7.Enabled,
					Context7Command:  cfg.Components.Context7.Command,
					Context7Package:  cfg.Components.Context7.Package,
				},
				Warnings: append([]string{}, statusDetails.Warnings...),
				Actions: []string{
					"Generate reviewable export artifacts under ~/.bitsentry-ai/exports/opencode/",
					"Do not modify OpenCode config files",
					"Require explicit future apply step with backup before touching OpenCode",
				},
			}

			for _, mcp := range mcpEntries {
				plan.MCPs = append(plan.MCPs, agents.OpenCodeMCPPlan{
					ID:         mcp.ID,
					Selected:   contains(mcpSummary.Selected, mcp.ID),
					Configured: mcp.Configured,
					Status:     string(mcp.Status),
				})
			}
			for _, skill := range skillEntries {
				plan.Skills = append(plan.Skills, agents.OpenCodeSkillPlan{
					ID:         skill.ID,
					Selected:   contains(skillsSummary.Selected, skill.ID),
					Configured: skill.Configured,
					Status:     string(skill.Status),
				})
			}

			if strings.TrimSpace(plan.TargetConfigPath) == "" {
				plan.Warnings = append(plan.Warnings, "No OpenCode config path candidate resolved.")
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve user home directory: %w", err)
			}

			exportDir := filepath.Join(homeDir, ".bitsentry-ai", "exports", "opencode")
			exportPlanPath := filepath.Join(exportDir, "export-plan.yaml")
			readmePath := filepath.Join(exportDir, "README.md")
			snippetPath := filepath.Join(exportDir, "opencode-snippet.yaml")

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] OpenCode export preview (file generation)")
				fmt.Fprintln(out, "No files were written.")
				fmt.Fprintln(out, "Files that would be written:")
				fmt.Fprintf(out, "- %s\n", exportPlanPath)
				fmt.Fprintf(out, "- %s\n", readmePath)
				fmt.Fprintf(out, "- %s\n", snippetPath)
				fmt.Fprintln(out, "Safety: writes are restricted to ~/.bitsentry-ai/exports/opencode/ only.")
				return nil
			}

			if err := os.MkdirAll(exportDir, 0o755); err != nil {
				return fmt.Errorf("create export directory: %w", err)
			}

			exportPayload := map[string]any{
				"target_agent":                 plan.TargetAgent,
				"generated_at":                 time.Now().UTC().Format(time.RFC3339),
				"target_config_path_candidate": plan.TargetConfigPath,
				"mcps":                         plan.MCPs,
				"skills":                       plan.Skills,
				"engram_config_summary":        map[string]any{"status": engramDetails.Status, "enabled": plan.Memory.EngramEnabled, "configured": plan.Memory.EngramConfigured, "binary_path": plan.Memory.EngramBinaryPath, "project": plan.Memory.EngramProject},
				"context7_config_summary":      map[string]any{"status": context7Details.Status, "enabled": plan.Memory.Context7Enabled, "command": plan.Memory.Context7Command, "package": plan.Memory.Context7Package},
				"warnings":                     plan.Warnings,
				"planned_actions":              plan.Actions,
				"safety":                       "No OpenCode files were modified. This export is review-only.",
			}

			yamlBytes, err := yaml.Marshal(exportPayload)
			if err != nil {
				return fmt.Errorf("marshal export plan yaml: %w", err)
			}

			if err := os.WriteFile(exportPlanPath, yamlBytes, 0o644); err != nil {
				return fmt.Errorf("write export plan: %w", err)
			}

			readme := strings.TrimSpace(fmt.Sprintf(`# OpenCode Export Artifacts (Review Only)

This directory was generated by:

- bitsentry-ai agents opencode export

## What was generated

- export-plan.yaml — plan model derived from bitsentry-ai config and runtime metadata
- opencode-snippet.yaml — illustrative, non-authoritative snippet for review only

## Safety and scope

- Review-only export: generated artifacts are intended for inspection
- No OpenCode config was modified
- No MCP servers were executed
- Writes are restricted to: %s

## Future apply step

If the plan is approved, a separate explicit apply command should:

1. Create a backup of existing OpenCode config
2. Validate target paths and schema
3. Apply changes with confirmation
`, exportDir)) + "\n"

			if err := os.WriteFile(readmePath, []byte(readme), 0o644); err != nil {
				return fmt.Errorf("write readme: %w", err)
			}

			snippet := strings.TrimSpace(`# Non-authoritative illustrative snippet
# Schema may differ from your installed OpenCode version.
agent: opencode
mcps:
  # Example only. Validate fields against official OpenCode docs.
  - id: engram
    enabled: true
skills:
  # Example only. Validate fields against official OpenCode docs.
  - id: bitsentry-sdd
    enabled: true
`) + "\n"

			if err := os.WriteFile(snippetPath, []byte(snippet), 0o644); err != nil {
				return fmt.Errorf("write snippet: %w", err)
			}

			fmt.Fprintln(out, "OpenCode export generated successfully.")
			fmt.Fprintf(out, "- export directory: %s\n", exportDir)
			fmt.Fprintf(out, "- file: %s\n", exportPlanPath)
			fmt.Fprintf(out, "- file: %s\n", readmePath)
			fmt.Fprintf(out, "- file: %s\n", snippetPath)
			fmt.Fprintln(out, "Safety: No OpenCode files were modified.")
			return nil
		},
	}
}

func newAgentsOpenCodeStatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show OpenCode detection and config path status",
		RunE: func(cmd *cobra.Command, args []string) error {
			details, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode status")
			fmt.Fprintf(out, "- binary detected: %s\n", yesNo(details.Found))
			fmt.Fprintf(out, "- binary path: %s\n", fallback(details.Path, "not found"))
			fmt.Fprintf(out, "- version: %s\n", fallback(details.Version, "unavailable"))
			fmt.Fprintf(out, "- can export config safely: %s\n", yesNo(details.CanExportSafely))
			fmt.Fprintf(out, "- target config candidate: %s\n", fallback(details.TargetConfigCandidate, "none"))

			fmt.Fprintln(out, "- config paths found:")
			if len(details.ConfigPathsFound) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, p := range details.ConfigPathsFound {
					fmt.Fprintf(out, "  - %s\n", p)
				}
			}

			if len(details.Notes) > 0 {
				fmt.Fprintln(out, "- notes:")
				for _, n := range details.Notes {
					fmt.Fprintf(out, "  - %s\n", n)
				}
			}

			return nil
		},
	}
}

func newAgentsOpenCodeApplyPlanCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "apply-plan",
		Short: "Prepare OpenCode apply plan without modifying files",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			_, mcpSummary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			_, skillsSummary := components.SkillsRegistry(cfg)

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve user home directory: %w", err)
			}

			exportDir := filepath.Join(homeDir, ".bitsentry-ai", "exports", "opencode")
			exportPlanPath := filepath.Join(exportDir, "export-plan.yaml")
			exportReadmePath := filepath.Join(exportDir, "README.md")
			backupDir := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode", time.Now().UTC().Format("20060102T150405Z"))

			filesToRead := []string{}
			if strings.TrimSpace(statusDetails.TargetConfigCandidate) != "" {
				filesToRead = append(filesToRead, statusDetails.TargetConfigCandidate)
			}
			filesToRead = append(filesToRead, exportPlanPath, exportReadmePath)

			filesToWriteFutureApply := []string{}
			if strings.TrimSpace(statusDetails.TargetConfigCandidate) != "" {
				filesToWriteFutureApply = append(filesToWriteFutureApply, statusDetails.TargetConfigCandidate)
			}

			filesToBackupFutureApply := []string{}
			if strings.TrimSpace(statusDetails.TargetConfigCandidate) != "" && len(statusDetails.ConfigPathsFound) > 0 {
				filesToBackupFutureApply = append(filesToBackupFutureApply, statusDetails.TargetConfigCandidate)
			}

			warnings := append([]string{}, statusDetails.Warnings...)
			warnings = append(warnings,
				"Read-only command: no directory creation and no file writes.",
				"OpenCode config is not modified in apply-plan.",
				"MCP servers are not executed.",
				"Skills are not copied to OpenCode.",
			)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode apply plan")
			fmt.Fprintln(out, "- target agent: opencode")
			fmt.Fprintf(out, "- target config path candidate: %s\n", fallback(statusDetails.TargetConfigCandidate, "none"))
			fmt.Fprintf(out, "- target exists: %s\n", yesNo(len(statusDetails.ConfigPathsFound) > 0))
			fmt.Fprintf(out, "- export directory path: %s\n", exportDir)
			fmt.Fprintf(out, "- export files available: %s\n", yesNo(fileExists(exportPlanPath) && fileExists(exportReadmePath)))
			fmt.Fprintf(out, "- backup directory that would be used: %s\n", backupDir)

			fmt.Fprintln(out, "- files that would be read:")
			for _, file := range filesToRead {
				fmt.Fprintf(out, "  - %s\n", file)
			}

			fmt.Fprintln(out, "- files that would be written in future apply:")
			if len(filesToWriteFutureApply) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, file := range filesToWriteFutureApply {
					fmt.Fprintf(out, "  - %s\n", file)
				}
			}

			fmt.Fprintln(out, "- files that would be backed up in future apply:")
			if len(filesToBackupFutureApply) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, file := range filesToBackupFutureApply {
					fmt.Fprintf(out, "  - %s\n", file)
				}
			}

			fmt.Fprintf(out, "- selected MCPs: %s\n", fallback(strings.Join(mcpSummary.Selected, ", "), "none"))
			fmt.Fprintf(out, "- selected Skills: %s\n", fallback(strings.Join(skillsSummary.Selected, ", "), "none"))

			if len(warnings) > 0 {
				fmt.Fprintln(out, "- warnings:")
				for _, w := range warnings {
					fmt.Fprintf(out, "  - %s\n", w)
				}
			}

			fmt.Fprintln(out, "Plan only. No OpenCode files were modified.")
			return nil
		},
	}
}

func newAgentsOpenCodeExportPreviewCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "export-preview",
		Short: "Preview OpenCode export plan without writing files",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, mcpSummary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, skillsSummary := components.SkillsRegistry(cfg)

			plan := agents.OpenCodeExportPlan{
				TargetAgent:      "opencode",
				TargetConfigPath: statusDetails.TargetConfigCandidate,
				Memory: agents.OpenCodeMemoryPlan{
					EngramEnabled:    cfg.Components.Engram.Enabled,
					EngramConfigured: cfg.Components.Engram.Configured,
					EngramBinaryPath: cfg.Components.Engram.BinaryPath,
					EngramProject:    cfg.Components.Engram.Project,
					Context7Enabled:  cfg.Components.Context7.Enabled,
					Context7Command:  cfg.Components.Context7.Command,
					Context7Package:  cfg.Components.Context7.Package,
				},
				Warnings: append([]string{}, statusDetails.Warnings...),
				Actions: []string{
					"Prepare OpenCode config patch in-memory only",
					"Do not write OpenCode files in this phase",
					"Keep export/apply operations as TODO until a later phase",
				},
			}

			for _, mcp := range mcpEntries {
				plan.MCPs = append(plan.MCPs, agents.OpenCodeMCPPlan{
					ID:         mcp.ID,
					Selected:   contains(mcpSummary.Selected, mcp.ID),
					Configured: mcp.Configured,
					Status:     string(mcp.Status),
				})
			}
			for _, skill := range skillEntries {
				plan.Skills = append(plan.Skills, agents.OpenCodeSkillPlan{
					ID:         skill.ID,
					Selected:   contains(skillsSummary.Selected, skill.ID),
					Configured: skill.Configured,
					Status:     string(skill.Status),
				})
			}

			if strings.TrimSpace(plan.TargetConfigPath) == "" {
				plan.Warnings = append(plan.Warnings, "No OpenCode config path candidate resolved.")
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode export preview")
			fmt.Fprintf(out, "- target agent: %s\n", plan.TargetAgent)
			fmt.Fprintf(out, "- target config path candidate: %s\n", fallback(plan.TargetConfigPath, "none"))
			fmt.Fprintf(out, "- opencode binary detected: %s\n", yesNo(statusDetails.Found))
			fmt.Fprintf(out, "- selected MCPs from bitsentry config: %s\n", fallback(strings.Join(mcpSummary.Selected, ", "), "none"))
			fmt.Fprintf(out, "- selected Skills from bitsentry config: %s\n", fallback(strings.Join(skillsSummary.Selected, ", "), "none"))
			fmt.Fprintf(out, "- engram status/config: status=%s enabled=%s configured=%s binary=%s project=%s\n",
				engramDetails.Status,
				yesNo(plan.Memory.EngramEnabled),
				yesNo(plan.Memory.EngramConfigured),
				fallback(plan.Memory.EngramBinaryPath, "n/a"),
				fallback(plan.Memory.EngramProject, "n/a"),
			)
			fmt.Fprintf(out, "- context7 status/config: status=%s enabled=%s command=%s package=%s\n",
				context7Details.Status,
				yesNo(plan.Memory.Context7Enabled),
				fallback(plan.Memory.Context7Command, "n/a"),
				fallback(plan.Memory.Context7Package, "n/a"),
			)

			if len(plan.Warnings) > 0 {
				fmt.Fprintln(out, "- warnings:")
				for _, w := range plan.Warnings {
					fmt.Fprintf(out, "  - %s\n", w)
				}
			}

			fmt.Fprintln(out, "- actions:")
			for _, action := range plan.Actions {
				fmt.Fprintf(out, "  - %s\n", action)
			}

			fmt.Fprintln(out, "Preview only. No OpenCode files were modified.")
			return nil
		},
	}
}

func newAgentsOpenCodeInspectConfigCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect-config",
		Short: "Inspect OpenCode config schema/format without modifying files",
		RunE: func(cmd *cobra.Command, args []string) error {
			details, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			wd, wdErr := os.Getwd()
			if wdErr != nil {
				wd = ""
			}

			candidates := agents.OpenCodeConfigPathCandidates(wd)
			if strings.TrimSpace(details.TargetConfigCandidate) != "" && !contains(candidates, details.TargetConfigCandidate) {
				candidates = append(candidates, details.TargetConfigCandidate)
			}

			found := make([]string, 0)
			for _, path := range candidates {
				if st, statErr := os.Stat(path); statErr == nil && st.IsDir() {
					found = append(found, path)
				}
			}

			if len(found) == 0 {
				found = append(found, details.ConfigPathsFound...)
			}

			allWarnings := append([]string{}, details.Warnings...)
			inspections := make([]agents.OpenCodeConfigFileInspection, 0)
			for _, dir := range found {
				files, warnings := agents.FindLikelyOpenCodeConfigFiles(dir)
				allWarnings = append(allWarnings, warnings...)
				for _, file := range files {
					inspections = append(inspections, agents.InspectOpenCodeConfigFile(file))
				}
			}

			formatSet := map[string]struct{}{}
			for _, in := range inspections {
				formatSet[in.Type] = struct{}{}
				if strings.TrimSpace(in.ParseWarning) != "" {
					allWarnings = append(allWarnings, fmt.Sprintf("%s: %s", in.Path, in.ParseWarning))
				}
			}

			formats := make([]string, 0, len(formatSet))
			for f := range formatSet {
				formats = append(formats, f)
			}
			sort.Strings(formats)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode config inspection")
			fmt.Fprintf(out, "- opencode binary detected: %s\n", yesNo(details.Found))
			fmt.Fprintf(out, "- opencode binary path: %s\n", fallback(details.Path, "not found"))
			fmt.Fprintf(out, "- opencode version: %s\n", fallback(details.Version, "unavailable"))

			fmt.Fprintln(out, "- config directories checked:")
			if len(candidates) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, path := range candidates {
					fmt.Fprintf(out, "  - %s\n", path)
				}
			}

			fmt.Fprintln(out, "- config directories found:")
			if len(found) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, path := range found {
					fmt.Fprintf(out, "  - %s\n", path)
				}
			}

			fmt.Fprintln(out, "- likely config files found:")
			if len(inspections) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, in := range inspections {
					fmt.Fprintf(out, "  - path: %s\n", in.Path)
					fmt.Fprintf(out, "    type: %s\n", in.Type)
					fmt.Fprintf(out, "    size_bytes: %d\n", in.SizeBytes)
					fmt.Fprintf(out, "    readable: %s\n", yesNo(in.Readable))
					fmt.Fprintf(out, "    top_level_keys: %s\n", fallback(strings.Join(in.TopLevelKeys, ", "), "none or non-object root"))
				}
			}

			fmt.Fprintf(out, "- inferred formats: %s\n", fallback(strings.Join(formats, ", "), "unknown"))

			if len(allWarnings) > 0 {
				fmt.Fprintln(out, "- warnings:")
				for _, warning := range allWarnings {
					fmt.Fprintf(out, "  - %s\n", warning)
				}
			}

			fmt.Fprintln(out, "- recommendation:")
			fmt.Fprintln(out, "  - Use this inspection to define a schema-aware apply phase that targets only known top-level keys and validates format compatibility before any write.")
			fmt.Fprintln(out, "  - Keep apply behind explicit confirmation and backup strategy in a separate command.")
			fmt.Fprintln(out, "Inspection only. No OpenCode files were modified.")
			return nil
		},
	}
}

type openCodeStatusDetails struct {
	agents.AgentDetectionResult
	ConfigPathsFound      []string
	TargetConfigCandidate string
	CanExportSafely       bool
	Notes                 []string
	Warnings              []string
}

func buildOpenCodeStatus(ctx context.Context, rt *Runtime) (openCodeStatusDetails, error) {
	res, err := agents.OpenCodeDetector{}.Detect(ctx)
	if err != nil {
		return openCodeStatusDetails{}, fmt.Errorf("detect opencode: %w", err)
	}

	wd, wdErr := os.Getwd()
	if wdErr != nil {
		wd = ""
	}
	candidates := agents.OpenCodeConfigPathCandidates(wd)
	found := agents.ExistingOpenCodeConfigPaths(wd)

	details := openCodeStatusDetails{
		AgentDetectionResult: res,
		ConfigPathsFound:     found,
	}

	if len(found) > 0 {
		details.TargetConfigCandidate = found[0]
	} else if len(candidates) > 0 {
		details.TargetConfigCandidate = candidates[0]
		details.Warnings = append(details.Warnings, "OpenCode config path is not present yet; preview uses a fallback candidate path.")
	}

	if res.Found {
		details.CanExportSafely = true
		details.Notes = append(details.Notes, "OpenCode binary is present; preview/export planning can run safely.")
	} else {
		details.CanExportSafely = false
		details.Warnings = append(details.Warnings, "OpenCode binary is missing; export apply must remain disabled.")
		details.Notes = append(details.Notes, "Status and preview are metadata-only and do not require OpenCode writes.")
	}

	if strings.TrimSpace(wd) != "" {
		projectLocal := filepath.Join(wd, ".opencode")
		if _, err := os.Stat(projectLocal); err == nil {
			details.Notes = append(details.Notes, "Project-local .opencode directory detected.")
		}
	}

	if len(found) == 0 {
		details.Notes = append(details.Notes, "No existing OpenCode config directories were found among common candidates.")
	}

	return details, nil
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == target {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}
