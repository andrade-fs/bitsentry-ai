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
		details.Notes = append(details.Notes, "Context7 is configured in bitsentry-ai metadata (no installation validation is performed).")
		if details.Detected {
			details.Notes = append(details.Notes, "A known Context7 command was also found in PATH.")
		}
		return details
	}

	if details.Detected {
		details.Status = StatusDetected
		details.Notes = append(details.Notes, "A known Context7 command was found in PATH.")
		return details
	}

	configPresent := details.ConfigEnabled || details.ConfigConfigured || details.ConfigCommand != "" || details.ConfigPackage != "" || details.ConfigNotes != ""
	if configPresent {
		details.Status = StatusDetected
		details.Notes = append(details.Notes, "Context7 metadata exists but is not fully configured and no runtime command was detected.")
		return details
	}

	details.Status = StatusMissing
	details.Notes = append(details.Notes, "No known Context7 command was found in PATH and no Context7 config metadata is present.")
	return details
}
