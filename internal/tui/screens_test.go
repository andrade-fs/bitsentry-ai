package tui

import (
	"strings"
	"testing"

	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/components"
)

func TestRenderInstallReviewShowsReadinessSummary(t *testing.T) {
	m := model{}
	m.install = installWizardState{
		CurrentStep: installStepReview,
		InstallMode: installModeEverything,
		MCPConfigPreview: capabilities.OpenCodeMCPConfigPreview{
			CurrentConfigState:       "readable_with_mcp",
			Exists:                   true,
			Readable:                 true,
			CurrentMCPConfigDetected: true,
			MCPReadinessState:        "configured",
			WouldWrite:               false,
			RequiresConfirmation:     true,
			BackupRequired:           true,
			ProposedSafeChanges:      []string{"Ensure mcp.context7 metadata entry exists without sensitive values (command/args only)."},
		},
		MCPReadiness: map[string]components.MCPReadiness{
			"engram":   components.BuildMCPReadiness(components.StatusConfigured, []string{"binary"}, nil, nil, true),
			"context7": components.BuildMCPReadiness(components.StatusManualStep, []string{"partial metadata"}, []string{"runtime missing"}, []string{"manual install"}, false),
		},
		TargetSelected: true,
	}
	out := renderInstall(m)
	for _, want := range []string{"Detected MCP readiness", "MCP config preview (PREVIEW ONLY)", "would_write: no", "requires_confirmation: yes", "backup_required: yes", "Engram status: configured", "Context7 status: manual_step_needed", "manual install"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output", want)
		}
	}
}

func TestRenderInstallDoneShowsPreviewContract(t *testing.T) {
	m := model{}
	m.install = installWizardState{
		CurrentStep: installStepDone,
		ResultStatus: "PASS WITH NOTES",
		MCPConfigPreview: capabilities.OpenCodeMCPConfigPreview{
			CurrentConfigState:   "missing_opencode_json",
			WouldWrite:           false,
			RequiresConfirmation: true,
			BackupRequired:       true,
			ManualSteps:          []string{"Create/repair opencode.json manually."},
		},
		MCPReadiness: map[string]components.MCPReadiness{
			"engram":   components.BuildMCPReadiness(components.StatusDetected, []string{"runtime"}, nil, nil, false),
			"context7": components.BuildMCPReadiness(components.StatusMissing, nil, nil, nil, false),
		},
	}
	out := renderInstall(m)
	for _, want := range []string{"MCP config preview contract (PREVIEW ONLY)", "requires_confirmation: yes", "backup_required: yes", "manual steps:"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in done output", want)
		}
	}
}
