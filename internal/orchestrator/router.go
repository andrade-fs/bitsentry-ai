package orchestrator

import (
	"fmt"
	"sort"
	"strings"

	"bitsentry-ai/internal/capabilities"
)

var errUnknownIntent = fmt.Errorf("unable to resolve flow safely from intent")

type Router struct{}

func NewRouter() Router {
	return Router{}
}

func (r Router) RouteIntentToFlow(req RouteRequest, flows []capabilities.DiscoveredFlow) (FlowRoute, error) {
	available := availableFlows(flows)
	hint := strings.TrimSpace(req.FlowHint)
	if hint != "" {
		if available[hint] {
			return FlowRoute{FlowID: hint, Reason: "flow hint"}, nil
		}
		return FlowRoute{}, fmt.Errorf("invalid flow hint %q: available flows=%s", hint, strings.Join(sortedKeys(available), ","))
	}

	intent := strings.ToLower(strings.TrimSpace(string(req.Intent)))
	if intent == "" {
		return FlowRoute{}, errUnknownIntent
	}

	matches := make([]FlowRoute, 0, 3)
	if hasAny(intent, []string{"spec", "design", "feature", "change"}) && available["sdd"] {
		matches = append(matches, FlowRoute{FlowID: "sdd", Reason: "sdd heuristic"})
	}
	if hasAny(intent, []string{"security", "detection", "incident", "threat", "logs"}) && available["sdr"] {
		matches = append(matches, FlowRoute{FlowID: "sdr", Reason: "sdr heuristic"})
	}
	if hasAny(intent, []string{"support", "bug", "help", "troubleshoot", "error"}) && available["support"] {
		matches = append(matches, FlowRoute{FlowID: "support", Reason: "support heuristic"})
	}

	if len(matches) == 1 {
		return matches[0], nil
	}
	if len(matches) > 1 {
		matched := make([]string, 0, len(matches))
		for _, m := range matches {
			matched = append(matched, m.FlowID)
		}
		sort.Strings(matched)
		return FlowRoute{}, fmt.Errorf("%w: ambiguous intent matched multiple flows=%s; provide flow hint", errUnknownIntent, strings.Join(matched, ","))
	}

	return FlowRoute{}, fmt.Errorf("%w: available flows=%s", errUnknownIntent, strings.Join(sortedKeys(available), ","))
}

func hasAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func availableFlows(flows []capabilities.DiscoveredFlow) map[string]bool {
	set := map[string]bool{}
	for _, f := range flows {
		id := strings.TrimSpace(f.ID)
		if id == "" {
			continue
		}
		set[id] = true
	}
	return set
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
