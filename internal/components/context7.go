package components

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"bitsentry-ai/internal/config"
)

var context7CommandCandidates = []string{
	"context7",
	"context7-mcp",
	"mcp-server-context7",
}

type Context7RuntimeDetails struct {
	Status                Status
	Readiness             MCPReadiness
	Detected              bool
	DetectedCommand       string
	DetectedPath          string
	ConfigEnabled         bool
	ConfigConfigured      bool
	ConfigCommand         string
	ConfigPackage         string
	ConfigNotes           string
	ConfigMinimumFieldsOK bool
	DetectionError        string
	Notes                 []string
}

func DetectContext7Runtime(_ context.Context, cfg config.Config) Context7RuntimeDetails {
	details := Context7RuntimeDetails{}

	for _, candidate := range context7CommandCandidates {
		path, err := exec.LookPath(candidate)
		if err == nil {
			details.Detected = true
			details.DetectedCommand = candidate
			details.DetectedPath = path
			break
		}
		if !errors.Is(err, exec.ErrNotFound) {
			details.DetectionError = fmt.Sprintf("binary lookup failed for %s: %v", candidate, err)
			details.Status = StatusError
			details.Readiness = BuildMCPReadiness(StatusError, nil, []string{"context7 binary lookup error"}, []string{"Retry status check without modifying secrets"}, false)
			details.Notes = append(details.Notes, "Unexpected error while checking Context7 command candidates.")
			return details
		}
	}

	context7Cfg := cfg.Components.Context7
	details.ConfigEnabled = context7Cfg.Enabled
	details.ConfigConfigured = context7Cfg.Configured
	details.ConfigCommand = strings.TrimSpace(context7Cfg.Command)
	details.ConfigPackage = strings.TrimSpace(context7Cfg.Package)
	details.ConfigNotes = strings.TrimSpace(context7Cfg.Notes)
	details.ConfigMinimumFieldsOK = details.ConfigCommand != "" && details.ConfigPackage != ""

	if details.ConfigEnabled && details.ConfigConfigured && details.ConfigMinimumFieldsOK {
		details.Status = StatusConfigured
		details.Readiness = BuildMCPReadiness(StatusConfigured,
			[]string{"context7 metadata configured"},
			nil,
			nil,
			true,
		)
		details.Notes = append(details.Notes, "Context7 is configured in bitsentry-ai metadata (no installation validation is performed).")
		if details.Detected {
			details.Notes = append(details.Notes, "A known Context7 command was also found in PATH.")
		}
		return details
	}

	if details.Detected {
		details.Status = StatusDetected
		details.Readiness = BuildMCPReadiness(StatusManualStep,
			[]string{"context7 command found in PATH"},
			[]string{"bitsentry metadata not fully configured"},
			[]string{"Configure Context7 metadata manually in bitsentry-ai"},
			false,
		)
		details.Notes = append(details.Notes, "A known Context7 command was found in PATH.")
		return details
	}

	configPresent := details.ConfigEnabled || details.ConfigConfigured || details.ConfigCommand != "" || details.ConfigPackage != "" || details.ConfigNotes != ""
	if configPresent {
		details.Status = StatusManualStep
		details.Readiness = BuildMCPReadiness(StatusManualStep,
			[]string{"context7 partial metadata present"},
			[]string{"context7 runtime not detected"},
			[]string{"Install or expose Context7 command in PATH", "Complete metadata fields: command and package"},
			false,
		)
		details.Notes = append(details.Notes, "Context7 metadata exists but is not fully configured and no runtime command was detected.")
		return details
	}

	details.Status = StatusMissing
	details.Readiness = BuildMCPReadiness(StatusMissing,
		nil,
		[]string{"context7 runtime and metadata missing"},
		[]string{"Install Context7 manually and configure bitsentry metadata"},
		false,
	)
	details.Notes = append(details.Notes, "No known Context7 command was found in PATH and no Context7 config metadata is present.")
	return details
}
