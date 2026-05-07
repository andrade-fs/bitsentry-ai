package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"bitsentry-ai/internal/components"
	"github.com/spf13/cobra"
)

func newComponentsCmd(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components",
		Short: "List known components",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			cfg, cfgErr := rt.App.ConfigManager.Load()
			engramDetails := components.EngramRuntimeDetails{
				Status: components.StatusError,
				Notes:  []string{"Engram runtime detection could not run because config failed to load."},
			}
			if cfgErr == nil {
				engramDetails = components.DetectEngramRuntime(context.Background(), cfg)
			} else {
				engramDetails.Notes = append(engramDetails.Notes, fmt.Sprintf("config error: %v", cfgErr))
			}
			context7Details := components.Context7RuntimeDetails{
				Status: components.StatusError,
				Notes:  []string{"Context7 runtime detection could not run because config failed to load."},
			}
			if cfgErr == nil {
				context7Details = components.DetectContext7Runtime(context.Background(), cfg)
			} else {
				context7Details.Notes = append(context7Details.Notes, fmt.Sprintf("config error: %v", cfgErr))
			}

			entries := components.Registry()
			skillsSummary := components.SkillsRegistrySummary{}
			if cfgErr == nil {
				_, skillsSummary = components.SkillsRegistry(cfg)
			}
			for i := range entries {
				if entries[i].ID == "engram" {
					entries[i].Status = engramDetails.Status
					entries[i].Notes = strings.Join(engramDetails.Notes, " ")
				}
				if entries[i].ID == "context7" {
					entries[i].Status = context7Details.Status
					entries[i].Notes = strings.Join(context7Details.Notes, " ")
				}
				if entries[i].ID == "mcps" {
					if cfg.Components.MCPs.Configured {
						entries[i].Status = components.StatusConfigured
					} else if cfg.Components.MCPs.Enabled {
						entries[i].Status = components.StatusDetected
					}
					entries[i].Notes = fmt.Sprintf("MCP metadata configured=%s selected=%s", yesNo(cfg.Components.MCPs.Configured), fallback(strings.Join(cfg.Components.MCPs.Selected, ","), "none"))
				}
				if entries[i].ID == "skills" {
					if skillsSummary.Configured {
						entries[i].Status = components.StatusConfigured
					} else if skillsSummary.Enabled {
						entries[i].Status = components.StatusDetected
					}
					entries[i].Notes = fmt.Sprintf("Skills metadata configured=%s selected=%s", yesNo(skillsSummary.Configured), fallback(strings.Join(skillsSummary.Selected, ","), "none"))
				}
			}

			fmt.Fprintln(out, "Components")
			fmt.Fprintln(out, "Note: Engram status is runtime-detected. Other components still show modeled metadata.")
			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tCATEGORY\tSTATUS\tREQUIRED\tINSTALL\tCONFIGURE")
			for _, c := range entries {
				required := "no"
				if c.Required {
					required = "yes"
				}
				install := "no"
				if c.InstallSupported {
					install = "yes"
				}
				configure := "no"
				if c.ConfigureSupported {
					configure = "yes"
				}

				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", c.ID, c.Name, c.Category, c.StatusLabel(), required, install, configure)
				fmt.Fprintf(tw, "\t↳ %s\t\t\t\t\t\n", c.Description)
				if c.Notes != "" {
					fmt.Fprintf(tw, "\t↳ note: %s\t\t\t\t\t\n", c.Notes)
				}
				if c.DocsURL != "" {
					fmt.Fprintf(tw, "\t↳ docs: %s\t\t\t\t\t\n", c.DocsURL)
				}
			}
			_ = tw.Flush()

			fmt.Fprintln(out, "\nEngram runtime details")
			fmt.Fprintf(out, "- binary found: %s\n", yesNo(engramDetails.BinaryFound))
			fmt.Fprintf(out, "- binary path: %s\n", fallback(engramDetails.BinaryPath, "n/a"))
			fmt.Fprintf(out, "- version: %s\n", fallback(engramDetails.Version, "unavailable"))
			fmt.Fprintf(out, "- data dir found: %s\n", yesNo(engramDetails.DataDirFound))
			fmt.Fprintf(out, "- data dir path: %s\n", fallback(engramDetails.DataDirPath, "n/a"))
			fmt.Fprintf(out, "- bitsentry config enabled: %s\n", yesNo(engramDetails.ConfigEnabled))
			fmt.Fprintf(out, "- bitsentry config configured: %s\n", yesNo(engramDetails.ConfigConfigured))
			if len(engramDetails.Notes) > 0 {
				fmt.Fprintln(out, "- notes:")
				for _, n := range engramDetails.Notes {
					fmt.Fprintf(out, "  - %s\n", n)
				}
			}

			fmt.Fprintln(out, "\nContext7 runtime details")
			fmt.Fprintf(out, "- command detected: %s\n", yesNo(context7Details.Detected))
			fmt.Fprintf(out, "- detected command: %s\n", fallback(context7Details.DetectedCommand, "not found"))
			fmt.Fprintf(out, "- detected path: %s\n", fallback(context7Details.DetectedPath, "not found"))
			fmt.Fprintf(out, "- bitsentry config enabled: %s\n", yesNo(context7Details.ConfigEnabled))
			fmt.Fprintf(out, "- bitsentry config configured: %s\n", yesNo(context7Details.ConfigConfigured))
			fmt.Fprintf(out, "- bitsentry config command: %s\n", fallback(context7Details.ConfigCommand, "n/a"))
			fmt.Fprintf(out, "- bitsentry config package: %s\n", fallback(context7Details.ConfigPackage, "n/a"))
			if len(context7Details.Notes) > 0 {
				fmt.Fprintln(out, "- notes:")
				for _, n := range context7Details.Notes {
					fmt.Fprintf(out, "  - %s\n", n)
				}
			}

			fmt.Fprintf(out, "\nSummary: %d components, %d required, %d optional. Engram/Context7 are runtime-evaluated; MCPs/Skills are config-backed metadata only.\n", len(entries), countRequired(entries), countOptional(entries))
		},
	}

	cmd.AddCommand(newComponentsEngramCmd(rt))
	cmd.AddCommand(newComponentsContext7Cmd(rt))
	cmd.AddCommand(newComponentsMCPsCmd(rt))
	cmd.AddCommand(newComponentsSkillsCmd(rt))
	return cmd
}

func newComponentsSkillsCmd(rt *Runtime) *cobra.Command {
	skillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Inspect and configure Skills metadata registry",
	}

	statusCmd := newComponentsSkillsStatusCmd(rt)
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Alias for skills status",
		Aliases: []string{"ls"},
		RunE:    statusCmd.RunE,
	}

	skillsCmd.AddCommand(statusCmd, listCmd, newComponentsSkillsConfigureCmd(rt))
	return skillsCmd
}

func newComponentsSkillsStatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show Skills metadata registry status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			entries, summary := components.SkillsRegistry(cfg)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Skills registry status")
			fmt.Fprintf(out, "- enabled in bitsentry config: %s\n", yesNo(summary.Enabled))
			fmt.Fprintf(out, "- configured in bitsentry config: %s\n", yesNo(summary.Configured))
			fmt.Fprintf(out, "- selected skills: %s\n", fallback(strings.Join(summary.Selected, ", "), "none"))
			fmt.Fprintf(out, "- total modeled skills: %d\n", summary.Total)

			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tCATEGORY\tSTATUS\tTARGET_AGENTS\tENABLED\tCONFIGURED")
			for _, skill := range entries {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", skill.ID, skill.Name, skill.Category, skill.Status, strings.Join(skill.TargetAgents, ","), yesNo(skill.Enabled), yesNo(skill.Configured))
				fmt.Fprintf(tw, "\t↳ %s\t\t\t\t\t\t\n", skill.Description)
				if skill.Notes != "" {
					fmt.Fprintf(tw, "\t↳ note: %s\t\t\t\t\t\t\n", skill.Notes)
				}
			}
			_ = tw.Flush()
			return nil
		},
	}
}

func newComponentsSkillsConfigureCmd(rt *Runtime) *cobra.Command {
	defaultSelected := []string{
		"bitsentry-research-init",
		"bitsentry-research-create",
		"bitsentry-research-validate",
		"bitsentry-sdd",
		"bitsentry-oscp-notes",
		"bitsentry-bugbounty-notes",
	}

	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Skills metadata defaults in bitsentry-ai",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			nextCfg := cfg
			nextCfg.Components.Skills.Enabled = true
			nextCfg.Components.Skills.Configured = true
			nextCfg.Components.Skills.Selected = append([]string{}, defaultSelected...)

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] Skills configure preview")
				fmt.Fprintln(out, "This only configures bitsentry-ai Skills metadata. It does not install skills or modify agent configs.")
				fmt.Fprintln(out, "The following fields would be written:")
				fmt.Fprintf(out, "- components.skills.enabled: %t\n", nextCfg.Components.Skills.Enabled)
				fmt.Fprintf(out, "- components.skills.configured: %t\n", nextCfg.Components.Skills.Configured)
				fmt.Fprintf(out, "- components.skills.selected: [%s]\n", strings.Join(nextCfg.Components.Skills.Selected, ", "))
				return nil
			}

			if err := rt.App.ConfigManager.Save(nextCfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(out, "Skills metadata configured successfully.")
			fmt.Fprintln(out, "This only configures bitsentry-ai Skills metadata. It does not install skills or modify agent configs.")
			fmt.Fprintf(out, "- selected skills: %s\n", strings.Join(nextCfg.Components.Skills.Selected, ", "))
			return nil
		},
	}
}

func newComponentsMCPsCmd(rt *Runtime) *cobra.Command {
	mcpsCmd := &cobra.Command{
		Use:   "mcps",
		Short: "Inspect and configure MCP metadata registry",
	}

	statusCmd := newComponentsMCPsStatusCmd(rt)
	listCmd := &cobra.Command{
		Use:     "list",
		Short:   "Alias for mcps status",
		Aliases: []string{"ls"},
		RunE:    statusCmd.RunE,
	}

	mcpsCmd.AddCommand(statusCmd, listCmd, newComponentsMCPsConfigureCmd(rt))
	return mcpsCmd
}

func newComponentsMCPsStatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show MCP registry status (metadata only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			entries, summary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "MCP registry status")
			fmt.Fprintf(out, "- enabled in bitsentry config: %s\n", yesNo(summary.Enabled))
			fmt.Fprintf(out, "- configured in bitsentry config: %s\n", yesNo(summary.Configured))
			fmt.Fprintf(out, "- selected MCPs: %s\n", fallback(strings.Join(summary.Selected, ", "), "none"))
			fmt.Fprintf(out, "- total modeled MCPs: %d\n", summary.Total)

			tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tNAME\tCATEGORY\tSTATUS\tENABLED\tCONFIGURED")
			for _, mcp := range entries {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", mcp.ID, mcp.Name, mcp.Category, mcp.Status, yesNo(mcp.Enabled), yesNo(mcp.Configured))
				if mcp.Command != "" {
					fmt.Fprintf(tw, "\t↳ command: %s %s\t\t\t\t\t\n", mcp.Command, strings.Join(mcp.Args, " "))
				}
				if mcp.Package != "" {
					fmt.Fprintf(tw, "\t↳ package: %s\t\t\t\t\t\n", mcp.Package)
				}
				if mcp.Notes != "" {
					fmt.Fprintf(tw, "\t↳ note: %s\t\t\t\t\t\n", mcp.Notes)
				}
			}
			_ = tw.Flush()
			return nil
		},
	}
}

func newComponentsMCPsConfigureCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure MCP metadata defaults in bitsentry-ai",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			nextCfg := cfg
			nextCfg.Components.MCPs.Enabled = true
			nextCfg.Components.MCPs.Configured = true
			nextCfg.Components.MCPs.Selected = []string{"engram", "context7"}

			out := cmd.OutOrStdout()
			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] MCP configure preview")
				fmt.Fprintln(out, "This only configures bitsentry-ai MCP metadata. It does not install MCP servers or modify agent configs.")
				fmt.Fprintln(out, "The following fields would be written:")
				fmt.Fprintf(out, "- components.mcps.enabled: %t\n", nextCfg.Components.MCPs.Enabled)
				fmt.Fprintf(out, "- components.mcps.configured: %t\n", nextCfg.Components.MCPs.Configured)
				fmt.Fprintf(out, "- components.mcps.selected: [%s]\n", strings.Join(nextCfg.Components.MCPs.Selected, ", "))
				return nil
			}

			if err := rt.App.ConfigManager.Save(nextCfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(out, "MCP metadata configured successfully.")
			fmt.Fprintln(out, "This only configures bitsentry-ai MCP metadata. It does not install MCP servers or modify agent configs.")
			fmt.Fprintf(out, "- selected MCPs: %s\n", strings.Join(nextCfg.Components.MCPs.Selected, ", "))
			return nil
		},
	}
}

func newComponentsContext7Cmd(rt *Runtime) *cobra.Command {
	context7Cmd := &cobra.Command{
		Use:   "context7",
		Short: "Inspect and configure Context7 component",
	}

	context7Cmd.AddCommand(
		newComponentsContext7StatusCmd(rt),
		newComponentsContext7ConfigureCmd(rt),
	)

	return context7Cmd
}

func newComponentsContext7StatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show detailed Context7 runtime status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			details := components.DetectContext7Runtime(context.Background(), cfg)
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, "Context7 status")
			fmt.Fprintf(out, "- command detected: %s\n", yesNo(details.Detected))
			fmt.Fprintf(out, "- detected command: %s\n", fallback(details.DetectedCommand, "not found"))
			fmt.Fprintf(out, "- detected path: %s\n", fallback(details.DetectedPath, "not found"))
			fmt.Fprintf(out, "- config enabled: %s\n", yesNo(details.ConfigEnabled))
			fmt.Fprintf(out, "- config configured: %s\n", yesNo(details.ConfigConfigured))
			fmt.Fprintf(out, "- config command: %s\n", fallback(details.ConfigCommand, "n/a"))
			fmt.Fprintf(out, "- config package: %s\n", fallback(details.ConfigPackage, "n/a"))
			fmt.Fprintf(out, "- final status: %s\n", string(details.Status))

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

func newComponentsContext7ConfigureCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Context7 component in bitsentry-ai",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			details := components.DetectContext7Runtime(context.Background(), cfg)
			out := cmd.OutOrStdout()

			command := "npx"
			if details.DetectedCommand != "" {
				command = details.DetectedCommand
			}

			nextCfg := cfg
			nextCfg.Components.Context7.Enabled = true
			nextCfg.Components.Context7.Configured = true
			nextCfg.Components.Context7.Command = command
			nextCfg.Components.Context7.Package = "@upstash/context7-mcp"
			nextCfg.Components.Context7.Notes = "bitsentry-ai metadata only; does not install or validate Context7 MCP server"

			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] Context7 configure preview")
				fmt.Fprintln(out, "The following fields would be written:")
				fmt.Fprintf(out, "- components.context7.enabled: %t\n", nextCfg.Components.Context7.Enabled)
				fmt.Fprintf(out, "- components.context7.configured: %t\n", nextCfg.Components.Context7.Configured)
				fmt.Fprintf(out, "- components.context7.command: %s\n", nextCfg.Components.Context7.Command)
				fmt.Fprintf(out, "- components.context7.package: %s\n", nextCfg.Components.Context7.Package)
				fmt.Fprintf(out, "- components.context7.notes: %s\n", nextCfg.Components.Context7.Notes)
				fmt.Fprintln(out, "Note: this updates bitsentry-ai metadata only. No installation or MCP runtime validation is performed.")
				return nil
			}

			if err := rt.App.ConfigManager.Save(nextCfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(out, "Context7 configured successfully.")
			fmt.Fprintf(out, "- command: %s\n", nextCfg.Components.Context7.Command)
			fmt.Fprintf(out, "- package: %s\n", nextCfg.Components.Context7.Package)
			fmt.Fprintln(out, "Note: this writes bitsentry-ai metadata only. It does NOT install or validate the Context7 MCP server.")
			return nil
		},
	}
}

func newComponentsEngramCmd(rt *Runtime) *cobra.Command {
	engramCmd := &cobra.Command{
		Use:   "engram",
		Short: "Inspect and configure Engram component",
	}

	engramCmd.AddCommand(
		newComponentsEngramStatusCmd(rt),
		newComponentsEngramConfigureCmd(rt),
	)

	return engramCmd
}

func newComponentsEngramStatusCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show detailed Engram runtime status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			details := components.DetectEngramRuntime(context.Background(), cfg)
			out := cmd.OutOrStdout()

			fmt.Fprintln(out, "Engram status")
			fmt.Fprintf(out, "- binary path: %s\n", fallback(details.BinaryPath, "not found"))
			fmt.Fprintf(out, "- version: %s\n", fallback(details.Version, "unavailable"))
			fmt.Fprintf(out, "- data dir: %s\n", fallback(details.DataDirPath, "n/a"))
			fmt.Fprintf(out, "- data dir found: %s\n", yesNo(details.DataDirFound))
			fmt.Fprintf(out, "- config enabled: %s\n", yesNo(details.ConfigEnabled))
			fmt.Fprintf(out, "- config configured: %s\n", yesNo(details.ConfigConfigured))
			fmt.Fprintf(out, "- config project: %s\n", fallback(cfg.Components.Engram.Project, "n/a"))
			fmt.Fprintf(out, "- final status: %s\n", string(details.Status))

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

func newComponentsEngramConfigureCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Engram component in bitsentry-ai",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			details := components.DetectEngramRuntime(context.Background(), cfg)
			out := cmd.OutOrStdout()

			if !details.BinaryFound {
				fmt.Fprintln(out, "Engram configure failed")
				fmt.Fprintln(out, "- Engram binary was not found in PATH.")
				fmt.Fprintln(out, "- Action: install Engram CLI and ensure `engram` is available in PATH, then rerun this command.")
				return errors.New("engram binary not found")
			}

			project := strings.TrimSpace(cfg.ActiveProfile)
			if project == "" {
				project = "bitsentry-ai"
			}

			nextCfg := cfg
			nextCfg.Components.Engram.Enabled = true
			nextCfg.Components.Engram.Configured = true
			nextCfg.Components.Engram.BinaryPath = details.BinaryPath
			if details.DataDirFound {
				nextCfg.Components.Engram.DataDir = details.DataDirPath
			} else {
				nextCfg.Components.Engram.DataDir = ""
			}
			nextCfg.Components.Engram.Project = project

			if rt.DryRun {
				fmt.Fprintln(out, "[dry-run] Engram configure preview")
				fmt.Fprintln(out, "The following fields would be written:")
				fmt.Fprintf(out, "- components.engram.enabled: %t\n", nextCfg.Components.Engram.Enabled)
				fmt.Fprintf(out, "- components.engram.configured: %t\n", nextCfg.Components.Engram.Configured)
				fmt.Fprintf(out, "- components.engram.binary_path: %s\n", fallback(nextCfg.Components.Engram.BinaryPath, ""))
				fmt.Fprintf(out, "- components.engram.data_dir: %s\n", nextCfg.Components.Engram.DataDir)
				fmt.Fprintf(out, "- components.engram.project: %s\n", nextCfg.Components.Engram.Project)
				if !details.DataDirFound {
					fmt.Fprintln(out, "Warning: ~/.engram directory was not found. Config would still be updated.")
				}
				return nil
			}

			if err := rt.App.ConfigManager.Save(nextCfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			fmt.Fprintln(out, "Engram configured successfully.")
			fmt.Fprintf(out, "- binary path: %s\n", nextCfg.Components.Engram.BinaryPath)
			fmt.Fprintf(out, "- project: %s\n", nextCfg.Components.Engram.Project)
			if details.DataDirFound {
				fmt.Fprintf(out, "- data dir: %s\n", nextCfg.Components.Engram.DataDir)
			} else {
				fmt.Fprintln(out, "- data dir: not found (saved as empty)")
				fmt.Fprintln(out, "Warning: ~/.engram directory was not found. You can continue, but some workflows may need it later.")
			}

			return nil
		},
	}
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func fallback(v string, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func countRequired(cs []components.Component) int {
	total := 0
	for _, c := range cs {
		if c.Required {
			total++
		}
	}
	return total
}

func countOptional(cs []components.Component) int {
	total := 0
	for _, c := range cs {
		if c.Optional {
			total++
		}
	}
	return total
}
