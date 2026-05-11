package capabilities

import (
	"fmt"
	"sort"
	"strings"
)

type RouteDecisionEnvelope struct {
	Input                    string   `json:"input"`
	MatchedIntent            string   `json:"matched_intent"`
	MatchedSignals           []string `json:"matched_signals"`
	Decision                 string   `json:"decision"`
	RecommendedFlow          string   `json:"recommended_flow"`
	RecommendedRoles         []string `json:"recommended_roles"`
	RecommendedSkills        []string `json:"recommended_skills"`
	Confidence               string   `json:"confidence"`
	Reason                   string   `json:"reason"`
	RequiresConfirmation     bool     `json:"requires_confirmation"`
	RequiresBoundedDiscovery bool     `json:"requires_bounded_discovery"`
	Gates                    []string `json:"gates"`
	Notes                    []string `json:"notes"`
}

func BuildRouteDecisionPreview(root string, input string) (RouteDecisionEnvelope, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return RouteDecisionEnvelope{}, fmt.Errorf("input prompt must not be empty")
	}

	cat, err := DiscoverAssets(root)
	if err != nil {
		return RouteDecisionEnvelope{}, fmt.Errorf("discover assets: %w", err)
	}

	intentID, reason, confidence, signals := classifyIntent(trimmed)
	intent := findIntentByID(cat.Intents, intentID)
	if intent == nil {
		return RouteDecisionEnvelope{}, fmt.Errorf("intent %q not found in assets/intents", intentID)
	}

	decision := strings.TrimSpace(intent.DefaultDecision)
	if decision == "" {
		decision = "ask_clarifying_question"
	}
	flow := strings.TrimSpace(intent.DefaultFlow)
	if flow == "none" {
		flow = ""
	}

	notes := []string{}
	if decision == "direct_answer" {
		notes = append(notes, "Direct answer selected to avoid over-engineering.")
	}
	if decision == "ask_clarifying_question" {
		notes = append(notes, "Clarifying question is required before safe route selection.")
	}

	gates := []string{
		"no_edits_in_preview",
		"no_persistence_in_preview",
		"no_flow_execution_in_preview",
	}
	if intent.RequiresConfirmation {
		gates = append(gates, "requires_confirmation")
	}
	if intent.RequiresBoundedDiscovery {
		gates = append(gates, "requires_bounded_discovery")
	}

	return RouteDecisionEnvelope{
		Input:                    trimmed,
		MatchedIntent:            intent.ID,
		MatchedSignals:           append([]string{}, signals...),
		Decision:                 decision,
		RecommendedFlow:          flow,
		RecommendedRoles:         append([]string{}, intent.PreFlowRoles...),
		RecommendedSkills:        append([]string{}, intent.PreFlowSkills...),
		Confidence:               confidence,
		Reason:                   reason,
		RequiresConfirmation:     intent.RequiresConfirmation,
		RequiresBoundedDiscovery: intent.RequiresBoundedDiscovery,
		Gates:                    uniqueSortedDecisionStrings(gates),
		Notes:                    notes,
	}, nil
}

func classifyIntent(input string) (intentID string, reason string, confidence string, signals []string) {
	t := normalizeIntentText(input)

	security := []string{"security", "secure", "seguridad", "appsec", "threat", "risk", "riesgo", "pentest", "hardening", "vulnerab", "owasp"}
	bug := []string{"bug", "error", "falla", "fallo", "regression", "regresión", "no funciona", "failing", "broken", "test roto", "tests rotos"}
	architecture := []string{"arquitect", "architecture", "refactor", "integration", "integración", "boundary", "boundaries", "projection", "opencode", "contrato", "design"}
	frontend := []string{"frontend", "ui", "ux", "tui", "wizard", "layout", "pantalla", "screen"}
	docs := []string{"docs", "documentation", "readme", "roadmap", "copy", "contenido", "documentación"}
	research := []string{"analiza", "analyze", "research", "investiga", "investigate", "repo externo", "sacar ideas", "comparar", "compare"}

	securitySignals := matchedKeywords(t, security)
	bugSignals := matchedKeywords(t, bug)
	architectureSignals := matchedKeywords(t, architecture)
	frontendSignals := matchedKeywords(t, frontend)
	docsSignals := matchedKeywords(t, docs)
	researchSignals := matchedKeywords(t, research)

	// Priority: security > bug > architecture > frontend > docs > research > direct
	switch {
	case len(securitySignals) > 0:
		return "security-review", "Request includes security/appsec/threat/risk signals and requires security-focused triage.", hitConfidence(len(securitySignals)), securitySignals
	case len(bugSignals) > 0:
		return "bug-investigation", "Request includes failure/regression/error signals; support-first triage is safer.", hitConfidence(len(bugSignals)), bugSignals
	case len(architectureSignals) > 0:
		return "architecture-change", "Request suggests architecture/integration/refactor impact and needs planning.", hitConfidence(len(architectureSignals)), architectureSignals
	case len(frontendSignals) > 0:
		return "frontend-ux-change", "Request targets UI/TUI/UX behavior and likely needs implementation planning.", hitConfidence(len(frontendSignals)), frontendSignals
	case len(docsSignals) > 0:
		return "documentation-change", "Request targets docs/content scope.", hitConfidence(len(docsSignals)), docsSignals
	case len(researchSignals) > 0:
		return "research-analysis", "Request asks for analysis/research/comparison work.", hitConfidence(len(researchSignals)), researchSignals
	default:
		if looksLikeSimpleQuestion(t) {
			return "direct-answer", "Prompt looks like a conceptual/simple question; direct answer is sufficient.", "high", []string{"question-pattern"}
		}
		return "direct-answer", "No high-risk implementation signals detected; defaulting to direct answer.", "low", []string{"no-high-risk-signals"}
	}
}

func findIntentByID(intents []DiscoveredIntent, id string) *DiscoveredIntent {
	for _, in := range intents {
		if in.ID == id {
			copy := in
			return &copy
		}
	}
	return nil
}

func normalizeIntentText(input string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(input))), " ")
}

func matchedKeywords(input string, keywords []string) []string {
	out := []string{}
	for _, k := range keywords {
		if strings.Contains(input, k) {
			out = append(out, k)
		}
	}
	return uniqueSortedDecisionStrings(out)
}

func looksLikeSimpleQuestion(input string) bool {
	if strings.HasPrefix(input, "qué ") || strings.HasPrefix(input, "que ") || strings.HasPrefix(input, "what ") {
		return true
	}
	return strings.Contains(input, "?")
}

func hitConfidence(hits int) string {
	switch {
	case hits >= 2:
		return "high"
	case hits == 1:
		return "medium"
	default:
		return "low"
	}
}

func uniqueSortedDecisionStrings(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" || set[t] {
			continue
		}
		set[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
