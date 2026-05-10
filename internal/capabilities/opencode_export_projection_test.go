package capabilities

import (
	"strings"
	"testing"
)

func TestBuildOpenCodeExportProjection_SelectBitsentrySDD(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"bitsentry-sdd"})
	if err != nil {
		t.Fatalf("build projection: %v", err)
	}
	if !hasProjectedFlow(p.IncludedFlows, "sdd") {
		t.Fatalf("expected sdd flow included")
	}
	if hasProjectedFlow(p.IncludedFlows, "sdr") {
		t.Fatalf("did not expect sdr flow included")
	}
	if !hasProjectedPack(p.IncludedSkillPacks, "sdd") || !hasProjectedPack(p.IncludedSkillPacks, "_shared") {
		t.Fatalf("expected sdd and _shared packs included")
	}
	if !hasProjectedPack(p.IncludedSkillPacks, "support") {
		t.Fatalf("expected support pack included for support skill dependencies")
	}
	if len(p.GeneratedFiles) == 0 || p.GeneratedFiles[0].Path != "bitsentry/skill-registry.md" {
		t.Fatalf("expected generated skill registry")
	}
}

func TestBuildOpenCodeExportProjection_SelectBitsentrySDR(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"bitsentry-sdr"})
	if err != nil {
		t.Fatalf("build projection: %v", err)
	}
	if !hasProjectedFlow(p.IncludedFlows, "sdr") {
		t.Fatalf("expected sdr flow included")
	}
	if hasProjectedFlow(p.IncludedFlows, "sdd") {
		t.Fatalf("did not expect sdd flow included")
	}
	if !hasProjectedPack(p.IncludedSkillPacks, "sdr") || !hasProjectedPack(p.IncludedSkillPacks, "_shared") {
		t.Fatalf("expected sdr and _shared packs included")
	}
}

func TestBuildOpenCodeExportProjection_SelectBitsentrySupport(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"bitsentry-support"})
	if err != nil {
		t.Fatalf("build projection: %v", err)
	}
	if !hasProjectedFlow(p.IncludedFlows, "support") {
		t.Fatalf("expected support flow included")
	}
	if !hasProjectedPack(p.IncludedSkillPacks, "support") || !hasProjectedPack(p.IncludedSkillPacks, "_shared") {
		t.Fatalf("expected support and _shared packs included")
	}
}

func TestBuildOpenCodeExportProjection_DynamicFakeFlow(t *testing.T) {
	cat, err := DiscoverAssets("../../")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	// Inject dynamic fake entries without hardcoding known packs.
	cat.Flows = append(cat.Flows, DiscoveredFlow{ID: "fake-flow", Name: "Fake", SkillPack: "fake-pack", SourcePath: "/tmp/fake-flow.yaml", Requires: map[string]any{}, Stages: []map[string]any{{"skill": "fake-pack/fake-skill"}}})
	cat.SkillPacks = append(cat.SkillPacks, DiscoveredSkillPack{ID: "fake-pack", SourcePath: "/tmp/fake-pack"})
	cat.Skills = append(cat.Skills, DiscoveredSkill{ID: "fake-pack/fake-skill", PackID: "fake-pack", RelativePath: "skills/fake-pack/fake-skill/SKILL.md", SourcePath: "/tmp/fake-skill/SKILL.md", Status: "valid"})

	p, err := BuildOpenCodeExportProjection(cat, []string{"fake-flow"})
	if err != nil {
		t.Fatalf("build projection: %v", err)
	}
	if !hasProjectedFlow(p.IncludedFlows, "fake-flow") || !hasProjectedPack(p.IncludedSkillPacks, "fake-pack") {
		t.Fatalf("expected fake flow/pack included dynamically")
	}
}

func TestBuildOpenCodeExportProjection_UnknownSelectionWarning(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(catalog, []string{"unknown-x"})
	if err != nil {
		t.Fatalf("build projection: %v", err)
	}
	if len(p.Warnings) == 0 || len(p.Skipped) == 0 {
		t.Fatalf("expected warning and skipped for unknown selection")
	}
}

func TestValidateOpenCodeSelectionIDs_Strict(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	if err := ValidateOpenCodeSelectionIDs(catalog, []string{"bitsentry-sdd", "sdr"}); err != nil {
		t.Fatalf("expected valid selection ids, got: %v", err)
	}
	if err := ValidateOpenCodeSelectionIDs(catalog, []string{"bitsentry-unknown"}); err == nil {
		t.Fatalf("expected strict error for unknown alias")
	}
}

func TestDefaultPresetBitsentryDev_IsExportable(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	preset, ok := PresetByID("bitsentry-dev", DefaultPresets())
	if !ok {
		t.Fatalf("bitsentry-dev preset not found")
	}
	selected := append(append([]string{}, preset.Flows...), preset.Skills...)
	if err := ValidateOpenCodeSelectionIDs(catalog, selected); err != nil {
		t.Fatalf("bitsentry-dev must be exportable, got %v", err)
	}
}

func TestGenerateSkillRegistryContainsCoreSections(t *testing.T) {
	p := OpenCodeExportProjection{
		IncludedFlows:           []ProjectedFlow{{ID: "sdd", Name: "Spec Driven Development"}},
		IncludedSkillPacks:      []ProjectedSkillPack{{ID: "sdd"}},
		IncludedSkills:          []ProjectedSkill{{ID: "sdd/sdd-init"}},
		IncludedSharedContracts: []ProjectedSharedContract{{ID: "result-envelope"}},
	}
	content := GenerateSkillRegistry(p)
	checks := []string{"## Included Flows", "sdd", "## Loading Rules", "## Persistence Roots", "Result Envelope"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("generated registry missing %q", c)
		}
	}
}

func TestGenerateOpenCodeUsageContainsBoundaries(t *testing.T) {
	p := OpenCodeExportProjection{
		IncludedFlows:      []ProjectedFlow{{ID: "sdd", Name: "Spec Driven Development"}},
		IncludedSkillPacks: []ProjectedSkillPack{{ID: "sdd"}},
		IncludedSkills:     []ProjectedSkill{{ID: "sdd/sdd-init"}},
	}
	content := GenerateOpenCodeUsage(p)
	checks := []string{"Bitsentry capability pack export", "native runtime integration", "is not modified automatically", "sdd/sdd-init"}
	for _, c := range checks {
		if !strings.Contains(content, c) {
			t.Fatalf("usage content missing %q", c)
		}
	}
}

func hasProjectedFlow(flows []ProjectedFlow, id string) bool {
	for _, f := range flows {
		if f.ID == id {
			return true
		}
	}
	return false
}

func hasProjectedPack(packs []ProjectedSkillPack, id string) bool {
	for _, p := range packs {
		if p.ID == id {
			return true
		}
	}
	return false
}
