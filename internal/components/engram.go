package components

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"bitsentry-ai/internal/config"
)

type EngramRuntimeDetails struct {
	Status                Status
	BinaryFound           bool
	BinaryPath            string
	Version               string
	VersionAvailable      bool
	DataDirFound          bool
	DataDirPath           string
	ConfigEnabled         bool
	ConfigConfigured      bool
	ConfigMinimumFieldsOK bool
	FullyConfigured       bool
	DetectionError        string
	Notes                 []string
}

func DetectEngramRuntime(ctx context.Context, cfg config.Config) EngramRuntimeDetails {
	details := EngramRuntimeDetails{}

	binaryPath, binaryErr := exec.LookPath("engram")
	if binaryErr == nil {
		details.BinaryFound = true
		details.BinaryPath = binaryPath
		version, found := detectEngramVersion(ctx, binaryPath)
		details.Version = version
		details.VersionAvailable = found
		if !found {
			details.Notes = append(details.Notes, "Engram binary found but version is unavailable.")
		}
	} else if !errors.Is(binaryErr, exec.ErrNotFound) {
		details.DetectionError = fmt.Sprintf("binary lookup failed: %v", binaryErr)
		details.Status = StatusError
		details.Notes = append(details.Notes, "Unexpected error while checking Engram binary.")
		return details
	}

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		details.DetectionError = fmt.Sprintf("home directory lookup failed: %v", homeErr)
		details.Status = StatusError
		details.Notes = append(details.Notes, "Unexpected error while resolving home directory.")
		return details
	}

	details.DataDirPath = filepath.Join(homeDir, ".engram")
	if st, err := os.Stat(details.DataDirPath); err == nil && st.IsDir() {
		details.DataDirFound = true
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		details.DetectionError = fmt.Sprintf("data directory check failed: %v", err)
		details.Status = StatusError
		details.Notes = append(details.Notes, "Unexpected error while checking Engram data directory.")
		return details
	}

	engramCfg := cfg.Components.Engram
	details.ConfigEnabled = engramCfg.Enabled
	details.ConfigConfigured = engramCfg.Configured
	details.ConfigMinimumFieldsOK = strings.TrimSpace(engramCfg.BinaryPath) != "" &&
		strings.TrimSpace(engramCfg.Project) != ""

	details.FullyConfigured = details.BinaryFound && details.ConfigEnabled && details.ConfigConfigured && details.ConfigMinimumFieldsOK

	if details.FullyConfigured {
		details.Status = StatusConfigured
		details.Notes = append(details.Notes, "Engram runtime and bitsentry-ai config are consistent.")
		return details
	}

	if details.BinaryFound || details.DataDirFound {
		details.Status = StatusDetected
		if !details.ConfigEnabled || !details.ConfigConfigured || !details.ConfigMinimumFieldsOK {
			details.Notes = append(details.Notes, "Engram is detected locally but not configured in bitsentry-ai.")
		}
		return details
	}

	details.Status = StatusMissing
	details.Notes = append(details.Notes, "Engram binary and ~/.engram directory were not found.")
	return details
}

func detectEngramVersion(ctx context.Context, binaryPath string) (string, bool) {
	commands := [][]string{
		{binaryPath, "--version"},
		{binaryPath, "version"},
	}

	for _, c := range commands {
		cmd := exec.CommandContext(ctx, c[0], c[1:]...)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		version := strings.TrimSpace(string(out))
		if version != "" {
			return version, true
		}
	}

	return "", false
}
