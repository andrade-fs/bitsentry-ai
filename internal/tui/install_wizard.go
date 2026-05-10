package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/components"
)

const (
	installStepTarget = iota
	installStepCapabilities
	installStepComponents
	installStepReview
	installStepInstall
	installStepDone
	installStepCount = 6
)

type installWizardState struct {
	CurrentStep int
	Cursor      int

	OpenCodeDetected  bool
	OpenCodeBinary    string
	OpenCodeConfig    string
	BitsentryPackRoot string
	PackInstalled     bool
	UsageExists       bool
	RegistryExists    bool

	TargetSelected bool
	PresetOrder    []string
	Preset         string
	SelectedMCPs   map[string]bool
	MCPStatus      map[string]string
	RegisterAgent  bool
	InstallCommands bool
	InstallNativeSkills bool
	ConfigureMCP bool

	ResultStatus string
	ResultNotes  []string
	NextPrompt   string
}

func (m *model) loadInstallWizard() {
	status := detectOpenCodeInstallStatus()
	engramCfg, context7Cfg := detectComponentConfiguredStatus(m)
	engramRuntime, context7Runtime := detectComponentRuntimeStatus(m)

	engramSelected := engramCfg || engramRuntime
	context7Selected := context7Cfg

	m.install = installWizardState{
		CurrentStep:       installStepTarget,
		Cursor:            0,
		OpenCodeDetected:  status.Detected,
		OpenCodeBinary:    status.BinaryPath,
		OpenCodeConfig:    status.ConfigRoot,
		BitsentryPackRoot: status.PackRoot,
		PackInstalled:     status.PackInstalled,
		UsageExists:       status.UsageExists,
		RegistryExists:    status.RegistryExists,
		TargetSelected:    status.Detected,
		PresetOrder:       []string{"bitsentry-dev", "bitsentry-full", "bitsentry-research", "bitsentry-blog"},
		Preset:            "bitsentry-dev",
		SelectedMCPs: map[string]bool{
			"engram":   engramSelected,
			"context7": context7Selected,
		},
		MCPStatus: map[string]string{
			"engram":   componentStatusLabel(engramCfg, engramRuntime),
			"context7": componentStatusLabel(context7Cfg, context7Runtime),
		},
		RegisterAgent: true,
		InstallCommands: true,
		InstallNativeSkills: true,
		ConfigureMCP: false,
		ResultNotes: []string{},
	}
}

func (m *model) refreshInstallWizardPreserveSelections() {
	prev := m.install
	status := detectOpenCodeInstallStatus()
	engramCfg, context7Cfg := detectComponentConfiguredStatus(m)
	engramRuntime, context7Runtime := detectComponentRuntimeStatus(m)

	targetSelected := status.Detected
	if prev.Preset != "" {
		targetSelected = prev.TargetSelected
	}
	if !status.Detected {
		targetSelected = false
	}

	selectedMCPs := map[string]bool{
		"engram":   engramCfg || engramRuntime,
		"context7": context7Cfg,
	}
	if prev.SelectedMCPs != nil {
		selectedMCPs["engram"] = prev.SelectedMCPs["engram"]
		selectedMCPs["context7"] = prev.SelectedMCPs["context7"]
	}

	preset := "bitsentry-dev"
	if prev.Preset != "" {
		preset = prev.Preset
	}
	currentStep := installStepTarget
	if prev.CurrentStep >= installStepTarget && prev.CurrentStep < installStepCount {
		currentStep = prev.CurrentStep
	}

	m.install = installWizardState{
		CurrentStep:       currentStep,
		Cursor:            prev.Cursor,
		OpenCodeDetected:  status.Detected,
		OpenCodeBinary:    status.BinaryPath,
		OpenCodeConfig:    status.ConfigRoot,
		BitsentryPackRoot: status.PackRoot,
		PackInstalled:     status.PackInstalled,
		UsageExists:       status.UsageExists,
		RegistryExists:    status.RegistryExists,
		TargetSelected:    targetSelected,
		PresetOrder:       []string{"bitsentry-dev", "bitsentry-full", "bitsentry-research", "bitsentry-blog"},
		Preset:            preset,
		SelectedMCPs:      selectedMCPs,
		MCPStatus: map[string]string{
			"engram":   componentStatusLabel(engramCfg, engramRuntime),
			"context7": componentStatusLabel(context7Cfg, context7Runtime),
		},
		RegisterAgent: prev.RegisterAgent,
		InstallCommands: prev.InstallCommands,
		InstallNativeSkills: prev.InstallNativeSkills,
		ConfigureMCP: prev.ConfigureMCP,
		ResultStatus: prev.ResultStatus,
		ResultNotes:  append([]string{}, prev.ResultNotes...),
		NextPrompt:   prev.NextPrompt,
	}
}

func (m *model) handleInstallKey(key string) bool {
	w := &m.install
	switch key {
	case "up", "k":
		if w.CurrentStep == installStepComponents && w.Cursor > 0 {
			w.Cursor--
		}
		return true
	case "down", "j":
		if w.CurrentStep == installStepComponents && w.Cursor < 1 {
			w.Cursor++
		}
		return true
	case "space", " ":
		switch w.CurrentStep {
		case installStepTarget:
			if w.OpenCodeDetected {
				w.TargetSelected = !w.TargetSelected
			}
		case installStepCapabilities:
			w.ShiftPreset(1)
		case installStepComponents:
			if w.Cursor == 0 {
				w.ToggleMCP("engram")
			} else {
				w.ToggleMCP("context7")
			}
		}
		return true
	case "[":
		if w.CurrentStep == installStepCapabilities {
			w.ShiftPreset(-1)
		}
		return true
	case "]":
		if w.CurrentStep == installStepCapabilities {
			w.ShiftPreset(1)
		}
		return true
	case "r":
		m.refreshInstallWizardPreserveSelections()
		m.message = "Install wizard refreshed."
		return true
	case "enter":
		return m.installEnterAction()
	}
	return false
}

func (m *model) installEnterAction() bool {
	w := &m.install
	switch w.CurrentStep {
	case installStepTarget:
		if !w.TargetSelected {
			m.errMsg = "Cannot continue: select at least one target agent."
			return true
		}
		w.CurrentStep = installStepCapabilities
		m.errMsg = ""
		return true
	case installStepCapabilities:
		w.CurrentStep = installStepComponents
		return true
	case installStepComponents:
		w.CurrentStep = installStepReview
		return true
	case installStepReview:
		if !w.TargetSelected {
			m.errMsg = "Cannot install yet: select at least one target agent."
			return true
		}
		w.CurrentStep = installStepInstall
		m.errMsg = ""
		return true
	case installStepInstall:
		m.runInstallWizard()
		w.CurrentStep = installStepDone
		return true
	case installStepDone:
		return true
	}
	return true
}

func (w *installWizardState) ShiftPreset(delta int) {
	if len(w.PresetOrder) == 0 {
		return
	}
	idx := 0
	for i, p := range w.PresetOrder {
		if p == w.Preset {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(w.PresetOrder)) % len(w.PresetOrder)
	w.Preset = w.PresetOrder[idx]
}

func (w *installWizardState) ToggleMCP(id string) {
	if w.SelectedMCPs == nil {
		w.SelectedMCPs = map[string]bool{}
	}
	if id != "engram" && id != "context7" {
		return
	}
	w.SelectedMCPs[id] = !w.SelectedMCPs[id]
}

func (m *model) runInstallWizard() {
	if !m.install.TargetSelected {
		m.errMsg = "Install blocked: OpenCode target is not selected."
		m.install.ResultStatus = "FAIL"
		return
	}
	if strings.TrimSpace(m.install.OpenCodeConfig) == "" {
		m.errMsg = "Install blocked: OpenCode config root is not resolved."
		m.install.ResultStatus = "FAIL"
		return
	}

	preset, ok := capabilities.PresetByID(m.install.Preset, capabilities.DefaultPresets())
	if !ok {
		m.errMsg = fmt.Sprintf("Install blocked: preset %q not found.", m.install.Preset)
		m.install.ResultStatus = "FAIL"
		return
	}

	selectedMCPs := []string{}
	for _, id := range preset.MCPs {
		if (id == "engram" || id == "context7") && m.install.SelectedMCPs[id] {
			selectedMCPs = append(selectedMCPs, id)
		}
	}
	draft := capabilities.SelectionDraft{
		TargetAgent: "opencode",
		Preset:      preset.ID,
		MCPs:        selectedMCPs,
		Skills:      append([]string{}, preset.Skills...),
		Flows:       append([]string{}, preset.Flows...),
	}
	if _, err := m.capSvc.ValidateSelection(draft); err != nil {
		m.errMsg = fmt.Sprintf("Install validation failed: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}
	if err := m.capSvc.SaveSelection(draft); err != nil {
		m.errMsg = fmt.Sprintf("Could not save install selection: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}

	catalog, err := capabilities.DiscoverAssets(".")
	if err != nil {
		m.errMsg = fmt.Sprintf("Asset discovery failed: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}
	selectedIDs := append(append([]string{}, draft.Flows...), draft.Skills...)
	if err := capabilities.ValidateOpenCodeSelectionIDs(catalog, selectedIDs); err != nil {
		m.errMsg = fmt.Sprintf("Install selection invalid for export: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}
	projection, err := capabilities.BuildOpenCodeExportProjection(catalog, selectedIDs)
	if err != nil {
		m.errMsg = fmt.Sprintf("Export projection failed: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}
	res, err := capabilities.ExecuteOpenCodeSkillsExport(projection, m.install.OpenCodeConfig, false)
	if err != nil {
		m.errMsg = fmt.Sprintf("Export failed: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}

	usagePath := filepath.Join(res.TargetRoot, "OPENCODE_USAGE.md")
	registryPath := filepath.Join(res.TargetRoot, "skill-registry.md")
	if !fileExistsPath(usagePath) || !fileExistsPath(registryPath) {
		m.errMsg = "Export verification failed: missing OPENCODE_USAGE.md or skill-registry.md"
		m.install.ResultStatus = "FAIL"
		return
	}
	nativeRes, err := capabilities.ExecuteOpenCodeNativeIntegration(projection, m.install.OpenCodeConfig, capabilities.OpenCodeNativeOptions{
		RegisterAgent:       m.install.RegisterAgent,
		InstallCommands:     m.install.InstallCommands,
		InstallNativeSkills: m.install.InstallNativeSkills,
		ConfigureMCP:        m.install.ConfigureMCP,
	})
	if err != nil {
		m.errMsg = fmt.Sprintf("Native integration failed: %v", err)
		m.install.ResultStatus = "FAIL"
		return
	}
	if m.install.RegisterAgent {
		if !fileExistsPath(nativeRes.AgentPromptFile) || !fileExistsPath(nativeRes.EntrypointFile) {
			m.errMsg = "Native integration verification failed: missing bitsentry agent/entrypoint files"
			m.install.ResultStatus = "FAIL"
			return
		}
	}
	if m.install.InstallNativeSkills && len(nativeRes.NativeSkillFiles) == 0 {
		m.errMsg = "Native integration verification failed: no native skill files installed"
		m.install.ResultStatus = "FAIL"
		return
	}

	m.install.BitsentryPackRoot = res.TargetRoot
	m.install.UsageExists = true
	m.install.RegistryExists = true
	m.install.PackInstalled = true
	m.install.ResultNotes = []string{}
	if len(selectedMCPs) > 0 {
		m.install.ResultNotes = append(m.install.ResultNotes, "MCP apply remains planned/manual in this wizard phase; opencode.json was updated only for bitsentry agent/commands registration.")
	}
	if len(projection.Warnings) > 0 {
		m.install.ResultNotes = append(m.install.ResultNotes, projection.Warnings...)
	}
	if nativeRes.ConfigBackupPath != "" {
		m.install.ResultNotes = append(m.install.ResultNotes, fmt.Sprintf("OpenCode config backup: %s", nativeRes.ConfigBackupPath))
	}
	if nativeRes.NativeBackupPath != "" {
		m.install.ResultNotes = append(m.install.ResultNotes, fmt.Sprintf("Native skills backup: %s", nativeRes.NativeBackupPath))
	}
	for _, w := range nativeRes.Warnings {
		m.install.ResultNotes = append(m.install.ResultNotes, w)
	}
	if len(m.install.ResultNotes) > 0 {
		m.install.ResultStatus = "PASS WITH NOTES"
	} else {
		m.install.ResultStatus = "PASS"
	}
	m.install.NextPrompt = buildOpenCodeDogfoodingPrompt(res.TargetRoot)
	m.message = fmt.Sprintf("Install completed: %s", m.install.ResultStatus)
	m.errMsg = ""
	m.refreshData()
}

func buildOpenCodeDogfoodingPrompt(packRoot string) string {
	usagePath := filepath.Join(packRoot, "OPENCODE_USAGE.md")
	registryPath := filepath.Join(packRoot, "skill-registry.md")
	return strings.Join([]string{
		"Read the exported Bitsentry capability pack:",
		"",
		fmt.Sprintf("- %s", usagePath),
		fmt.Sprintf("- %s", registryPath),
		"",
		"Use the Bitsentry SDD flow to prepare a change proposal for this repository.",
		"",
		"Constraints:",
		"- Do not modify code yet.",
		"- Do not modify opencode.json.",
		"- Do not execute runtime flows.",
		"- Return selected flow, selected skills, brief, goals, non-goals, handoff sequence, risks, and verdict.",
	}, "\n")
}

func detectComponentConfiguredStatus(m *model) (engramConfigured bool, context7Configured bool) {
	cfg, err := m.app.ConfigManager.Load()
	if err != nil {
		return false, false
	}
	return cfg.Components.Engram.Configured, cfg.Components.Context7.Configured
}

func detectComponentRuntimeStatus(m *model) (engramAvailable bool, context7Available bool) {
	cfg, err := m.app.ConfigManager.Load()
	if err != nil {
		return false, false
	}
	eng := components.DetectEngramRuntime(context.Background(), cfg)
	ctx7 := components.DetectContext7Runtime(context.Background(), cfg)
	return eng.BinaryFound, ctx7.Detected
}

type openCodeInstallStatus struct {
	Detected       bool
	BinaryPath     string
	ConfigRoot     string
	PackRoot       string
	PackInstalled  bool
	UsageExists    bool
	RegistryExists bool
}

func detectOpenCodeInstallStatus() openCodeInstallStatus {
	ctx := context.Background()
	res, _ := agents.OpenCodeDetector{}.Detect(ctx)
	wd, _ := os.Getwd()
	root := ""
	found := agents.ExistingOpenCodeConfigPaths(wd)
	if len(found) > 0 {
		root = found[0]
	} else {
		candidates := agents.OpenCodeConfigPathCandidates(wd)
		if len(candidates) > 0 {
			root = candidates[0]
		}
	}
	packRoot := ""
	usageOK := false
	registryOK := false
	installed := false
	if strings.TrimSpace(root) != "" {
		packRoot = filepath.Join(root, "bitsentry")
		usageOK = fileExistsPath(filepath.Join(packRoot, "OPENCODE_USAGE.md"))
		registryOK = fileExistsPath(filepath.Join(packRoot, "skill-registry.md"))
		installed = usageOK && registryOK
	}
	return openCodeInstallStatus{Detected: res.Found, BinaryPath: res.Path, ConfigRoot: root, PackRoot: packRoot, PackInstalled: installed, UsageExists: usageOK, RegistryExists: registryOK}
}

func fileExistsPath(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func openCodeStatusLine(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func componentStatusLabel(cfgEnabled bool, detected bool) string {
	if cfgEnabled {
		return "configured"
	}
	if detected {
		return "available"
	}
	return "not configured"
}

func boolLabel(v bool, yes string, no string) string {
	if v {
		return yes
	}
	return no
}

func selectedMCPList(set map[string]bool) []string {
	out := []string{}
	if set["engram"] {
		out = append(out, "engram")
	}
	if set["context7"] {
		out = append(out, "context7")
	}
	return out
}
