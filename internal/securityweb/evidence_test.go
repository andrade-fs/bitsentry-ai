package securityweb

import (
	"strings"
	"testing"
)

func TestEvidenceEntryContainsRequiredContractFields(t *testing.T) {
	adapter := NewOfflineAdapter(DefaultPolicyEvaluator{}, DefaultDryRunPlanner{}, NewDefaultEvidenceRecorder(DefaultRedactor{}))
	entry := EvidenceEntry{
		EvidenceID:        "WEB-EV-1",
		SessionMode:       "sess-1 / planning_only",
		AuthorizationRef:  "auth-1",
		ScopeRef:          "scope-1",
		PlannedRequestRef: "req-1",
		PolicyDecision:    "allow",
		RedactionApplied:  true,
		LinkedFindingIDs:  []string{"F-1"},
		NotesAssumptions:  "offline",
	}
	tpl := adapter.RenderEvidenceTemplate(entry)
	required := []string{
		"- Evidence ID:",
		"- Session/Mode:",
		"- Authorization Ref:",
		"- Scope Ref:",
		"- Planned Request Ref:",
		"- Policy Decision:",
		"- Redaction Applied:",
		"- Linked Finding IDs:",
		"- Notes / Assumptions:",
	}
	for _, token := range required {
		if !strings.Contains(tpl, token) {
			t.Fatalf("missing required evidence field token %q", token)
		}
	}
}
