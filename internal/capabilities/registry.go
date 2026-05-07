package capabilities

import (
	"sort"
	"strings"

	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
)

type Kind string

const (
	KindMCP    Kind = "mcp"
	KindSkill  Kind = "skill"
	KindFlow   Kind = "flow"
	KindPreset Kind = "preset"
	KindTarget Kind = "target"
)

type Entry struct {
	ID          string
	Name        string
	Kind        Kind
	Category    string
	Description string
	Status      components.Status
	Selected    bool
	Enabled     bool
	Configured  bool
	Targets     []string
	Notes       string
}

type Preset struct {
	ID      string
	Name    string
	MCPs    []string
	Skills  []string
	Flows   []string
	Targets []string
}

type Catalog struct {
	MCPs    []Entry
	Skills  []Entry
	Flows   []Entry
	Targets []Entry
	Presets []Preset
}

func BuildCatalog(cfg config.Config, mcpEntries []components.MCP, skillEntries []components.Skill) Catalog {
	mcps := make([]Entry, 0, len(mcpEntries))
	for _, m := range mcpEntries {
		mcps = append(mcps, Entry{
			ID:          m.ID,
			Name:        m.Name,
			Kind:        KindMCP,
			Category:    m.Category,
			Description: m.Description,
			Status:      m.Status,
			Selected:    m.Enabled,
			Enabled:     m.Enabled,
			Configured:  m.Configured,
			Targets:     []string{"opencode"},
			Notes:       m.Notes,
		})
	}

	skills := make([]Entry, 0, len(skillEntries))
	for _, s := range skillEntries {
		skills = append(skills, Entry{
			ID:          s.ID,
			Name:        s.Name,
			Kind:        KindSkill,
			Category:    s.Category,
			Description: s.Description,
			Status:      s.Status,
			Selected:    s.Enabled,
			Enabled:     s.Enabled,
			Configured:  s.Configured,
			Targets:     append([]string{}, s.TargetAgents...),
			Notes:       s.Notes,
		})
	}

	flowSelected := map[string]bool{}
	for _, id := range cfg.Components.Flows.Selected {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			flowSelected[trimmed] = true
		}
	}
	flows := []Entry{
		{ID: "sdd", Name: "SDD", Kind: KindFlow, Category: "development", Description: "Spec-Driven Development flow.", Status: components.StatusAvailable, Selected: flowSelected["sdd"], Enabled: cfg.Components.Flows.Enabled && flowSelected["sdd"], Configured: cfg.Components.Flows.Configured && flowSelected["sdd"], Targets: []string{"opencode"}, Notes: "Planned router/orchestrator execution path."},
		{ID: "sdr", Name: "SDR", Kind: KindFlow, Category: "research", Description: "Spec-Driven Research flow.", Status: components.StatusAvailable, Selected: flowSelected["sdr"], Enabled: cfg.Components.Flows.Enabled && flowSelected["sdr"], Configured: cfg.Components.Flows.Configured && flowSelected["sdr"], Targets: []string{"opencode"}, Notes: "Declarative flow selection for future router."},
		{ID: "redteam", Name: "RedTeam", Kind: KindFlow, Category: "security", Description: "Red Team workflow.", Status: components.StatusNotImplemented, Selected: flowSelected["redteam"], Enabled: cfg.Components.Flows.Enabled && flowSelected["redteam"], Configured: cfg.Components.Flows.Configured && flowSelected["redteam"], Targets: []string{"opencode"}, Notes: "Execution layer not implemented yet."},
		{ID: "notes", Name: "Notes", Kind: KindFlow, Category: "knowledge", Description: "Note capture and synthesis workflow.", Status: components.StatusAvailable, Selected: flowSelected["notes"], Enabled: cfg.Components.Flows.Enabled && flowSelected["notes"], Configured: cfg.Components.Flows.Configured && flowSelected["notes"], Targets: []string{"opencode"}, Notes: "Can be chained from SDR/RedTeam in future router."},
	}

	targetSelected := map[string]bool{}
	for _, id := range cfg.Components.Targets.Selected {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			targetSelected[trimmed] = true
		}
	}
	targets := []Entry{
		{ID: "opencode", Name: "OpenCode", Kind: KindTarget, Category: "agent", Description: "Primary supported target in this phase.", Status: components.StatusAvailable, Selected: targetSelected["opencode"], Enabled: targetSelected["opencode"], Configured: targetSelected["opencode"], Notes: "First real adapter target."},
	}

	return Catalog{
		MCPs:    mcps,
		Skills:  skills,
		Flows:   flows,
		Targets: targets,
		Presets: DefaultPresets(),
	}
}

func DefaultPresets() []Preset {
	return []Preset{
		{ID: "bitsentry-full", Name: "Bitsentry Full", MCPs: []string{"engram", "context7", "postgres"}, Skills: []string{"bitsentry-sdd", "bitsentry-research-init", "bitsentry-research-create", "bitsentry-research-validate", "bitsentry-oscp-notes", "bitsentry-bugbounty-notes", "bitsentry-redteam-notes"}, Flows: []string{"sdd", "sdr", "redteam", "notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-dev", Name: "Bitsentry Dev", MCPs: []string{"engram", "context7"}, Skills: []string{"bitsentry-sdd"}, Flows: []string{"sdd", "notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-research", Name: "Bitsentry Research", MCPs: []string{"engram", "context7"}, Skills: []string{"bitsentry-research-init", "bitsentry-research-create", "bitsentry-research-validate"}, Flows: []string{"sdr", "notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-oscp", Name: "Bitsentry OSCP", MCPs: []string{"engram", "context7"}, Skills: []string{"bitsentry-oscp-notes"}, Flows: []string{"notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-bugbounty", Name: "Bitsentry BugBounty", MCPs: []string{"engram", "context7", "postgres"}, Skills: []string{"bitsentry-bugbounty-notes", "bitsentry-research-init"}, Flows: []string{"redteam", "notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-redteam", Name: "Bitsentry RedTeam", MCPs: []string{"engram", "context7", "postgres"}, Skills: []string{"bitsentry-redteam-notes"}, Flows: []string{"redteam", "notes"}, Targets: []string{"opencode"}},
		{ID: "bitsentry-blog", Name: "Bitsentry Blog", MCPs: []string{"engram", "context7"}, Skills: []string{"bitsentry-research-init", "bitsentry-research-create", "bitsentry-research-validate"}, Flows: []string{"sdr", "notes"}, Targets: []string{"opencode"}},
		{ID: "custom", Name: "Custom", MCPs: []string{}, Skills: []string{}, Flows: []string{}, Targets: []string{"opencode"}},
	}
}

func PresetByID(id string, presets []Preset) (Preset, bool) {
	for _, p := range presets {
		if p.ID == id {
			return p, true
		}
	}
	return Preset{}, false
}

func SortedIDs(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.ID)
	}
	sort.Strings(out)
	return out
}
