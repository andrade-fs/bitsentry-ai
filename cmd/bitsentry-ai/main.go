package main

import (
	"fmt"
	"os"

	"bitsentry-ai/internal/cli"
)

func main() {
	rootCmd, err := cli.NewRootCmd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap error: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "run error: %v\n", err)
		os.Exit(1)
	}
}
