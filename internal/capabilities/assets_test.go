package capabilities

import (
	"os"
	"testing"
)

func TestFlowManifestsExist(t *testing.T) {
	paths := []string{
		"../../assets/flows/sdd.yaml",
		"../../assets/flows/sdr.yaml",
		"../../assets/flows/support.yaml",
		"../../assets/flows/source-security-review.yaml",
	}
	for _, p := range paths {
		if !fileExists(p) {
			t.Fatalf("expected flow manifest to exist: %s", p)
		}
	}
}

func TestPhase6IntentAndRoleAssetsExist(t *testing.T) {
	paths := []string{
		"../../assets/intents/direct-answer.yaml",
		"../../assets/intents/architecture-change.yaml",
		"../../assets/intents/frontend-ux-change.yaml",
		"../../assets/intents/bug-investigation.yaml",
		"../../assets/intents/research-analysis.yaml",
		"../../assets/intents/security-review.yaml",
		"../../assets/intents/documentation-change.yaml",
		"../../assets/roles/codebase-onboarding.md",
		"../../assets/roles/software-architect.md",
		"../../assets/roles/backend-engineer.md",
		"../../assets/roles/frontend-engineer.md",
		"../../assets/roles/test-engineer.md",
		"../../assets/roles/code-reviewer.md",
		"../../assets/roles/security-reviewer.md",
		"../../assets/roles/appsec-reviewer.md",
		"../../assets/roles/threat-modeler.md",
		"../../assets/roles/product-analyst.md",
		"../../assets/roles/ux-flow-designer.md",
		"../../assets/roles/technical-writer.md",
		"../../assets/roles/bug-triage-engineer.md",
		"../../assets/roles/incident-analyst.md",
	}
	for _, p := range paths {
		if !fileExists(p) {
			t.Fatalf("expected phase 6 asset to exist: %s", p)
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
