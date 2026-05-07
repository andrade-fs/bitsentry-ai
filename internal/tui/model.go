package tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/app"
	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/config"
	"bitsentry-ai/internal/profiles"
	"bitsentry-ai/internal/system"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenHome screen = iota
	screenInstall
	screenSystem
	screenAgents
	screenComponents
	screenCapabilities
	screenProfiles
	screenWorkflows
	screenSettings
	screenExit
)

type model struct {
	app      *app.App
	dryRun   bool
	menu     []menuItem
	selected int
	screen   screen
	styles   styles
	message  string
	errMsg   string
	profiles []profiles.Profile
	active   string
	system   system.SystemInfo
	shell    string
	deps     []system.DependencyStatus
	agents   []agents.AgentDetectionResult
	width    int
	height   int
	capSvc   *capabilities.Service

	capPreset       string
	capTarget       string
	capMCPs         map[string]bool
	capFlows        map[string]bool
	capSkills       map[string]bool
	capPresetOrder  []string
	capAwaitConfirm bool
	capLastPlan     string
	capLastValidate string
	capDirty        bool
}

func newModel(a *app.App, dryRun bool) model {
	m := model{
		app:    a,
		dryRun: dryRun,
		menu:   defaultMenu(),
		screen: screenHome,
		styles: newStyles(),
	}
	exe, _ := os.Executable()
	m.capSvc = capabilities.NewService(a, exe)
	m.refreshData()
	return m
}

func (m *model) refreshData() {
	cfg, err := m.app.ConfigManager.Load()
	if err != nil {
		m.errMsg = fmt.Sprintf("Could not load config: %v", err)
		m.active = "unknown"
	} else {
		m.active = cfg.ActiveProfile
		m.errMsg = ""
	}
	m.system = system.DetectSystem()
	m.shell = system.DetectShell()
	m.deps = system.CheckDependencies()
	m.profiles = m.app.ProfileStore.List(m.active)
	results, err := m.app.AgentRegistry.List(context.Background())
	if err != nil {
		m.errMsg = fmt.Sprintf("Could not detect agents: %v", err)
		m.agents = nil
		return
	}
	m.agents = results
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if (m.screen == screenHome || m.screen == screenProfiles) && m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			limit := len(m.menu) - 1
			if m.screen == screenProfiles {
				limit = len(m.profiles) - 1
			}
			if (m.screen == screenHome || m.screen == screenProfiles) && m.selected < limit {
				m.selected++
			}
		case "enter":
			if m.screen == screenHome {
				next := m.menu[m.selected].Screen
				if next == screenExit {
					return m, tea.Quit
				}
				m.screen = next
				m.selected = 0
				m.message = ""
				m.refreshData()
				if m.screen == screenCapabilities {
					m.loadCapabilityState()
				}
				return m, nil
			}
			if m.screen == screenProfiles {
				if len(m.profiles) == 0 {
					return m, nil
				}
				chosen := m.profiles[m.selected].ID
				if m.dryRun {
					m.message = fmt.Sprintf("[dry-run] Profile %q would be activated.", chosen)
					return m, nil
				}
				if err := m.app.SetActiveProfile(chosen); err != nil {
					m.errMsg = fmt.Sprintf("Could not activate profile %q: %v", chosen, err)
					return m, nil
				}
				m.message = fmt.Sprintf("Active profile set to %q", chosen)
				m.refreshData()
			}
		case "esc", "backspace":
			if m.screen == screenCapabilities && m.capAwaitConfirm {
				m.capAwaitConfirm = false
				m.message = "Apply confirmation cancelled."
				return m, nil
			}
			if m.screen != screenHome {
				m.screen = screenHome
				m.selected = 0
				m.refreshData()
			}
		case "[":
			if m.screen == screenCapabilities {
				m.shiftCapabilityPreset(-1)
			}
		case "]":
			if m.screen == screenCapabilities {
				m.shiftCapabilityPreset(1)
			}
		case "m":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capMCPs, "engram")
			}
		case "c":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capMCPs, "context7")
			}
		case "p":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capMCPs, "postgres")
			}
		case "1":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capFlows, "sdd")
			}
		case "2":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capFlows, "sdr")
			}
		case "3":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capFlows, "notes")
			}
		case "4":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capFlows, "redteam")
			}
		case "z":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capSkills, "bitsentry-sdd")
			}
		case "x":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capSkills, "bitsentry-research-init")
			}
		case "n":
			if m.screen == screenCapabilities {
				m.toggleCapability(m.capSkills, "bitsentry-bugbounty-notes")
			}
		case "s":
			if m.screen == screenCapabilities {
				m.saveCapabilityConfig()
			}
		case "v":
			if m.screen == screenCapabilities {
				m.validateCapabilityDraft()
			}
		case "l":
			if m.screen == screenCapabilities {
				m.previewCapabilityPlan()
			}
		case "d":
			if m.screen == screenCapabilities {
				m.applyCapabilities(true)
			}
		case "a":
			if m.screen == screenCapabilities {
				if m.capAwaitConfirm {
					return m, nil
				}
				if !m.validateCapabilityDraft() {
					m.errMsg = "Validation failed. Real apply blocked."
					return m, nil
				}
				m.capAwaitConfirm = true
				m.message = "Confirm real apply: press y to continue, esc/backspace to cancel."
			}
		case "y":
			if m.screen == screenCapabilities && m.capAwaitConfirm {
				m.capAwaitConfirm = false
				m.applyCapabilities(false)
			}
		}
	}

	if m.screen == screenProfiles {
		if m.selected >= len(m.profiles) {
			m.selected = max(len(m.profiles)-1, 0)
		}
	}

	return m, nil
}

func (m *model) loadCapabilityState() {
	if m.capMCPs == nil {
		m.capMCPs = map[string]bool{}
	}
	if m.capFlows == nil {
		m.capFlows = map[string]bool{}
	}
	if m.capSkills == nil {
		m.capSkills = map[string]bool{}
	}
	m.capAwaitConfirm = false
	m.capLastPlan = ""
	m.capLastValidate = ""
	m.capDirty = false

	draft, err := m.capSvc.LoadSelection()
	if err != nil {
		m.errMsg = fmt.Sprintf("Could not load capability config: %v", err)
		return
	}

	m.capPresetOrder = make([]string, 0)
	for _, p := range capabilities.DefaultPresets() {
		m.capPresetOrder = append(m.capPresetOrder, p.ID)
	}
	m.capPreset = draft.Preset
	m.capTarget = draft.TargetAgent

	for k := range m.capMCPs {
		delete(m.capMCPs, k)
	}
	for k := range m.capFlows {
		delete(m.capFlows, k)
	}
	for k := range m.capSkills {
		delete(m.capSkills, k)
	}
	for _, id := range draft.MCPs {
		m.capMCPs[id] = true
	}
	for _, id := range draft.Flows {
		m.capFlows[id] = true
	}
	for _, id := range draft.Skills {
		m.capSkills[id] = true
	}
}

func (m *model) shiftCapabilityPreset(delta int) {
	if len(m.capPresetOrder) == 0 {
		return
	}
	idx := 0
	for i, id := range m.capPresetOrder {
		if id == m.capPreset {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(m.capPresetOrder)) % len(m.capPresetOrder)
	m.capPreset = m.capPresetOrder[idx]
	if m.capPreset == "custom" {
		m.message = "Preset set to custom. Use toggles and save (s)."
		m.capDirty = true
		return
	}
	if preset, ok := capabilities.PresetByID(m.capPreset, capabilities.DefaultPresets()); ok {
		m.capTarget = "opencode"
		m.capMCPs = map[string]bool{}
		m.capFlows = map[string]bool{}
		m.capSkills = map[string]bool{}
		for _, id := range preset.MCPs {
			m.capMCPs[id] = true
		}
		for _, id := range preset.Flows {
			m.capFlows[id] = true
		}
		for _, id := range preset.Skills {
			m.capSkills[id] = true
		}
		m.message = fmt.Sprintf("Preset %q loaded into draft. Press s to save.", m.capPreset)
		m.capDirty = true
	}
}

func (m *model) toggleCapability(set map[string]bool, id string) {
	set[id] = !set[id]
	m.capPreset = "custom"
	m.message = fmt.Sprintf("Toggled %s. Press s to save draft.", id)
	m.capDirty = true
}

func (m *model) saveCapabilityConfig() {
	draft := capabilities.SelectionDraft{
		TargetAgent: m.capTarget,
		Preset:      m.capPreset,
		MCPs:        sortedSelected(m.capMCPs),
		Skills:      sortedSelected(m.capSkills),
		Flows:       sortedSelected(m.capFlows),
	}

	if err := m.capSvc.SaveSelection(draft); err != nil {
		m.errMsg = fmt.Sprintf("Could not save capability config: %v", err)
		return
	}
	m.message = "Capability draft saved to config."
	m.capDirty = false
	m.refreshData()
}

func (m *model) validateCapabilityDraft() bool {
	draft := capabilities.SelectionDraft{TargetAgent: m.capTarget, Preset: m.capPreset, MCPs: sortedSelected(m.capMCPs), Skills: sortedSelected(m.capSkills), Flows: sortedSelected(m.capFlows)}
	res, err := m.capSvc.ValidateSelection(draft)
	if err != nil {
		m.capLastValidate = "INVALID: " + err.Error()
		m.errMsg = m.capLastValidate
		return false
	}
	if !res.Valid {
		m.capLastValidate = "INVALID: " + strings.Join(res.Issues, "; ")
		m.errMsg = m.capLastValidate
		return false
	}
	m.capLastValidate = "VALID"
	m.message = "Capability draft validation passed."
	return true
}

func (m *model) previewCapabilityPlan() {
	if !m.validateCapabilityDraft() {
		return
	}
	draft := capabilities.SelectionDraft{TargetAgent: m.capTarget, Preset: m.capPreset, MCPs: sortedSelected(m.capMCPs), Skills: sortedSelected(m.capSkills), Flows: sortedSelected(m.capFlows)}
	planRes, err := m.capSvc.BuildPlan(draft)
	if err != nil {
		m.errMsg = fmt.Sprintf("Plan preview failed: %v", err)
		return
	}
	projection := planRes.Projection
	m.capLastPlan = fmt.Sprintf("target=%s preset=%s managed=[%s] skipped=[%s] declarative flows=[%s] skills=[%s]",
		planRes.Plan.TargetAgent,
		planRes.Plan.Preset,
		strings.Join(projection.ManagedMCPs, ", "),
		strings.Join(projection.SkippedMCPs, ", "),
		strings.Join(projection.DeclarativeFlows, ", "),
		strings.Join(projection.DeclarativeSkills, ", "),
	)
	m.message = "Plan preview generated (read-only)."
}

func (m *model) applyCapabilities(forceDryRun bool) {
	if !m.validateCapabilityDraft() {
		m.errMsg = "Validation failed. Apply blocked."
		return
	}
	if m.capDirty {
		m.errMsg = "Draft has unsaved changes. Press 's' before apply/dry-run."
		return
	}
	if m.dryRun && !forceDryRun {
		forceDryRun = true
	}
	if forceDryRun {
		m.message = "Running capability apply in dry-run mode..."
	} else {
		m.message = "Running capability apply..."
	}

	draft := capabilities.SelectionDraft{TargetAgent: m.capTarget, Preset: m.capPreset, MCPs: sortedSelected(m.capMCPs), Skills: sortedSelected(m.capSkills), Flows: sortedSelected(m.capFlows)}
	var res capabilities.ApplyResult
	var err error
	if forceDryRun {
		res, err = m.capSvc.ApplyDryRun(draft)
	} else {
		res, err = m.capSvc.Apply(draft)
	}
	if err != nil {
		m.errMsg = fmt.Sprintf("Apply command failed: %v", err)
		return
	}
	m.message = strings.TrimSpace(res.Output)
	if strings.TrimSpace(res.LatestReportPath) != "" {
		m.capLastPlan = fmt.Sprintf("latest report: %s", res.LatestReportPath)
	}
}
func sortedSelected(set map[string]bool) []string {
	out := make([]string, 0)
	for id, enabled := range set {
		if enabled {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func (m model) View() string {
	return renderScreen(m)
}

func packageManager(deps []system.DependencyStatus) string {
	for _, d := range deps {
		if !d.Found {
			continue
		}
		switch d.Name {
		case "brew", "apt", "yum", "pacman":
			return fmt.Sprintf("%s (%s)", d.Name, d.Path)
		}
	}
	return "not detected"
}

func dependencySummary(deps []system.DependencyStatus) string {
	parts := make([]string, 0, len(deps))
	for _, d := range deps {
		status := "missing"
		if d.Found {
			status = "ok"
		}
		parts = append(parts, fmt.Sprintf("%s:%s", d.Name, status))
	}
	return strings.Join(parts, ", ")
}

func configPaths() (string, string) {
	path := config.ConfigPath()
	return config.ConfigDir(), path
}
