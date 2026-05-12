package tui

import (
	"strings"
	"testing"

	"bitsentry-ai/internal/components"
)

func TestRenderInstallReviewShowsReadinessSummary(t *testing.T) {
	m := model{}
	m.install = installWizardState{
		CurrentStep: installStepReview,
		InstallMode: installModeEverything,
		MCPReadiness: map[string]components.MCPReadiness{
			"engram":   components.BuildMCPReadiness(components.StatusConfigured, []string{"binary"}, nil, nil, true),
			"context7": components.BuildMCPReadiness(components.StatusManualStep, []string{"partial metadata"}, []string{"runtime missing"}, []string{"manual install"}, false),
		},
		TargetSelected: true,
	}
	out := renderInstall(m)
	for _, want := range []string{"Detected MCP readiness", "Engram status: configured", "Context7 status: manual_step_needed", "manual install"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in output", want)
		}
	}
}
