package capabilities

import (
	"fmt"
	"sort"
	"strings"

	"bitsentry-ai/internal/components"
)

type Plan struct {
	TargetAgent string   `yaml:"target_agent" json:"target_agent"`
	Preset      string   `yaml:"preset" json:"preset"`
	MCPs        []string `yaml:"mcps" json:"mcps"`
	Skills      []string `yaml:"skills" json:"skills"`
	Flows       []string `yaml:"flows" json:"flows"`
	Warnings    []string `yaml:"warnings" json:"warnings"`
	Actions     []string `yaml:"actions" json:"actions"`
}

func BuildPlan(catalog Catalog, targetAgent string, presetID string) (Plan, error) {
	targetAgent = strings.TrimSpace(targetAgent)
	if targetAgent == "" {
		targetAgent = "opencode"
	}
	if targetAgent != "opencode" {
		return Plan{}, fmt.Errorf("unsupported target agent %q in this phase; only opencode is supported", targetAgent)
	}

	selectedPreset := strings.TrimSpace(presetID)
	if selectedPreset == "" {
		selectedPreset = "custom"
	}

	preset, found := PresetByID(selectedPreset, catalog.Presets)
	if !found {
		return Plan{}, fmt.Errorf("unknown preset %q", selectedPreset)
	}

	plan := basePlan(targetAgent, preset.ID)
	plan.MCPs = append([]string{}, preset.MCPs...)
	plan.Skills = append([]string{}, preset.Skills...)
	plan.Flows = append([]string{}, preset.Flows...)

	return finalizePlan(plan, catalog), nil
}

func BuildPlanFromSelection(catalog Catalog, targetAgent string, presetID string, mcps []string, skills []string, flows []string) (Plan, error) {
	targetAgent = strings.TrimSpace(targetAgent)
	if targetAgent == "" {
		targetAgent = "opencode"
	}
	if targetAgent != "opencode" {
		return Plan{}, fmt.Errorf("unsupported target agent %q in this phase; only opencode is supported", targetAgent)
	}

	selectedPreset := strings.TrimSpace(presetID)
	if selectedPreset == "" {
		selectedPreset = "custom"
	}

	plan := basePlan(targetAgent, selectedPreset)
	plan.MCPs = normalizeIDs(mcps)
	plan.Skills = normalizeIDs(skills)
	plan.Flows = normalizeIDs(flows)

	return finalizePlan(plan, catalog), nil
}

func basePlan(targetAgent string, preset string) Plan {
	return Plan{
		TargetAgent: targetAgent,
		Preset:      preset,
		Actions: []string{
			"Review the capability plan",
			"Export/patch OpenCode config through adapter",
			"Apply only managed capabilities with backup",
		},
	}
}

func finalizePlan(plan Plan, catalog Catalog) Plan {
	knownMCP := makeSet(catalog.MCPs)
	knownSkills := makeSet(catalog.Skills)
	knownFlows := makeSet(catalog.Flows)

	plan.Warnings = append(plan.Warnings, warnUnknown("MCP", plan.MCPs, knownMCP)...)
	plan.Warnings = append(plan.Warnings, warnUnknown("Skill", plan.Skills, knownSkills)...)
	plan.Warnings = append(plan.Warnings, warnUnknown("Flow", plan.Flows, knownFlows)...)

	if contains(plan.MCPs, "postgres") {
		plan.Warnings = append(plan.Warnings, "postgres MCP is declared in presets but OpenCode adapter write support is not implemented yet.")
	}
	if contains(plan.Flows, "redteam") {
		plan.Warnings = append(plan.Warnings, "redteam flow is declarative in this phase; runtime execution is not implemented.")
	}

	sort.Strings(plan.MCPs)
	sort.Strings(plan.Skills)
	sort.Strings(plan.Flows)
	return plan
}

type OpenCodeProjection struct {
	ManagedMCPs       []string `yaml:"managed_mcps" json:"managed_mcps"`
	SkippedMCPs       []string `yaml:"skipped_mcps" json:"skipped_mcps"`
	DeclarativeFlows  []string `yaml:"declarative_flows" json:"declarative_flows"`
	DeclarativeSkills []string `yaml:"declarative_skills" json:"declarative_skills"`
	Warnings          []string `yaml:"warnings" json:"warnings"`
}

func ProjectOpenCode(plan Plan) OpenCodeProjection {
	managedSet := map[string]bool{"engram": true, "context7": true}

	managed := make([]string, 0)
	skipped := make([]string, 0)
	for _, id := range plan.MCPs {
		if managedSet[id] {
			managed = append(managed, id)
		} else {
			skipped = append(skipped, id)
		}
	}

	warnings := append([]string{}, plan.Warnings...)
	if len(skipped) > 0 {
		warnings = append(warnings, "Some MCPs are not managed by OpenCode adapter yet and will be skipped on apply.")
	}

	sort.Strings(managed)
	sort.Strings(skipped)

	return OpenCodeProjection{
		ManagedMCPs:       managed,
		SkippedMCPs:       skipped,
		DeclarativeFlows:  append([]string{}, plan.Flows...),
		DeclarativeSkills: append([]string{}, plan.Skills...),
		Warnings:          warnings,
	}
}

func StatusFromComponent(status components.Status) string {
	return string(status)
}

func makeSet(entries []Entry) map[string]bool {
	set := map[string]bool{}
	for _, e := range entries {
		set[e.ID] = true
	}
	return set
}

func warnUnknown(kind string, selected []string, known map[string]bool) []string {
	w := make([]string, 0)
	for _, id := range selected {
		if !known[id] {
			w = append(w, fmt.Sprintf("%s %q is not modeled in current registry", kind, id))
		}
	}
	return w
}

func contains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func normalizeIDs(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		id := strings.TrimSpace(v)
		if id == "" || set[id] {
			continue
		}
		set[id] = true
		out = append(out, id)
	}
	return out
}
