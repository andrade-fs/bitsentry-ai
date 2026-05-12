package capabilities

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type flowManifest struct {
	ID                string            `yaml:"id"`
	Kind              string            `yaml:"kind"`
	Selectable        bool              `yaml:"selectable"`
	TopLevelFlow      bool              `yaml:"top_level_flow"`
	SkillPack         string            `yaml:"skill_pack"`
	OrchestratorSkill string            `yaml:"orchestrator_skill"`
	Triggers          []string          `yaml:"triggers"`
	Contracts         []string          `yaml:"contracts"`
	Persistence       map[string]string `yaml:"persistence"`
	StageGraph        map[string]any    `yaml:"stage_graph"`
	Stages            []struct {
		Skill string `yaml:"skill"`
	} `yaml:"stages"`
	Requires struct {
		MCPs    []string `yaml:"mcps"`
		Targets []string `yaml:"targets"`
	} `yaml:"requires"`
}

func TestFlowManifestsDynamicFields(t *testing.T) {
	paths := []string{
		"../../assets/flows/sdd.yaml",
		"../../assets/flows/sdr.yaml",
		"../../assets/flows/support.yaml",
		"../../assets/flows/source-security-review.yaml",
	}
	for _, p := range paths {
		m := readFlowManifest(t, p)
		if strings.TrimSpace(m.Kind) == "" {
			t.Fatalf("manifest %s missing kind", p)
		}
		if len(m.SkillPack) == 0 {
			t.Fatalf("manifest %s missing skill_pack", p)
		}
		if strings.TrimSpace(m.OrchestratorSkill) == "" {
			t.Fatalf("manifest %s missing orchestrator_skill", p)
		}
		if len(m.Triggers) == 0 {
			t.Fatalf("manifest %s missing triggers", p)
		}
		if len(m.Contracts) == 0 {
			t.Fatalf("manifest %s missing contracts", p)
		}
		if len(m.Persistence) == 0 {
			t.Fatalf("manifest %s missing persistence", p)
		}
		if len(m.StageGraph) == 0 {
			t.Fatalf("manifest %s missing stage_graph", p)
		}
	}
}

func TestFlowManifestSpecificContracts(t *testing.T) {
	sdd := readFlowManifest(t, "../../assets/flows/sdd.yaml")
	if sdd.Kind != "flow" || sdd.SkillPack != "sdd" {
		t.Fatalf("unexpected sdd kind/pack: %s/%s", sdd.Kind, sdd.SkillPack)
	}
	for _, mcp := range sdd.Requires.MCPs {
		if mcp == "opencode" {
			t.Fatalf("sdd requires.mcps must not include opencode")
		}
	}
	if len(sdd.Requires.Targets) == 0 || sdd.Requires.Targets[0] != "opencode" {
		t.Fatalf("sdd requires.targets must include opencode")
	}

	sdr := readFlowManifest(t, "../../assets/flows/sdr.yaml")
	if sdr.Kind != "flow" || sdr.SkillPack != "sdr" {
		t.Fatalf("unexpected sdr kind/pack: %s/%s", sdr.Kind, sdr.SkillPack)
	}

	support := readFlowManifest(t, "../../assets/flows/support.yaml")
	if support.Kind != "support" {
		t.Fatalf("support manifest kind must be support")
	}
	if support.TopLevelFlow {
		t.Fatalf("support top_level_flow must be false")
	}
}

func TestFlowManifestSkillRefsArePackPrefixed(t *testing.T) {
	paths := []string{
		"../../assets/flows/sdd.yaml",
		"../../assets/flows/sdr.yaml",
		"../../assets/flows/support.yaml",
		"../../assets/flows/source-security-review.yaml",
	}
	for _, p := range paths {
		m := readFlowManifest(t, p)
		for _, st := range m.Stages {
			if !strings.Contains(st.Skill, "/") {
				t.Fatalf("manifest %s has non-prefixed skill ref: %s", p, st.Skill)
			}
		}
	}
}

func readFlowManifest(t *testing.T, path string) flowManifest {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read manifest %s: %v", path, err)
	}
	var m flowManifest
	if err := yaml.Unmarshal(raw, &m); err != nil {
		t.Fatalf("unmarshal manifest %s: %v", path, err)
	}
	return m
}
