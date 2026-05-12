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
	Readiness             MCPReadiness
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
		details.Readiness = BuildMCPReadiness(StatusError, nil, []string{"engram binary detection error"}, []string{"Retry status check; do not change credential files"}, false)
		details.Notes = append(details.Notes, "Unexpected error while checking Engram binary.")
		return details
	}

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		details.DetectionError = fmt.Sprintf("home directory lookup failed: %v", homeErr)
		details.Status = StatusError
		details.Readiness = BuildMCPReadiness(StatusError, nil, []string{"home directory resolution failed"}, []string{"Fix environment permissions/path and retry"}, false)
		details.Notes = append(details.Notes, "Unexpected error while resolving home directory.")
		return details
	}

	details.DataDirPath = filepath.Join(homeDir, ".engram")
	if st, err := os.Stat(details.DataDirPath); err == nil && st.IsDir() {
		details.DataDirFound = true
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		details.DetectionError = fmt.Sprintf("data directory check failed: %v", err)
		details.Status = StatusError
		details.Readiness = BuildMCPReadiness(StatusError, nil, []string{"engram data directory check failed"}, []string{"Fix file-system access and retry"}, false)
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
		details.Readiness = BuildMCPReadiness(StatusConfigured,
			[]string{"engram binary detected", "engram bitsentry metadata configured"},
			nil,
			nil,
			true,
		)
		return details
	}

	if details.BinaryFound || details.DataDirFound {
		details.Status = StatusDetected
		if !details.ConfigEnabled || !details.ConfigConfigured || !details.ConfigMinimumFieldsOK {
			details.Notes = append(details.Notes, "Engram is detected locally but not configured in bitsentry-ai.")
			details.Readiness = BuildMCPReadiness(StatusManualStep,
				[]string{"engram runtime evidence found"},
				[]string{"bitsentry metadata incomplete for Engram"},
				[]string{"Enable/configure Engram in bitsentry-ai metadata", "Keep credential/token files unchanged in this phase"},
				false,
			)
			return details
		}
		details.Readiness = BuildMCPReadiness(StatusDetected,
			[]string{"engram runtime evidence found"},
			nil,
			[]string{"Optional: align bitsentry metadata for fully configured state"},
			false,
		)
		return details
	}

	details.Status = StatusMissing
	details.Notes = append(details.Notes, "Engram binary and ~/.engram directory were not found.")
	details.Readiness = BuildMCPReadiness(StatusMissing,
		nil,
		[]string{"engram runtime not detected"},
		[]string{"Install Engram manually and configure bitsentry metadata"},
		false,
	)
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
