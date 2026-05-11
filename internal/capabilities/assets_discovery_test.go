package capabilities

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverAssets_RepoRoot(t *testing.T) {
	catalog, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	if len(catalog.Flows) < 3 {
		t.Fatalf("expected at least 3 flows, got %d", len(catalog.Flows))
	}
	if !hasFlowID(catalog.Flows, "sdd") || !hasFlowID(catalog.Flows, "sdr") || !hasFlowID(catalog.Flows, "support") {
		t.Fatalf("expected sdd/sdr/support flows")
	}
	if !hasPackID(catalog.SkillPacks, "sdd") || !hasPackID(catalog.SkillPacks, "sdr") || !hasPackID(catalog.SkillPacks, "support") {
		t.Fatalf("expected sdd/sdr/support skill packs")
	}
	if !hasPackID(catalog.SkillPacks, "_shared") {
		t.Fatalf("expected _shared skill pack")
	}
	if len(catalog.Skills) == 0 {
		t.Fatalf("expected discovered skills")
	}
	if len(catalog.Intents) < 7 {
		t.Fatalf("expected at least 7 intents, got %d", len(catalog.Intents))
	}
	if len(catalog.Roles) < 14 {
		t.Fatalf("expected at least 14 roles, got %d", len(catalog.Roles))
	}
	for _, s := range catalog.Skills {
		if s.Status != "valid" {
			t.Fatalf("expected all repository skills to be valid, got %s for %s", s.Status, s.ID)
		}
	}
}

func TestDiscoverAssets_HandlesMissingOrchestratorsDirectory(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "assets/flows/fake.yaml"), "id: fake\nname: Fake\nkind: flow\nselectable: true\ntop_level_flow: true\nfamily: fake\nskill_pack: fake\norchestrator_skill: fake/orch\nstatus: available\ntriggers: [\"fake\"]\ncontracts: []\nrequires: {}\npersistence: {}\nstages: []\nstage_graph: {}\nhandoffs: []\nfinal_artifacts: []\noutputs: []\n")
	mustWrite(t, filepath.Join(root, "assets/skills/fake/fake-orchestrator/SKILL.md"), validSkillContent("fake-orchestrator"))
	mustWrite(t, filepath.Join(root, "assets/skills/_shared/x.md"), "# x\ncontent\n")

	catalog, err := DiscoverAssets(root)
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	if len(catalog.Orchestrators) != 0 {
		t.Fatalf("expected no orchestrators when directory is absent")
	}
	if len(catalog.Intents) != 0 {
		t.Fatalf("expected no intents when directory is absent")
	}
	if len(catalog.Roles) != 0 {
		t.Fatalf("expected no roles when directory is absent")
	}
}

func TestDiscoverAssets_DynamicPackAndSkillValidation(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "assets/flows/fake-flow.yaml"), "id: fake-flow\nname: Fake Flow\nkind: flow\nselectable: true\ntop_level_flow: true\nfamily: fake\nskill_pack: fake-pack\norchestrator_skill: fake-pack/fake-orchestrator\nstatus: available\ntriggers: [\"fake\"]\ncontracts: [\"_shared/result-envelope.md\"]\nrequires: {}\npersistence: {}\nstages: [{name: fake, skill: fake-pack/fake-orchestrator}]\nstage_graph: {fake: []}\nhandoffs: []\nfinal_artifacts: []\noutputs: []\n")
	mustWrite(t, filepath.Join(root, "assets/skills/fake-pack/fake-orchestrator/SKILL.md"), validSkillContent("fake-orchestrator"))
	mustWrite(t, filepath.Join(root, "assets/skills/fake-pack/bad-skill/SKILL.md"), "# Skill: bad-skill\n\n## Purpose\nOnly purpose section\n")
	mustWrite(t, filepath.Join(root, "assets/skills/_shared/result-envelope.md"), "# Result Envelope\nNon-empty\n")
	mustWrite(t, filepath.Join(root, "assets/orchestrators/fake.md"), "# Fake Orchestrator\nDetails\n")
	mustWrite(t, filepath.Join(root, "assets/intents/direct-answer.yaml"), "id: direct-answer\ndescription: x\ndefault_decision: direct_answer\ndefault_flow: none\nalternative_flow: support\ncomplexity_threshold: low\npre_flow_roles: []\npre_flow_skills: []\nexpected_context_outputs: []\ndirect_answer_allowed: true\nrequires_confirmation: false\nrequires_bounded_discovery: false\nforbidden_actions: []\n")
	mustWrite(t, filepath.Join(root, "assets/roles/codebase-onboarding.md"), "---\nid: codebase-onboarding\ncategory: engineering\nkind: specialist\nusable_in: [sdd]\npermissions: {read: allow, edit: ask}\n---\n# Role: codebase-onboarding\n")

	catalog, err := DiscoverAssets(filepath.Join(root, "assets"))
	if err != nil {
		t.Fatalf("discover assets: %v", err)
	}
	if !hasFlowID(catalog.Flows, "fake-flow") {
		t.Fatalf("expected dynamically discovered fake-flow")
	}
	if !hasPackID(catalog.SkillPacks, "fake-pack") {
		t.Fatalf("expected dynamically discovered fake-pack")
	}
	if !hasSkillID(catalog.Skills, "fake-pack/fake-orchestrator") {
		t.Fatalf("expected discovered fake-pack/fake-orchestrator skill")
	}
	bad := findSkill(catalog.Skills, "fake-pack/bad-skill")
	if bad == nil {
		t.Fatalf("expected discovered fake-pack/bad-skill")
	}
	if bad.Status != "invalid" {
		t.Fatalf("expected bad skill invalid, got %s", bad.Status)
	}
	if len(bad.RequiredHeadingsMissing) == 0 {
		t.Fatalf("expected missing headings for invalid skill")
	}
	if len(catalog.Orchestrators) != 1 || catalog.Orchestrators[0].ID != "fake" {
		t.Fatalf("expected one discovered orchestrator file")
	}
	if len(catalog.Shared) != 1 || catalog.Shared[0].ID != "result-envelope" {
		t.Fatalf("expected one discovered shared contract")
	}
	if len(catalog.Intents) != 1 || catalog.Intents[0].ID != "direct-answer" {
		t.Fatalf("expected one discovered intent")
	}
	if len(catalog.Roles) != 1 || catalog.Roles[0].ID != "codebase-onboarding" {
		t.Fatalf("expected one discovered role")
	}
}

func validSkillContent(name string) string {
	return "---\nname: " + name + "\ndescription: test skill\n---\n\n# Skill: " + name + "\n\n## Purpose\nP\n\n## Use When\nU\n\n## Inputs\nI\n\n## Workflow\nW\n\n## Outputs\nO\n\n## Boundaries\nB\n\n## Persistence Actions\nPA\n\n## Result Envelope\nRE\n\n## Handoffs\nH\n\n## Quality Checklist\nQ\n"
}

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func hasFlowID(flows []DiscoveredFlow, id string) bool {
	for _, f := range flows {
		if f.ID == id {
			return true
		}
	}
	return false
}

func hasPackID(packs []DiscoveredSkillPack, id string) bool {
	for _, p := range packs {
		if p.ID == id {
			return true
		}
	}
	return false
}

func hasSkillID(skills []DiscoveredSkill, id string) bool {
	for _, s := range skills {
		if s.ID == id {
			return true
		}
	}
	return false
}

func findSkill(skills []DiscoveredSkill, id string) *DiscoveredSkill {
	for _, s := range skills {
		if s.ID == id {
			copy := s
			return &copy
		}
	}
	return nil
}
