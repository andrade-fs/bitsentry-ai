package cli

import (
	"fmt"
	"path/filepath"

	"bitsentry-ai/internal/system"
	"github.com/spf13/cobra"
)

func newDoctorCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect local environment and dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			info := system.DetectSystem()
			shell := system.DetectShell()
			deps := system.CheckDependencies()

			pkgMgr := "not detected"
			for _, d := range deps {
				if !d.Found {
					continue
				}
				switch d.Name {
				case "brew", "apt", "yum", "pacman":
					pkgMgr = fmt.Sprintf("%s (%s)", d.Name, d.Path)
				}
			}

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Doctor report")
			fmt.Fprintf(out, "- OS: %s\n", info.OS)
			fmt.Fprintf(out, "- Arch: %s\n", info.Arch)
			fmt.Fprintf(out, "- Shell: %s\n", shell)
			fmt.Fprintf(out, "- Package manager: %s\n", pkgMgr)
			fmt.Fprintf(out, "- Config dir: %s\n", filepath.Dir(rt.App.ConfigManager.Path()))
			fmt.Fprintf(out, "- Config file: %s\n", rt.App.ConfigManager.Path())
			fmt.Fprintf(out, "- Active profile: %s\n", cfg.ActiveProfile)
			fmt.Fprintln(out, "- Dependencies:")
			for _, d := range deps {
				status := "not found"
				if d.Found {
					status = "found"
				}
				mandatory := "optional"
				if d.Mandatory {
					mandatory = "required"
				}
				fmt.Fprintf(out, "  - %s: %s (%s)", d.Name, status, mandatory)
				if d.Path != "" {
					fmt.Fprintf(out, " [%s]", d.Path)
				}
				fmt.Fprintln(out)
			}

			return nil
		},
	}
}
