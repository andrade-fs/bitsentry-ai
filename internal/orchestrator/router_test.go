package orchestrator

import (
	"testing"

	"bitsentry-ai/internal/capabilities"
)

func TestRouteIntentToFlow_SDD(t *testing.T) {
	r := NewRouter()
	got, err := r.RouteIntentToFlow(RouteRequest{Intent: "design a feature change"}, testFlows())
	if err != nil {
		t.Fatalf("route intent: %v", err)
	}
	if got.FlowID != "sdd" {
		t.Fatalf("expected sdd, got %s", got.FlowID)
	}
}

func TestRouteIntentToFlow_SDR(t *testing.T) {
	r := NewRouter()
	got, err := r.RouteIntentToFlow(RouteRequest{Intent: "security incident logs analysis"}, testFlows())
	if err != nil {
		t.Fatalf("route intent: %v", err)
	}
	if got.FlowID != "sdr" {
		t.Fatalf("expected sdr, got %s", got.FlowID)
	}
}

func TestRouteIntentToFlow_Support(t *testing.T) {
	r := NewRouter()
	got, err := r.RouteIntentToFlow(RouteRequest{Intent: "help troubleshoot this error"}, testFlows())
	if err != nil {
		t.Fatalf("route intent: %v", err)
	}
	if got.FlowID != "support" {
		t.Fatalf("expected support, got %s", got.FlowID)
	}
}

func TestFlowHintOverridesHeuristicsWhenValid(t *testing.T) {
	r := NewRouter()
	got, err := r.RouteIntentToFlow(RouteRequest{Intent: "design feature", FlowHint: "sdr"}, testFlows())
	if err != nil {
		t.Fatalf("route intent: %v", err)
	}
	if got.FlowID != "sdr" {
		t.Fatalf("expected hinted sdr, got %s", got.FlowID)
	}
}

func TestUnknownIntentReturnsSafeError(t *testing.T) {
	r := NewRouter()
	_, err := r.RouteIntentToFlow(RouteRequest{Intent: "totally unrelated tokens"}, testFlows())
	if err == nil {
		t.Fatalf("expected routing error for unknown intent")
	}
}

func testFlows() []capabilities.DiscoveredFlow {
	return []capabilities.DiscoveredFlow{{ID: "sdd"}, {ID: "sdr"}, {ID: "support"}}
}
