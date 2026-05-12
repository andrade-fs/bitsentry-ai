package capabilities

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestAllSkillFilesContainRequiredSections(t *testing.T) {
	groups := []string{"sdd", "sdr", "support", "security"}
	skillFiles := make([]string, 0)
	for _, group := range groups {
		pattern := filepath.Join("..", "..", "assets", "skills", group, "*", "SKILL.md")
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		skillFiles = append(skillFiles, matches...)
	}
	sort.Strings(skillFiles)
	if len(skillFiles) == 0 {
		t.Fatal("no SKILL.md files found under assets/skills/{sdd,sdr,support,security}")
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
			if !hasRequiredHeading(content, h) {
				t.Fatalf("skill file %s missing heading %q", path, h)
			}
		}
	}
}

func TestHasExactHeadingLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		wantOK bool
	}{
		{name: "exact", line: "## Handoffs", wantOK: true},
		{name: "trimmed exact", line: "  ## Handoffs  ", wantOK: true},
		{name: "duplicated hashes", line: "## ## Handoffs", wantOK: false},
		{name: "different level", line: "### Handoffs", wantOK: false},
		{name: "prefixed text", line: "texto ## Handoffs", wantOK: false},
		{name: "suffix text", line: "## Handoffs extra", wantOK: false},
		{name: "partial heading", line: "## Hand", wantOK: false},
		{name: "inside paragraph", line: "En este párrafo va ## Handoffs incrustado.", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := strings.Join([]string{"# Skill: x", tt.line}, "\n")
			got := hasExactHeadingLine(content, "## Handoffs")
			if got != tt.wantOK {
				t.Fatalf("hasExactHeadingLine(%q) = %v, want %v", tt.line, got, tt.wantOK)
			}
		})
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
		"../../assets/flows/source-security-review.yaml",
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
