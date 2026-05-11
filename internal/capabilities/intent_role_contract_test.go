package capabilities

import (
	"os"
	"strings"
	"testing"
)

func TestIntentContracts_MinimumFields(t *testing.T) {
	cat, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	required := []string{
		"direct-answer",
		"architecture-change",
		"frontend-ux-change",
		"bug-investigation",
		"research-analysis",
		"security-review",
		"documentation-change",
	}
	for _, id := range required {
		in := findIntent(cat.Intents, id)
		if in == nil {
			t.Fatalf("missing intent %s", id)
		}
		if strings.TrimSpace(in.DefaultDecision) == "" || strings.TrimSpace(in.ComplexityThreshold) == "" {
			t.Fatalf("intent %s missing default_decision or complexity_threshold", id)
		}
		if strings.TrimSpace(in.Description) == "" || len(in.ForbiddenActions) == 0 {
			t.Fatalf("intent %s missing description or forbidden_actions", id)
		}
	}
}

func TestRoleContracts_MinimumFrontmatterAndSections(t *testing.T) {
	cat, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(cat.Roles) < 14 {
		t.Fatalf("expected role pack MVP, got %d", len(cat.Roles))
	}
	headings := []string{"## Mission", "## Use When", "## Inputs", "## Workflow", "## Outputs", "## Boundaries", "## Handoff back to bitsentry"}
	for _, r := range cat.Roles {
		if strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.Category) == "" || strings.TrimSpace(r.Kind) == "" {
			t.Fatalf("role missing frontmatter identity fields: %+v", r)
		}
		if len(r.UsableIn) == 0 || len(r.Permissions) == 0 {
			t.Fatalf("role %s missing usable_in or permissions", r.ID)
		}
		raw, err := os.ReadFile(r.SourcePath)
		if err != nil {
			t.Fatalf("read role %s: %v", r.ID, err)
		}
		text := string(raw)
		for _, h := range headings {
			if !strings.Contains(text, h) {
				t.Fatalf("role %s missing heading %q", r.ID, h)
			}
		}
	}
}

func findIntent(intents []DiscoveredIntent, id string) *DiscoveredIntent {
	for _, in := range intents {
		if in.ID == id {
			copy := in
			return &copy
		}
	}
	return nil
}
