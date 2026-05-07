package cli

import (
	"fmt"
	"path/filepath"

	"bitsentry-ai/internal/config"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect local configuration",
	}

	configCmd.AddCommand(&cobra.Command{
		Use:   "path",
		Short: "Print config directory and file path",
		Run: func(cmd *cobra.Command, args []string) {
			path := config.ConfigPath()
			fmt.Fprintf(cmd.OutOrStdout(), "Config dir: %s\n", filepath.Dir(path))
			fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", path)
		},
	})

	return configCmd
}
