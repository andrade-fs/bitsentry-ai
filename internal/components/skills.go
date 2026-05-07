package components

import (
	"sort"
	"strings"

	"bitsentry-ai/internal/config"
)

type Skill struct {
	ID           string
	Name         string
	Description  string
	Category     string
	Enabled      bool
	Configured   bool
	Required     bool
	Status       Status
	TargetAgents []string
	Notes        string
}

type SkillsRegistrySummary struct {
	Enabled    bool
	Configured bool
	Selected   []string
	Total      int
}

func SkillsRegistry(cfg config.Config) ([]Skill, SkillsRegistrySummary) {
	skillsCfg := cfg.Components.Skills
	selected := make([]string, 0, len(skillsCfg.Selected))
	for _, v := range skillsCfg.Selected {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			selected = append(selected, trimmed)
		}
	}

	selectedSet := map[string]bool{}
	for _, id := range selected {
		selectedSet[id] = true
	}

	entries := []Skill{
		{
			ID:           "bitsentry-research-init",
			Name:         "BitSentry Research Init",
			Description:  "Strategic filter to evaluate raw content ideas before drafting.",
			Category:     "research",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no skill file installation or execution is performed.",
		},
		{
			ID:           "bitsentry-research-create",
			Name:         "BitSentry Research Create",
			Description:  "Turns approved ideas into draft-ready notes with the expected structure.",
			Category:     "research",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no skill file installation or execution is performed.",
		},
		{
			ID:           "bitsentry-research-validate",
			Name:         "BitSentry Research Validate",
			Description:  "Audits draft notes and decides publish readiness versus rewrite/archive.",
			Category:     "research",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no skill file installation or execution is performed.",
		},
		{
			ID:           "bitsentry-sdd",
			Name:         "BitSentry SDD",
			Description:  "Spec-Driven Development skill set for proposal/spec/design/tasks/apply/verify flows.",
			Category:     "development",
			Required:     true,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no workflow execution is performed in this batch.",
		},
		{
			ID:           "bitsentry-sdr",
			Name:         "BitSentry SDR",
			Description:  "Structured Discovery/Research skill family for ideas, notes and knowledge workflows.",
			Category:     "research",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Declarative core skill family. Runtime orchestration is intentionally out of scope in this phase.",
		},
		{
			ID:           "bitsentry-support",
			Name:         "BitSentry Support Skills",
			Description:  "Reusable helper skills shared by SDD, SDR and future orchestrator phases.",
			Category:     "platform",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Declarative helper set; includes skill-registry, judgment-day, go-testing and delivery helpers.",
		},
		{
			ID:           "bitsentry-oscp-notes",
			Name:         "BitSentry OSCP Notes",
			Description:  "Learning notes skill metadata for OSCP-oriented study and capture.",
			Category:     "learning",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no external note sync/config writes are performed.",
		},
		{
			ID:           "bitsentry-bugbounty-notes",
			Name:         "BitSentry BugBounty Notes",
			Description:  "Learning/security notes skill metadata for bug bounty research capture.",
			Category:     "security",
			Required:     false,
			Status:       StatusAvailable,
			TargetAgents: []string{"opencode"},
			Notes:        "Metadata only; no external note sync/config writes are performed.",
		},
		{
			ID:           "bitsentry-redteam-notes",
			Name:         "BitSentry RedTeam Notes",
			Description:  "Security notes skill metadata for red team knowledge capture workflows.",
			Category:     "security",
			Required:     false,
			Status:       StatusNotImplemented,
			TargetAgents: []string{"opencode"},
			Notes:        "Not implemented yet; intentionally excluded from default selected set.",
		},
	}

	for i := range entries {
		entries[i].Enabled = skillsCfg.Enabled && selectedSet[entries[i].ID]
		entries[i].Configured = skillsCfg.Configured && selectedSet[entries[i].ID]
	}

	sort.Strings(selected)
	summary := SkillsRegistrySummary{
		Enabled:    skillsCfg.Enabled,
		Configured: skillsCfg.Configured,
		Selected:   selected,
		Total:      len(entries),
	}

	return entries, summary
}
