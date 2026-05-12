package components

import (
	"context"
	"testing"

	"bitsentry-ai/internal/config"
)

func TestBuildMCPReadinessCopiesSlices(t *testing.T) {
	evidence := []string{"a"}
	r := BuildMCPReadiness(StatusDetected, evidence, nil, nil, false)
	evidence[0] = "b"
	if r.DetectedEvidence[0] != "a" {
		t.Fatalf("expected readiness to copy slices")
	}
}

func TestContext7ReadinessMissing(t *testing.T) {
	d := DetectContext7Runtime(context.Background(), config.Config{})
	if d.Readiness.Status != StatusMissing {
		t.Fatalf("expected missing readiness, got %s", d.Readiness.Status)
	}
	if d.Readiness.SafeUsable {
		t.Fatalf("missing readiness cannot be safe usable")
	}
}

func TestEngramReadinessMissing(t *testing.T) {
	d := DetectEngramRuntime(context.Background(), config.Config{})
	if d.Readiness.Status != StatusMissing && d.Readiness.Status != StatusDetected && d.Readiness.Status != StatusManualStep {
		t.Fatalf("expected missing/detected readiness, got %s", d.Readiness.Status)
	}
}

func TestMCPRegistryIncludesModeledOnlyAndUnsupported(t *testing.T) {
	entries, _ := MCPRegistry(context.Background(), config.Config{}, EngramRuntimeDetails{}, Context7RuntimeDetails{})
	statusByID := map[string]Status{}
	for _, e := range entries {
		statusByID[e.ID] = e.Status
	}
	if statusByID["postgres"] != StatusModeledOnly {
		t.Fatalf("expected postgres modeled_only")
	}
	if statusByID["browser"] != StatusNotImplemented {
		t.Fatalf("expected browser not_implemented")
	}
}
