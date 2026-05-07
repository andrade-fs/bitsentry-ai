package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProfilesCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "profiles",
		Short: "List available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			items := rt.App.ProfileStore.List(cfg.ActiveProfile)
			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Profiles")
			for _, p := range items {
				marker := " "
				if p.IsActive {
					marker = "*"
				}
				fmt.Fprintf(out, "%s %s (%s) - %s\n", marker, p.Name, p.ID, p.Description)
			}
			fmt.Fprintln(out, "* active profile")

			return nil
		},
	}
}

func newProfileCmd(rt *Runtime) *cobra.Command {
	profileCmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage active profile",
	}

	profileCmd.AddCommand(&cobra.Command{
		Use:   "use <name>",
		Short: "Set active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if _, ok := rt.App.ProfileStore.Get(name); !ok {
				return fmt.Errorf("profile %q does not exist; run '%s profiles' to see valid options", name, cmd.Root().Name())
			}

			if rt.DryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "[dry-run] profile %q is valid and would be set as active\n", name)
				return nil
			}

			if err := rt.App.SetActiveProfile(name); err != nil {
				return fmt.Errorf("set active profile: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Active profile updated to %q\n", name)
			return nil
		},
	})

	return profileCmd
}
