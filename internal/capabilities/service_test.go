package capabilities

import (
	"os"
	"path/filepath"
	"testing"

	"bitsentry-ai/internal/app"
	"bitsentry-ai/internal/config"
	"bitsentry-ai/internal/profiles"
)

func TestServiceSaveLoadValidateBuildPlan(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		t.Fatalf("load initial config: %v", err)
	}

	a := &app.App{ConfigManager: cm, ProfileStore: profiles.NewInMemoryStore()}
	svc := NewService(a, "/nonexistent/bitsentry-ai")

	draft := SelectionDraft{
		TargetAgent: "opencode",
		Preset:      "custom",
		MCPs:        []string{"engram", "context7"},
		Skills:      []string{"bitsentry-sdd"},
		Flows:       []string{"sdd", "notes"},
	}

	if err := svc.SaveSelection(draft); err != nil {
		t.Fatalf("save selection: %v", err)
	}

	loaded, err := svc.LoadSelection()
	if err != nil {
		t.Fatalf("load selection: %v", err)
	}
	if loaded.TargetAgent != "opencode" {
		t.Fatalf("expected opencode target, got %q", loaded.TargetAgent)
	}

	vr, err := svc.ValidateSelection(draft)
	if err != nil {
		t.Fatalf("validate selection err: %v", err)
	}
	if !vr.Valid {
		t.Fatalf("expected valid draft, got issues: %#v", vr.Issues)
	}

	planRes, err := svc.BuildPlan(draft)
	if err != nil {
		t.Fatalf("build plan err: %v", err)
	}
	if planRes.Plan.TargetAgent != "opencode" {
		t.Fatalf("expected opencode in plan, got %q", planRes.Plan.TargetAgent)
	}
	if len(planRes.Projection.ManagedMCPs) == 0 {
		t.Fatalf("expected managed MCPs in projection")
	}
}

func TestServiceValidateRejectsInvalidDraft(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		t.Fatalf("load initial config: %v", err)
	}
	a := &app.App{ConfigManager: cm, ProfileStore: profiles.NewInMemoryStore()}
	svc := NewService(a, filepath.Join(tmp, "bitsentry-ai"))

	vr, err := svc.ValidateSelection(SelectionDraft{TargetAgent: "cursor", MCPs: []string{"unknown"}})
	if err != nil {
		t.Fatalf("validate selection returned unexpected error: %v", err)
	}
	if vr.Valid {
		t.Fatalf("expected invalid selection")
	}
}

func TestServiceApplyDryRunWithoutExecutableFailsSafely(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		t.Fatalf("load initial config: %v", err)
	}
	a := &app.App{ConfigManager: cm, ProfileStore: profiles.NewInMemoryStore()}
	svc := NewService(a, "")

	_, err := svc.ApplyDryRun(SelectionDraft{TargetAgent: "opencode", Preset: "custom"})
	if err == nil {
		t.Fatalf("expected error when executable path is missing")
	}
	if _, statErr := os.Stat(config.ConfigPath()); statErr != nil {
		t.Fatalf("config path should still exist: %v", statErr)
	}
}
