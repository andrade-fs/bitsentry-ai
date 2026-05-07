package cli

import (
	"fmt"

	"bitsentry-ai/internal/app"
	"bitsentry-ai/internal/tui"
	"github.com/spf13/cobra"
)

type Runtime struct {
	App    *app.App
	DryRun bool
}

func NewRootCmd() (*cobra.Command, error) {
	a, err := app.New()
	if err != nil {
		return nil, fmt.Errorf("init app: %w", err)
	}

	rt := &Runtime{App: a}

	rootCmd := &cobra.Command{
		Use:   app.Name,
		Short: "Bootstrap CLI for BitSentry AI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return tui.Run(rt.App, rt.DryRun)
		},
	}

	rootCmd.PersistentFlags().BoolVar(&rt.DryRun, "dry-run", false, "Show actions without changing local state")

	rootCmd.AddCommand(
		newVersionCmd(),
		newDoctorCmd(rt),
		newAgentsCmd(rt),
		newProfilesCmd(rt),
		newProfileCmd(rt),
		newComponentsCmd(rt),
		newCapabilitiesCmd(rt),
		newConfigCmd(),
	)

	return rootCmd, nil
}
