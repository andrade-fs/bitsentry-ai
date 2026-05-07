package capabilities

import (
	"os"
	"strings"
	"testing"
)

func TestAllSkillFilesContainRequiredSections(t *testing.T) {
	skillFiles := []string{
		"../../assets/skills/sdd/sdd-orchestrator/SKILL.md",
		"../../assets/skills/sdd/sdd-init/SKILL.md",
		"../../assets/skills/sdd/sdd-explore/SKILL.md",
		"../../assets/skills/sdd/sdd-propose/SKILL.md",
		"../../assets/skills/sdd/sdd-spec/SKILL.md",
		"../../assets/skills/sdd/sdd-design/SKILL.md",
		"../../assets/skills/sdd/sdd-tasks/SKILL.md",
		"../../assets/skills/sdd/sdd-apply/SKILL.md",
		"../../assets/skills/sdd/sdd-verify/SKILL.md",
		"../../assets/skills/sdd/sdd-archive/SKILL.md",
		"../../assets/skills/sdr/sdr-orchestrator/SKILL.md",
		"../../assets/skills/sdr/sdr-capture/SKILL.md",
		"../../assets/skills/sdr/sdr-research/SKILL.md",
		"../../assets/skills/sdr/sdr-synthesis/SKILL.md",
		"../../assets/skills/sdr/sdr-questions/SKILL.md",
		"../../assets/skills/sdr/sdr-structure/SKILL.md",
		"../../assets/skills/sdr/sdr-validate/SKILL.md",
		"../../assets/skills/sdr/sdr-archive/SKILL.md",
		"../../assets/skills/support/skill-registry/SKILL.md",
		"../../assets/skills/support/judgment-day/SKILL.md",
		"../../assets/skills/support/go-testing/SKILL.md",
		"../../assets/skills/support/skill-creator/SKILL.md",
		"../../assets/skills/support/issue-creation/SKILL.md",
		"../../assets/skills/support/branch-pr/SKILL.md",
	}

	required := []string{
		"## Purpose",
		"## Use When",
		"## Inputs",
		"## Workflow",
		"## Outputs",
		"## Boundaries",
		"## Persistence Actions",
		"## Result Envelope",
		"## Handoffs",
		"## Quality Checklist",
	}

	for _, path := range skillFiles {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read skill file %s: %v", path, err)
		}
		content := string(raw)
		if strings.TrimSpace(content) == "" {
			t.Fatalf("skill file is empty: %s", path)
		}
		for _, h := range required {
			if !strings.Contains(content, h) {
				t.Fatalf("skill file %s missing heading %q", path, h)
			}
		}
	}
}

func TestSharedContractsAndFlowsAreNonEmpty(t *testing.T) {
	paths := []string{
		"../../assets/skills/_shared/persistence-contract.md",
		"../../assets/skills/_shared/result-envelope.md",
		"../../assets/skills/_shared/handoff-contract.md",
		"../../assets/skills/_shared/engram-convention.md",
		"../../assets/skills/_shared/opencode-convention.md",
		"../../assets/skills/_shared/skill-loading.md",
		"../../assets/flows/sdd.yaml",
		"../../assets/flows/sdr.yaml",
		"../../assets/flows/support.yaml",
	}
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		if strings.TrimSpace(string(raw)) == "" {
			t.Fatalf("file is empty: %s", p)
		}
	}
}
