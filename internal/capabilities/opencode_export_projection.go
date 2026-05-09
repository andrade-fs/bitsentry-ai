package capabilities

import (
	"fmt"
	"sort"
	"strings"
)

type OpenCodeExportProjection struct {
	TargetAgent             string
	SelectedIDs             []string
	IncludedFlows           []ProjectedFlow
	IncludedSkillPacks      []ProjectedSkillPack
	IncludedSkills          []ProjectedSkill
	IncludedSharedContracts []ProjectedSharedContract
	IncludedOrchestrators   []ProjectedOrchestrator
	GeneratedFiles          []ProjectedGeneratedFile
	Warnings                []string
	Skipped                 []string
}

type ProjectedFlow struct {
	ID         string
	Name       string
	SourcePath string
	TargetPath string
}

type ProjectedSkillPack struct {
	ID          string
	SourcePath  string
	TargetRoot  string
	SkillCount  int
	RelatedFlow []string
}

type ProjectedSkill struct {
	ID         string
	PackID     string
	SourcePath string
	TargetPath string
	Status     string
}

type ProjectedSharedContract struct {
	ID         string
	SourcePath string
	TargetPath string
}

type ProjectedOrchestrator struct {
	ID         string
	SourcePath string
	TargetPath string
}

type ProjectedGeneratedFile struct {
	Path           string
	ContentSummary string
	Content        string
}

func BuildOpenCodeExportProjection(catalog AssetCatalog, selectedIDs []string) (OpenCodeExportProjection, error) {
	projection := OpenCodeExportProjection{TargetAgent: "opencode", SelectedIDs: normalizeIDs(selectedIDs)}
	if len(projection.SelectedIDs) == 0 {
		projection.Warnings = append(projection.Warnings, "No selections provided; projection is empty.")
		return projection, nil
	}

	flowByID := map[string]DiscoveredFlow{}
	for _, f := range catalog.Flows {
		flowByID[f.ID] = f
	}
	packByID := map[string]DiscoveredSkillPack{}
	for _, p := range catalog.SkillPacks {
		packByID[p.ID] = p
	}
	skillByID := map[string]DiscoveredSkill{}
	for _, s := range catalog.Skills {
		skillByID[s.ID] = s
	}

	includeFlows := map[string]bool{}
	includePacks := map[string]bool{}
	includeSkills := map[string]bool{}
	requireSupportPack := false

	for _, raw := range projection.SelectedIDs {
		id := resolveSelectionAlias(raw)
		_, flowExists := flowByID[id]
		_, packExists := packByID[id]
		switch {
		case flowExists:
			includeFlows[id] = true
			if pack := strings.TrimSpace(flowByID[id].SkillPack); pack != "" {
				includePacks[pack] = true
			}
		case packExists:
			includePacks[id] = true
			for _, f := range catalog.Flows {
				if f.SkillPack == id {
					includeFlows[f.ID] = true
				}
			}
		default:
			projection.Warnings = append(projection.Warnings, fmt.Sprintf("Selection %q is not a discovered flow or skill pack", raw))
			projection.Skipped = append(projection.Skipped, raw)
		}
	}

	for fid := range includeFlows {
		flow := flowByID[fid]
		for _, st := range flow.Stages {
			skillRef := strings.TrimSpace(asString(st["skill"]))
			if skillRef == "" {
				continue
			}
			if _, ok := skillByID[skillRef]; ok {
				includeSkills[skillRef] = true
				continue
			}
			projection.Warnings = append(projection.Warnings, fmt.Sprintf("Stage skill %q referenced by flow %q was not discovered", skillRef, fid))
		}
		for _, sref := range supportRefsFromFlow(flow) {
			if _, ok := skillByID[sref]; ok {
				includeSkills[sref] = true
				requireSupportPack = true
			} else {
				projection.Warnings = append(projection.Warnings, fmt.Sprintf("Support skill %q referenced by flow %q was not discovered", sref, fid))
			}
		}
	}

	if includePacks["support"] {
		requireSupportPack = true
	}
	if requireSupportPack {
		if _, ok := packByID["support"]; ok {
			includePacks["support"] = true
			projection.Warnings = append(projection.Warnings, "Support pack included because selected flow(s) require support skills.")
		}
	}

	if len(includeFlows)+len(includePacks) > 0 {
		includePacks["_shared"] = true
	}

	projection.IncludedFlows = projectFlows(catalog.Flows, includeFlows)
	projection.IncludedSkillPacks = projectSkillPacks(catalog.SkillPacks, includePacks)
	projection.IncludedSkills = projectSkills(catalog.Skills, includePacks, includeSkills)
	projection.IncludedSharedContracts = projectShared(catalog.Shared, includePacks["_shared"])
	projection.IncludedOrchestrators = projectOrchestrators(catalog.Orchestrators, len(projection.IncludedFlows) > 0)

	registry := GenerateSkillRegistry(projection)
	projection.GeneratedFiles = []ProjectedGeneratedFile{{
		Path:           "bitsentry/skill-registry.md",
		ContentSummary: fmt.Sprintf("Generated registry for %d flows, %d packs, %d skills", len(projection.IncludedFlows), len(projection.IncludedSkillPacks), len(projection.IncludedSkills)),
		Content:        registry,
	}}
	sort.Strings(projection.Warnings)
	sort.Strings(projection.Skipped)
	return projection, nil
}

func GenerateSkillRegistry(p OpenCodeExportProjection) string {
	b := &strings.Builder{}
	b.WriteString("# BitSentry Skill Registry\n\n")
	b.WriteString("## Included Flows\n")
	if len(p.IncludedFlows) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, f := range p.IncludedFlows {
			b.WriteString(fmt.Sprintf("- %s (%s)\n", f.ID, f.Name))
		}
	}
	b.WriteString("\n## Included Skill Packs\n")
	if len(p.IncludedSkillPacks) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, sp := range p.IncludedSkillPacks {
			b.WriteString(fmt.Sprintf("- %s\n", sp.ID))
		}
	}
	b.WriteString("\n## Included Skills\n")
	if len(p.IncludedSkills) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, s := range p.IncludedSkills {
			b.WriteString(fmt.Sprintf("- %s\n", s.ID))
		}
	}
	b.WriteString("\n## Shared Contracts\n")
	if len(p.IncludedSharedContracts) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, sc := range p.IncludedSharedContracts {
			b.WriteString(fmt.Sprintf("- %s\n", sc.ID))
		}
	}
	b.WriteString("\n## Handoffs\n")
	b.WriteString("- Handoffs are derived from included flow manifests and support skill references.\n")
	b.WriteString("\n## Persistence Roots\n")
	b.WriteString("- Persistence roots are defined in included flow manifests under `persistence`.\n")
	b.WriteString("\n## Loading Rules\n")
	b.WriteString("- Load `_shared` contracts first.\n")
	b.WriteString("- Load selected flow manifests and their skill packs.\n")
	b.WriteString("- Ensure `## Result Envelope` is present in each included skill contract.\n")
	return b.String()
}

func resolveSelectionAlias(id string) string {
	switch strings.TrimSpace(id) {
	case "bitsentry-sdd":
		return "sdd"
	case "bitsentry-sdr":
		return "sdr"
	case "bitsentry-support":
		return "support"
	default:
		return strings.TrimSpace(id)
	}
}

func supportRefsFromFlow(f DiscoveredFlow) []string {
	out := []string{}
	requires, ok := f.Requires["support_skills"]
	if !ok {
		return out
	}
	if arr, ok := requires.([]any); ok {
		for _, v := range arr {
			s := strings.TrimSpace(asString(v))
			if s != "" {
				out = append(out, s)
			}
		}
	}
	sort.Strings(out)
	return out
}

func projectFlows(flows []DiscoveredFlow, include map[string]bool) []ProjectedFlow {
	out := []ProjectedFlow{}
	for _, f := range flows {
		if !include[f.ID] {
			continue
		}
		out = append(out, ProjectedFlow{ID: f.ID, Name: f.Name, SourcePath: f.SourcePath, TargetPath: fmt.Sprintf("bitsentry/flows/%s.yaml", f.ID)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func projectSkillPacks(packs []DiscoveredSkillPack, include map[string]bool) []ProjectedSkillPack {
	out := []ProjectedSkillPack{}
	for _, p := range packs {
		if !include[p.ID] {
			continue
		}
		out = append(out, ProjectedSkillPack{ID: p.ID, SourcePath: p.SourcePath, TargetRoot: fmt.Sprintf("bitsentry/skills/%s", p.ID), SkillCount: p.SkillCount, RelatedFlow: append([]string{}, p.RelatedFlowIDs...)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func projectSkills(skills []DiscoveredSkill, includePacks map[string]bool, includeByRef map[string]bool) []ProjectedSkill {
	out := []ProjectedSkill{}
	for _, s := range skills {
		if !includePacks[s.PackID] && !includeByRef[s.ID] {
			continue
		}
		out = append(out, ProjectedSkill{ID: s.ID, PackID: s.PackID, SourcePath: s.SourcePath, TargetPath: fmt.Sprintf("bitsentry/%s", s.RelativePath), Status: s.Status})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func projectShared(shared []DiscoveredSharedContract, include bool) []ProjectedSharedContract {
	if !include {
		return []ProjectedSharedContract{}
	}
	out := make([]ProjectedSharedContract, 0, len(shared))
	for _, s := range shared {
		out = append(out, ProjectedSharedContract{ID: s.ID, SourcePath: s.Path, TargetPath: fmt.Sprintf("bitsentry/skills/_shared/%s.md", s.ID)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func projectOrchestrators(orchestrators []DiscoveredOrchestrator, include bool) []ProjectedOrchestrator {
	if !include {
		return []ProjectedOrchestrator{}
	}
	out := make([]ProjectedOrchestrator, 0, len(orchestrators))
	for _, o := range orchestrators {
		out = append(out, ProjectedOrchestrator{ID: o.ID, SourcePath: o.Path, TargetPath: fmt.Sprintf("bitsentry/orchestrators/%s.md", o.ID)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
