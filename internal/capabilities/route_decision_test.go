package capabilities

import "testing"

func TestBuildRouteDecisionPreview_SimpleConceptual(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "qué es un embedding en RAG?")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "direct-answer" || res.Decision != "direct_answer" {
		t.Fatalf("expected direct-answer/direct_answer, got %s/%s", res.MatchedIntent, res.Decision)
	}
}

func TestBuildRouteDecisionPreview_Architecture(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "Quiero cambiar cómo se proyectan los skills a OpenCode")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "architecture-change" || res.Decision != "use_flow_sdd" {
		t.Fatalf("expected architecture-change/use_flow_sdd, got %s/%s", res.MatchedIntent, res.Decision)
	}
}

func TestBuildRouteDecisionPreview_FrontendTUI(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "Quiero mejorar el wizard del TUI")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "frontend-ux-change" || res.Decision != "use_flow_sdd" {
		t.Fatalf("expected frontend-ux-change/use_flow_sdd, got %s/%s", res.MatchedIntent, res.Decision)
	}
}

func TestBuildRouteDecisionPreview_Bug(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "me falla export-preview con --select")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "bug-investigation" || res.Decision != "use_flow_support" {
		t.Fatalf("expected bug-investigation/use_flow_support, got %s/%s", res.MatchedIntent, res.Decision)
	}
}

func TestBuildRouteDecisionPreview_Research(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "Analiza este repo externo para sacar ideas")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "research-analysis" {
		t.Fatalf("expected research-analysis, got %s", res.MatchedIntent)
	}
}

func TestBuildRouteDecisionPreview_Security(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "haz threat model y review de appsec")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "security-review" {
		t.Fatalf("expected security-review, got %s", res.MatchedIntent)
	}
}

func TestBuildRouteDecisionPreview_Docs(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "actualiza README y roadmap")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "documentation-change" {
		t.Fatalf("expected documentation-change, got %s", res.MatchedIntent)
	}
}

func TestBuildRouteDecisionPreview_OverlapSecurityWins(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "hay un bug de seguridad en export")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if res.MatchedIntent != "security-review" {
		t.Fatalf("expected security-review to win overlap, got %s", res.MatchedIntent)
	}
	if len(res.MatchedSignals) == 0 {
		t.Fatalf("expected matched signals for overlap decision")
	}
}

func TestBuildRouteDecisionPreview_GatesAlwaysSafe(t *testing.T) {
	res, err := BuildRouteDecisionPreview("../..", "Quiero mejorar el wizard del TUI")
	if err != nil {
		t.Fatalf("decision preview: %v", err)
	}
	if !containsToken(res.Gates, "no_edits_in_preview") || !containsToken(res.Gates, "no_persistence_in_preview") || !containsToken(res.Gates, "no_flow_execution_in_preview") {
		t.Fatalf("missing mandatory preview gates: %#v", res.Gates)
	}
}

func containsToken(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}
