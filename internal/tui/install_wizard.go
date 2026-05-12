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
	installStepMode
	installStepReview
	installStepInstall
	installStepDone
	installStepCount = 6
)

const (
	installModeEverything = iota
	installModePackOnly
	installModeUpdateReinstall
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

	TargetSelected      bool
	InstallMode         int
	SelectedMCPs        map[string]bool
	MCPReadiness        map[string]components.MCPReadiness
	RegisterAgent       bool
	InstallCommands     bool
	InstallNativeSkills bool
	ConfigureMCP        bool
	MCPConfigPreview    capabilities.OpenCodeMCPConfigPreview
	NativeIntegrationOK bool
	ConfigBackupPath    string
	NativeBackupPath    string

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
		InstallMode:       defaultInstallMode(status.PackInstalled),
		SelectedMCPs: map[string]bool{
			"engram":   engramSelected,
			"context7": context7Selected,
		},
		MCPReadiness: map[string]components.MCPReadiness{
			"engram":   deriveMCPReadiness("engram", engramCfg, engramRuntime),
			"context7": deriveMCPReadiness("context7", context7Cfg, context7Runtime),
		},
		RegisterAgent:       true,
		InstallCommands:     true,
		InstallNativeSkills: true,
		ConfigureMCP:        false,
		ResultNotes:         []string{},
	}
	m.install.MCPConfigPreview = capabilities.BuildOpenCodeMCPConfigPreview(m.install.OpenCodeConfig, selectedMCPList(m.install.SelectedMCPs))
}

func (m *model) refreshInstallWizardPreserveSelections() {
	prev := m.install
	status := detectOpenCodeInstallStatus()
	engramCfg, context7Cfg := detectComponentConfiguredStatus(m)
	engramRuntime, context7Runtime := detectComponentRuntimeStatus(m)

	targetSelected := status.Detected
	targetSelected = prev.TargetSelected
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

	installMode := prev.InstallMode
	if installMode < installModeEverything || installMode > installModeUpdateReinstall {
		installMode = defaultInstallMode(status.PackInstalled)
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
		InstallMode:       installMode,
		SelectedMCPs:      selectedMCPs,
		MCPReadiness: map[string]components.MCPReadiness{
			"engram":   deriveMCPReadiness("engram", engramCfg, engramRuntime),
			"context7": deriveMCPReadiness("context7", context7Cfg, context7Runtime),
		},
		RegisterAgent:       prev.RegisterAgent,
		InstallCommands:     prev.InstallCommands,
		InstallNativeSkills: prev.InstallNativeSkills,
		ConfigureMCP:        prev.ConfigureMCP,
		NativeIntegrationOK: prev.NativeIntegrationOK,
		ConfigBackupPath:    prev.ConfigBackupPath,
		NativeBackupPath:    prev.NativeBackupPath,
		ResultStatus:        prev.ResultStatus,
		ResultNotes:         append([]string{}, prev.ResultNotes...),
		NextPrompt:          prev.NextPrompt,
	}
	m.install.MCPConfigPreview = capabilities.BuildOpenCodeMCPConfigPreview(m.install.OpenCodeConfig, selectedMCPList(m.install.SelectedMCPs))
}

func (m *model) handleInstallKey(key string) bool {
	w := &m.install
	switch key {
	case "up", "k":
		if (w.CurrentStep == installStepMode || w.CurrentStep == installStepTarget) && w.Cursor > 0 {
			w.Cursor--
		}
		return true
	case "down", "j":
		if (w.CurrentStep == installStepMode || w.CurrentStep == installStepTarget) && w.Cursor < 2 {
			w.Cursor++
		}
		return true
	case "space", " ":
		switch w.CurrentStep {
		case installStepTarget:
			if w.OpenCodeDetected {
				w.TargetSelected = !w.TargetSelected
			}
		case installStepMode:
			w.InstallMode = w.Cursor
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
		w.CurrentStep = installStepMode
		m.errMsg = ""
		return true
	case installStepMode:
		w.InstallMode = w.Cursor
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

	presetID := presetForMode(m.install.InstallMode)
	preset, ok := capabilities.PresetByID(presetID, capabilities.DefaultPresets())
	if !ok {
		m.errMsg = fmt.Sprintf("Install blocked: preset %q not found.", presetID)
		m.install.ResultStatus = "FAIL"
		return
	}

	selectedMCPs := selectedMCPList(m.install.SelectedMCPs)
	registerAgent, installCommands, installNativeSkills := nativeOptionsForMode(m.install.InstallMode)
	m.install.RegisterAgent = registerAgent
	m.install.InstallCommands = installCommands
	m.install.InstallNativeSkills = installNativeSkills
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
		RegisterAgent:       registerAgent,
		InstallCommands:     installCommands,
		InstallNativeSkills: installNativeSkills,
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
	m.install.NativeIntegrationOK = true
	m.install.ConfigBackupPath = nativeRes.ConfigBackupPath
	m.install.NativeBackupPath = nativeRes.NativeBackupPath
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

func defaultInstallMode(packInstalled bool) int {
	if packInstalled {
		return installModeUpdateReinstall
	}
	return installModeEverything
}

func presetForMode(mode int) string {
	if mode == installModePackOnly {
		return "bitsentry-dev"
	}
	return "bitsentry-full"
}

func nativeOptionsForMode(mode int) (registerAgent bool, installCommands bool, installNativeSkills bool) {
	if mode == installModePackOnly {
		return false, false, false
	}
	return true, true, true
}

func installModeLabel(mode int) string {
	switch mode {
	case installModeEverything:
		return "Install Everything"
	case installModePackOnly:
		return "Install Bitsentry Pack"
	case installModeUpdateReinstall:
		return "Update/Reinstall Bitsentry Pack"
	default:
		return "Install Everything"
	}
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

func deriveMCPReadiness(id string, cfgEnabled bool, detected bool) components.MCPReadiness {
	if cfgEnabled {
		return components.BuildMCPReadiness(components.StatusConfigured, []string{"bitsentry metadata configured"}, nil, nil, true)
	}
	if detected {
		return components.BuildMCPReadiness(components.StatusDetected, []string{"runtime evidence detected"}, []string{"metadata not configured"}, []string{"manual metadata step needed"}, false)
	}
	if id == "context7" {
		return components.BuildMCPReadiness(components.StatusMissing, nil, []string{"runtime command not detected"}, []string{"manual install + metadata configuration needed"}, false)
	}
	return components.BuildMCPReadiness(components.StatusMissing, nil, []string{"runtime evidence not detected"}, []string{"manual setup needed if this MCP is required"}, false)
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
