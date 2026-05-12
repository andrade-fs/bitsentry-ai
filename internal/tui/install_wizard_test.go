package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/app"
	"bitsentry-ai/internal/config"
	"bitsentry-ai/internal/profiles"
)

func TestSelectedMCPListOrder(t *testing.T) {
	ids := selectedMCPList(map[string]bool{"context7": true, "engram": true})
	if len(ids) != 2 || ids[0] != "engram" || ids[1] != "context7" {
		t.Fatalf("unexpected selected MCP ordering: %#v", ids)
	}
}

func TestComponentStatusLabel(t *testing.T) {
	if got := componentStatusLabel(true, false); got != "configured" {
		t.Fatalf("configured status mismatch: %s", got)
	}
	if got := componentStatusLabel(false, true); got != "available" {
		t.Fatalf("available status mismatch: %s", got)
	}
	if got := componentStatusLabel(false, false); got != "not configured" {
		t.Fatalf("not configured status mismatch: %s", got)
	}
}

func TestDetectOpenCodeInstallStatusPrefersExistingConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	opencodeA := filepath.Join(home, ".opencode", "bitsentry")
	opencodeB := filepath.Join(home, ".config", "opencode", "bitsentry")
	if err := os.MkdirAll(opencodeA, 0o755); err != nil {
		t.Fatalf("mkdir opencodeA: %v", err)
	}
	if err := os.MkdirAll(opencodeB, 0o755); err != nil {
		t.Fatalf("mkdir opencodeB: %v", err)
	}
	if err := os.WriteFile(filepath.Join(opencodeB, "OPENCODE_USAGE.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write usage: %v", err)
	}
	if err := os.WriteFile(filepath.Join(opencodeB, "skill-registry.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	status := detectOpenCodeInstallStatus()
	if status.ConfigRoot == "" {
		t.Fatalf("expected config root to be resolved")
	}
	if status.ConfigRoot != filepath.Join(home, ".config", "opencode") {
		t.Fatalf("expected .config/opencode preference, got %s", status.ConfigRoot)
	}
	if !status.PackInstalled || !status.UsageExists || !status.RegistryExists {
		t.Fatalf("expected installed pack flags true, got installed=%t usage=%t registry=%t", status.PackInstalled, status.UsageExists, status.RegistryExists)
	}
}

func TestLoadInstallWizard_DefaultSelections(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	if err := os.MkdirAll(filepath.Join(home, ".config", "opencode"), 0o755); err != nil {
		t.Fatalf("mkdir opencode config: %v", err)
	}

	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	a := &app.App{ConfigManager: cm, ProfileStore: profiles.NewInMemoryStore(), AgentRegistry: agents.NewRegistry(agents.OpenCodeDetector{})}
	m := newModel(a, false)
	m.loadInstallWizard()

	if m.install.OpenCodeDetected && !m.install.TargetSelected {
		t.Fatalf("expected OpenCode target selected by default when detected")
	}
	if !m.install.OpenCodeDetected && m.install.TargetSelected {
		t.Fatalf("expected target unselected when OpenCode is not detected")
	}
	if m.install.SelectedMCPs["context7"] {
		t.Fatalf("expected context7 default selection false when not configured")
	}
	if m.install.InstallMode != installModeEverything {
		t.Fatalf("expected default install mode to be Install Everything when pack not installed")
	}
}

func TestInstallWizardEnterNavigationAndBlock(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	if err := os.MkdirAll(filepath.Join(home, ".config", "opencode"), 0o755); err != nil {
		t.Fatalf("mkdir opencode config: %v", err)
	}

	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		t.Fatalf("load config: %v", err)
	}
	a := &app.App{ConfigManager: cm, ProfileStore: profiles.NewInMemoryStore(), AgentRegistry: agents.NewRegistry(agents.OpenCodeDetector{})}
	m := newModel(a, false)
	m.loadInstallWizard()

	m.install.TargetSelected = false
	if !m.installEnterAction() {
		t.Fatalf("expected enter handler true")
	}
	if m.install.CurrentStep != installStepTarget {
		t.Fatalf("expected to stay on step 1 when target not selected")
	}

	m.install.TargetSelected = true
	m.installEnterAction()
	if m.install.CurrentStep != installStepMode {
		t.Fatalf("expected step 2, got %d", m.install.CurrentStep)
	}
	m.installEnterAction()
	if m.install.CurrentStep != installStepReview {
		t.Fatalf("expected step 3, got %d", m.install.CurrentStep)
	}
}

func TestInstallWizardSpaceTogglesTarget(t *testing.T) {
	m := model{screen: screenInstall}
	m.install = installWizardState{CurrentStep: installStepTarget, OpenCodeDetected: true, TargetSelected: true}

	if !m.handleInstallKey(" ") {
		t.Fatalf("expected install key handled")
	}
	if m.install.TargetSelected {
		t.Fatalf("expected target to toggle off")
	}
	if !m.handleInstallKey(" ") {
		t.Fatalf("expected install key handled on second toggle")
	}
	if !m.install.TargetSelected {
		t.Fatalf("expected target to toggle on")
	}
}

func TestInstallWizardInstallModeSelectionPersistsAcrossSteps(t *testing.T) {
	m := model{screen: screenInstall}
	m.install = installWizardState{
		CurrentStep:      installStepMode,
		Cursor:           installModeEverything,
		InstallMode:      installModeEverything,
		OpenCodeDetected: true,
		TargetSelected:   true,
	}

	m.handleInstallKey("down")
	m.handleInstallKey(" ")
	if m.install.InstallMode != installModePackOnly {
		t.Fatalf("expected pack-only mode selected")
	}

	m.installEnterAction() // mode -> review
	if m.install.CurrentStep != installStepReview {
		t.Fatalf("expected review step")
	}
	m.install.CurrentStep = installStepMode // simulate back
	if m.install.InstallMode != installModePackOnly {
		t.Fatalf("expected install mode selection persisted")
	}
}

func TestInstallModeHelpers(t *testing.T) {
	if got := defaultInstallMode(false); got != installModeEverything {
		t.Fatalf("default mode mismatch for fresh install: %d", got)
	}
	if got := defaultInstallMode(true); got != installModeUpdateReinstall {
		t.Fatalf("default mode mismatch for installed pack: %d", got)
	}
	if got := presetForMode(installModeEverything); got != "bitsentry-full" {
		t.Fatalf("preset mismatch for everything mode: %s", got)
	}
	if got := presetForMode(installModePackOnly); got != "bitsentry-dev" {
		t.Fatalf("preset mismatch for pack-only mode: %s", got)
	}
	reg, cmd, skills := nativeOptionsForMode(installModePackOnly)
	if reg || cmd || skills {
		t.Fatalf("pack-only mode should disable native integration options")
	}
}

func TestInstallWizardReviewBlocksWhenNoTargetSelected(t *testing.T) {
	m := model{screen: screenInstall}
	m.install = installWizardState{CurrentStep: installStepReview, TargetSelected: false}
	if !m.installEnterAction() {
		t.Fatalf("expected enter action handled")
	}
	if m.install.CurrentStep != installStepReview {
		t.Fatalf("expected to remain in review step when no target selected")
	}
}

func TestBuildOpenCodeDogfoodingPrompt_MultilineIncludesFilesAndConstraints(t *testing.T) {
	root := "/tmp/opencode/bitsentry"
	prompt := buildOpenCodeDogfoodingPrompt(root)
	if !strings.Contains(prompt, root+"/OPENCODE_USAGE.md") {
		t.Fatalf("expected usage file in prompt")
	}
	if !strings.Contains(prompt, root+"/skill-registry.md") {
		t.Fatalf("expected registry file in prompt")
	}
	checks := []string{
		"Use the Bitsentry SDD flow",
		"Do not modify code yet.",
		"Do not modify opencode.json.",
		"Do not execute runtime flows.",
	}
	for _, c := range checks {
		if !strings.Contains(prompt, c) {
			t.Fatalf("prompt missing %q", c)
		}
	}
	if !strings.Contains(prompt, "\n") {
		t.Fatalf("expected multiline prompt")
	}
}
