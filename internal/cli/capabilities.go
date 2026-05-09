package cli

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func newCapabilitiesCmd(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "Inspect capability registry and installation plans",
	}

	cmd.AddCommand(
		newCapabilitiesStatusCmd(rt),
		newCapabilitiesInspectCmd(rt),
		newCapabilitiesReportCmd(),
		newCapabilitiesConfigureCmd(rt),
		newCapabilitiesValidateCmd(rt),
		newCapabilitiesPlanCmd(rt),
		newCapabilitiesExportPreviewCmd(rt),
		newCapabilitiesExportCmd(rt),
		newCapabilitiesApplyCmd(rt),
	)

	return cmd
}

func newCapabilitiesExportPreviewCmd(rt *Runtime) *cobra.Command {
	var targetAgent string
	var selected []string
	cmd := &cobra.Command{
		Use:   "export-preview",
		Short: "Preview selection-aware capabilities export to target managed area",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCapabilitiesExport(rt, cmd, true, targetAgent, selected)
		},
	}
	cmd.Flags().StringVar(&targetAgent, "target-agent", "opencode", "Target agent adapter (phase support: opencode)")
	cmd.Flags().StringArrayVar(&selected, "select", []string{}, "Explicit selection IDs (flow/pack aliases), repeatable")
	return cmd
}

func newCapabilitiesExportCmd(rt *Runtime) *cobra.Command {
	var targetAgent string
	var selected []string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export selection-aware capabilities assets to target managed area",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCapabilitiesExport(rt, cmd, rt.DryRun, targetAgent, selected)
		},
	}
	cmd.Flags().StringVar(&targetAgent, "target-agent", "opencode", "Target agent adapter (phase support: opencode)")
	cmd.Flags().StringArrayVar(&selected, "select", []string{}, "Explicit selection IDs (flow/pack aliases), repeatable")
	return cmd
}

func runCapabilitiesExport(rt *Runtime, cmd *cobra.Command, preview bool, targetAgent string, selected []string) error {
	if strings.TrimSpace(targetAgent) == "" {
		targetAgent = "opencode"
	}
	if targetAgent != "opencode" {
		return fmt.Errorf("unsupported target agent %q in this phase; only opencode is supported", targetAgent)
	}
	cfg, err := rt.App.ConfigManager.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	resolved := uniqueSorted(selected)
	if len(resolved) == 0 {
		resolved = uniqueSorted(append(append([]string{}, cfg.Components.Flows.Selected...), cfg.Components.Skills.Selected...))
	}
	catalog, err := capabilities.DiscoverAssets(".")
	if err != nil {
		return fmt.Errorf("discover assets: %w", err)
	}
	projection, err := capabilities.BuildOpenCodeExportProjection(catalog, resolved)
	if err != nil {
		return fmt.Errorf("build export projection: %w", err)
	}
	details, err := buildOpenCodeStatus(context.Background(), rt)
	if err != nil {
		return err
	}
	configRoot := strings.TrimSpace(details.TargetConfigCandidate)
	if configRoot == "" {
		return fmt.Errorf("no OpenCode config root candidate resolved")
	}
	result, err := capabilities.ExecuteOpenCodeSkillsExport(projection, configRoot, preview)
	if err != nil {
		_, _ = capabilities.WriteOpenCodeSkillsExportReport(capabilities.SkillsExportResult{DryRun: preview, Status: "failed", TargetRoot: filepath.Join(configRoot, "bitsentry"), SelectedIDs: resolved, Warnings: projection.Warnings, Skipped: projection.Skipped})
		return err
	}
	reportPath, reportErr := capabilities.WriteOpenCodeSkillsExportReport(result)
	out := cmd.OutOrStdout()
	mode := "export"
	if preview {
		mode = "preview"
	}
	fmt.Fprintf(out, "Capabilities %s\n", mode)
	fmt.Fprintf(out, "- target-agent: %s\n", targetAgent)
	fmt.Fprintf(out, "- selected IDs: %s\n", fallback(strings.Join(result.SelectedIDs, ", "), "none"))
	fmt.Fprintf(out, "- managed target root: %s\n", result.TargetRoot)
	fmt.Fprintf(out, "- included flows: %s\n", fallback(strings.Join(result.IncludedFlows, ", "), "none"))
	fmt.Fprintf(out, "- included skill packs: %s\n", fallback(strings.Join(result.IncludedPacks, ", "), "none"))
	fmt.Fprintf(out, "- included skills count: %d\n", result.IncludedSkills)
	fmt.Fprintf(out, "- generated files: %s\n", fallback(strings.Join(result.GeneratedFiles, ", "), "none"))
	fmt.Fprintf(out, "- written files count: %d\n", len(result.WrittenFiles))
	if result.BackupPath != "" {
		fmt.Fprintf(out, "- backup path: %s\n", result.BackupPath)
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(out, "- warnings:")
		for _, w := range result.Warnings {
			fmt.Fprintf(out, "  - %s\n", w)
		}
	}
	if len(result.Skipped) > 0 {
		fmt.Fprintln(out, "- skipped:")
		for _, s := range result.Skipped {
			fmt.Fprintf(out, "  - %s\n", s)
		}
	}
	if reportErr == nil {
		fmt.Fprintf(out, "- report path: %s\n", reportPath)
	} else {
		fmt.Fprintf(out, "- report warning: %v\n", reportErr)
	}
	if preview {
		fmt.Fprintln(out, "No files were modified.")
	}
	return nil
}

func newCapabilitiesValidateCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate saved capability configuration against current registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			issues := capabilities.ValidateSavedConfig(catalog, cfg)
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Capability config validation")
			fmt.Fprintf(out, "- preset: %s\n", fallback(cfg.Components.Preset, "custom"))
			fmt.Fprintf(out, "- targets: %s\n", fallback(strings.Join(cfg.Components.Targets.Selected, ", "), "none"))
			fmt.Fprintf(out, "- selected MCPs: %s\n", fallback(strings.Join(cfg.Components.MCPs.Selected, ", "), "none"))
			fmt.Fprintf(out, "- selected skills: %s\n", fallback(strings.Join(cfg.Components.Skills.Selected, ", "), "none"))
			fmt.Fprintf(out, "- selected flows: %s\n", fallback(strings.Join(cfg.Components.Flows.Selected, ", "), "none"))

			if len(issues) == 0 {
				fmt.Fprintln(out, "- result: VALID")
				return nil
			}

			fmt.Fprintln(out, "- result: INVALID")
			fmt.Fprintln(out, "- issues:")
			for _, issue := range issues {
				fmt.Fprintf(out, "  - %s\n", issue)
			}
			return fmt.Errorf("capability config validation failed with %d issue(s)", len(issues))
		},
	}
}

func newCapabilitiesStatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show modeled capabilities (MCPs, skills, flows, targets, presets)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Capability registry")
			fmt.Fprintln(out, "- target support in this phase: opencode only")

			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "KIND\tID\tSTATUS\tSELECTED\tNOTES")
			for _, e := range catalog.MCPs {
				fmt.Fprintf(tw, "mcp\t%s\t%s\t%s\t%s\n", e.ID, e.Status, yesNo(e.Selected), trimNote(e.Notes))
			}
			for _, e := range catalog.Skills {
				fmt.Fprintf(tw, "skill\t%s\t%s\t%s\t%s\n", e.ID, e.Status, yesNo(e.Selected), trimNote(e.Notes))
			}
			for _, e := range catalog.Flows {
				fmt.Fprintf(tw, "flow\t%s\t%s\t%s\t%s\n", e.ID, e.Status, yesNo(e.Selected), trimNote(e.Notes))
			}
			for _, e := range catalog.Targets {
				fmt.Fprintf(tw, "target\t%s\t%s\t%s\t%s\n", e.ID, e.Status, yesNo(e.Selected), trimNote(e.Notes))
			}
			_ = tw.Flush()

			fmt.Fprintln(out, "\nPresets")
			for _, p := range catalog.Presets {
				fmt.Fprintf(out, "- %s\n", p.ID)
				fmt.Fprintf(out, "  mcps: %s\n", fallback(strings.Join(p.MCPs, ", "), "none"))
				fmt.Fprintf(out, "  skills: %s\n", fallback(strings.Join(p.Skills, ", "), "none"))
				fmt.Fprintf(out, "  flows: %s\n", fallback(strings.Join(p.Flows, ", "), "none"))
				fmt.Fprintf(out, "  targets: %s\n", fallback(strings.Join(p.Targets, ", "), "none"))
			}

			return nil
		},
	}
}

func newCapabilitiesPlanCmd(rt *Runtime) *cobra.Command {
	var targetAgent string
	var presetID string

	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Build capability installation plan and OpenCode projection",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			resolvedTarget := targetAgent
			if !cmd.Flags().Changed("target-agent") {
				if len(cfg.Components.Targets.Selected) > 0 {
					resolvedTarget = cfg.Components.Targets.Selected[0]
				}
			}

			resolvedPreset := presetID
			if !cmd.Flags().Changed("preset") {
				resolvedPreset = strings.TrimSpace(cfg.Components.Preset)
			}

			var plan capabilities.Plan
			if strings.TrimSpace(resolvedPreset) == "" || strings.TrimSpace(resolvedPreset) == "custom" {
				plan, err = capabilities.BuildPlanFromSelection(
					catalog,
					resolvedTarget,
					firstNonEmpty(resolvedPreset, "custom"),
					cfg.Components.MCPs.Selected,
					cfg.Components.Skills.Selected,
					cfg.Components.Flows.Selected,
				)
			} else {
				plan, err = capabilities.BuildPlan(catalog, resolvedTarget, resolvedPreset)
			}
			if err != nil {
				return err
			}
			projection := capabilities.ProjectOpenCode(plan)

			payload := map[string]any{
				"source": map[string]any{
					"target_agent": resolvedTarget,
					"preset":       firstNonEmpty(resolvedPreset, "custom"),
					"config_based": !cmd.Flags().Changed("preset") && !cmd.Flags().Changed("target-agent"),
				},
				"plan":                plan,
				"opencode_projection": projection,
			}
			raw, err := yaml.Marshal(payload)
			if err != nil {
				return fmt.Errorf("marshal plan: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(raw))
			return nil
		},
	}

	cmd.Flags().StringVar(&targetAgent, "target-agent", "opencode", "Target agent adapter (phase support: opencode)")
	cmd.Flags().StringVar(&presetID, "preset", "", "Capability preset (bitsentry-full/dev/research/oscp/bugbounty/redteam/blog/custom). If omitted, use saved config selection.")
	return cmd
}

func newCapabilitiesConfigureCmd(rt *Runtime) *cobra.Command {
	var targetAgent string
	var presetID string
	var selectedMCPs []string
	var selectedSkills []string
	var selectedFlows []string
	var clearMCPs bool
	var clearSkills bool
	var clearFlows bool
	var clearTargets bool
	var resetAll bool
	var clearAll bool

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Persist capability selection into bitsentry-ai config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			resolvedTarget := strings.TrimSpace(targetAgent)
			if resolvedTarget == "" {
				resolvedTarget = "opencode"
			}

			resolvedPreset := strings.TrimSpace(presetID)
			if resolvedPreset == "" {
				resolvedPreset = "custom"
			}

			if clearAll {
				resetAll = true
			}

			if err := capabilities.ValidateSelections(catalog, resolvedTarget, selectedMCPs, selectedSkills, selectedFlows); err != nil {
				return err
			}

			nextCfg := cfg

			if resetAll {
				nextCfg.Components.Preset = "custom"
				nextCfg.Components.MCPs.Enabled = false
				nextCfg.Components.MCPs.Configured = false
				nextCfg.Components.MCPs.Selected = []string{}
				nextCfg.Components.Skills.Enabled = false
				nextCfg.Components.Skills.Configured = false
				nextCfg.Components.Skills.Selected = []string{}
				nextCfg.Components.Flows.Enabled = false
				nextCfg.Components.Flows.Configured = false
				nextCfg.Components.Flows.Selected = []string{}
				nextCfg.Components.Targets.Selected = []string{"opencode"}
			} else {
				if clearTargets {
					nextCfg.Components.Targets.Selected = []string{}
				} else {
					nextCfg.Components.Targets.Selected = []string{resolvedTarget}
				}

				if clearMCPs {
					nextCfg.Components.MCPs.Enabled = false
					nextCfg.Components.MCPs.Configured = false
					nextCfg.Components.MCPs.Selected = []string{}
				}
				if clearSkills {
					nextCfg.Components.Skills.Enabled = false
					nextCfg.Components.Skills.Configured = false
					nextCfg.Components.Skills.Selected = []string{}
				}
				if clearFlows {
					nextCfg.Components.Flows.Enabled = false
					nextCfg.Components.Flows.Configured = false
					nextCfg.Components.Flows.Selected = []string{}
				}
			}

			if !resetAll && len(selectedMCPs) == 0 && len(selectedSkills) == 0 && len(selectedFlows) == 0 && resolvedPreset != "custom" {
				preset, ok := capabilities.PresetByID(resolvedPreset, catalog.Presets)
				if !ok {
					return fmt.Errorf("unknown preset %q", resolvedPreset)
				}
				nextCfg.Components.Preset = preset.ID
				nextCfg.Components.MCPs.Enabled = true
				nextCfg.Components.MCPs.Configured = true
				nextCfg.Components.MCPs.Selected = uniqueSorted(preset.MCPs)
				nextCfg.Components.Skills.Enabled = true
				nextCfg.Components.Skills.Configured = true
				nextCfg.Components.Skills.Selected = uniqueSorted(preset.Skills)
				nextCfg.Components.Flows.Enabled = true
				nextCfg.Components.Flows.Configured = true
				nextCfg.Components.Flows.Selected = uniqueSorted(preset.Flows)
			} else if !resetAll {
				nextCfg.Components.Preset = "custom"
				if len(selectedMCPs) > 0 {
					nextCfg.Components.MCPs.Enabled = true
					nextCfg.Components.MCPs.Configured = true
					nextCfg.Components.MCPs.Selected = uniqueSorted(selectedMCPs)
				}
				if len(selectedSkills) > 0 {
					nextCfg.Components.Skills.Enabled = true
					nextCfg.Components.Skills.Configured = true
					nextCfg.Components.Skills.Selected = uniqueSorted(selectedSkills)
				}
				if len(selectedFlows) > 0 {
					nextCfg.Components.Flows.Enabled = true
					nextCfg.Components.Flows.Configured = true
					nextCfg.Components.Flows.Selected = uniqueSorted(selectedFlows)
				}
			}

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] Capability configure preview")
				printCapabilityConfig(out, nextCfg)
				return nil
			}

			if err := rt.App.ConfigManager.Save(nextCfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(out, "Capabilities configured successfully.")
			printCapabilityConfig(out, nextCfg)
			return nil
		},
	}

	cmd.Flags().StringVar(&targetAgent, "target-agent", "opencode", "Target agent adapter (phase support: opencode)")
	cmd.Flags().StringVar(&presetID, "preset", "custom", "Preset to apply (bitsentry-full/dev/research/oscp/bugbounty/redteam/blog/custom)")
	cmd.Flags().StringArrayVar(&selectedMCPs, "mcp", []string{}, "Select MCP capability (repeatable)")
	cmd.Flags().StringArrayVar(&selectedSkills, "skill", []string{}, "Select skill capability (repeatable)")
	cmd.Flags().StringArrayVar(&selectedFlows, "flow", []string{}, "Select flow capability (repeatable)")
	cmd.Flags().BoolVar(&clearMCPs, "clear-mcps", false, "Clear selected MCP capabilities")
	cmd.Flags().BoolVar(&clearSkills, "clear-skills", false, "Clear selected skill capabilities")
	cmd.Flags().BoolVar(&clearFlows, "clear-flows", false, "Clear selected flow capabilities")
	cmd.Flags().BoolVar(&clearTargets, "clear-targets", false, "Clear selected target agents")
	cmd.Flags().BoolVar(&resetAll, "reset-all", false, "Reset capability selection to clean defaults")
	cmd.Flags().BoolVar(&clearAll, "clear-all", false, "Alias for --reset-all")
	return cmd
}

func newCapabilitiesApplyCmd(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply saved capability plan through target adapter (OpenCode first)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			target := "opencode"
			if len(cfg.Components.Targets.Selected) > 0 {
				target = cfg.Components.Targets.Selected[0]
			}

			plan, err := capabilities.BuildPlanFromSelection(
				catalog,
				target,
				firstNonEmpty(cfg.Components.Preset, "custom"),
				cfg.Components.MCPs.Selected,
				cfg.Components.Skills.Selected,
				cfg.Components.Flows.Selected,
			)
			if err != nil {
				return err
			}
			projection := capabilities.ProjectOpenCode(plan)

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, capabilities.ApplySummary(plan, projection))
				if len(projection.Warnings) > 0 {
					fmt.Fprintln(out, "- warnings:")
					for _, w := range projection.Warnings {
						fmt.Fprintf(out, "  - %s\n", w)
					}
				}
				reportPath, reportErr := capabilities.WriteApplyReport(cfg, plan, projection, "dry-run", "preview", "No files were modified.")
				if reportErr == nil {
					fmt.Fprintf(out, "- report saved: %s\n", reportPath)
				}
				fmt.Fprintln(out, "No files were modified.")
				return nil
			}

			if len(projection.ManagedMCPs) == 0 {
				fmt.Fprintln(out, "No managed MCP capabilities selected for apply.")
				fmt.Fprintln(out, "Nothing to do. Skills and flows remain declarative in this phase.")
				reportPath, reportErr := capabilities.WriteApplyReport(cfg, plan, projection, "apply", "noop", "No managed MCP capabilities selected.")
				if reportErr == nil {
					fmt.Fprintf(out, "- report saved: %s\n", reportPath)
				}
				return nil
			}

			applyCmd := newAgentsOpenCodeApplyCmd(rt)
			applyCmd.SetOut(out)
			applyCmd.SetErr(cmd.ErrOrStderr())
			if applyCmd.RunE == nil {
				return fmt.Errorf("opencode apply command is not executable")
			}
			if err := applyCmd.RunE(applyCmd, []string{}); err != nil {
				_, _ = capabilities.WriteApplyReport(cfg, plan, projection, "apply", "failed", err.Error())
				return err
			}
			reportPath, reportErr := capabilities.WriteApplyReport(cfg, plan, projection, "apply", "applied", "Applied managed OpenCode MCP projection.")
			if reportErr == nil {
				fmt.Fprintf(out, "- capability apply report: %s\n", reportPath)
			}
			return nil
		},
	}
	return cmd
}

func printApplySummary(out io.Writer, plan capabilities.Plan, projection capabilities.OpenCodeProjection) {
	fmt.Fprintln(out, capabilities.ApplySummary(plan, projection))
}

func newCapabilitiesInspectCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "inspect <id>",
		Short: "Inspect a capability entry by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			skillEntries, _ := components.SkillsRegistry(cfg)
			catalog := capabilities.BuildCatalog(cfg, mcpEntries, skillEntries)

			id := strings.TrimSpace(args[0])
			for _, e := range append(append(append(catalog.MCPs, catalog.Skills...), catalog.Flows...), catalog.Targets...) {
				if e.ID != id {
					continue
				}
				out := cmd.OutOrStdout()
				fmt.Fprintln(out, "Capability inspect")
				fmt.Fprintf(out, "- id: %s\n", e.ID)
				fmt.Fprintf(out, "- name: %s\n", e.Name)
				fmt.Fprintf(out, "- kind: %s\n", e.Kind)
				fmt.Fprintf(out, "- category: %s\n", e.Category)
				fmt.Fprintf(out, "- status: %s\n", e.Status)
				fmt.Fprintf(out, "- selected: %s\n", yesNo(e.Selected))
				fmt.Fprintf(out, "- enabled: %s\n", yesNo(e.Enabled))
				fmt.Fprintf(out, "- configured: %s\n", yesNo(e.Configured))
				fmt.Fprintf(out, "- targets: %s\n", fallback(strings.Join(e.Targets, ", "), "none"))
				fmt.Fprintf(out, "- description: %s\n", e.Description)
				fmt.Fprintf(out, "- notes: %s\n", fallback(e.Notes, "none"))
				return nil
			}
			return fmt.Errorf("capability %q not found", id)
		},
	}
}

func newCapabilitiesReportCmd() *cobra.Command {
	reportCmd := &cobra.Command{Use: "report", Short: "Inspect persisted capability apply reports"}
	reportCmd.AddCommand(&cobra.Command{
		Use:   "latest",
		Short: "Print the latest capability apply report",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := capabilities.ReadLatestApplyReport()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), raw)
			return nil
		},
	})
	return reportCmd
}

func printCapabilityConfig(out io.Writer, cfg config.Config) {
	fmt.Fprintf(out, "- components.preset: %s\n", fallback(cfg.Components.Preset, "custom"))
	fmt.Fprintf(out, "- components.targets.selected: %s\n", fallback(strings.Join(cfg.Components.Targets.Selected, ", "), "none"))
	fmt.Fprintf(out, "- components.mcps.selected: %s\n", fallback(strings.Join(cfg.Components.MCPs.Selected, ", "), "none"))
	fmt.Fprintf(out, "- components.skills.selected: %s\n", fallback(strings.Join(cfg.Components.Skills.Selected, ", "), "none"))
	fmt.Fprintf(out, "- components.flows.selected: %s\n", fallback(strings.Join(cfg.Components.Flows.Selected, ", "), "none"))
}

func uniqueSorted(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" || set[t] {
			continue
		}
		set[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func trimNote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) <= 72 {
		return v
	}
	return v[:69] + "..."
}
