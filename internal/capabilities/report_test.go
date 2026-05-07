package capabilities

import (
	"strings"
	"testing"

	"bitsentry-ai/internal/config"
)

func TestApplySummaryIncludesManagedSkippedAndDeclarative(t *testing.T) {
	plan := Plan{TargetAgent: "opencode", Preset: "custom"}
	projection := OpenCodeProjection{
		ManagedMCPs:       []string{"engram", "context7"},
		SkippedMCPs:       []string{"postgres"},
		DeclarativeSkills: []string{"bitsentry-sdd"},
		DeclarativeFlows:  []string{"sdd", "notes"},
	}

	summary := ApplySummary(plan, projection)
	for _, expected := range []string{"managed MCPs to apply: engram, context7", "skipped MCPs: postgres", "declarative skills: bitsentry-sdd", "declarative flows: sdd, notes"} {
		if !strings.Contains(summary, expected) {
			t.Fatalf("summary missing %q\n%s", expected, summary)
		}
	}
}

func TestWriteApplyReportWritesTimestampAndLatest(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg := config.Config{}
	cfg.Components.Preset = "custom"
	plan := Plan{TargetAgent: "opencode", Preset: "custom"}
	projection := OpenCodeProjection{ManagedMCPs: []string{"engram"}, SkippedMCPs: []string{"postgres"}}

	reportPath, err := WriteApplyReport(cfg, plan, projection, "dry-run", "preview", "No files were modified.")
	if err != nil {
		t.Fatalf("write report failed: %v", err)
	}
	if !strings.Contains(reportPath, ".bitsentry-ai/exports/capabilities/apply/") {
		t.Fatalf("unexpected report path: %s", reportPath)
	}

	latestRaw, err := ReadLatestApplyReport()
	if err != nil {
		t.Fatalf("read latest report failed: %v", err)
	}
	if !strings.Contains(latestRaw, "mode: dry-run") {
		t.Fatalf("latest report missing mode: %s", latestRaw)
	}
}
