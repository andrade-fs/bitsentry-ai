package orchestrator

import "testing"

func TestOrchestratorUsesDiscoveredAssets(t *testing.T) {
	o := New("../..")
	res, err := o.Route(RouteRequest{Intent: "design feature change"})
	if err != nil {
		t.Fatalf("route orchestration: %v", err)
	}
	if res.Flow != "sdd" {
		t.Fatalf("expected sdd flow, got %s", res.Flow)
	}
	if res.InitialSkill == "" {
		t.Fatalf("expected initial skill from discovered flow stages")
	}
}

func TestNoOpenCodeMutationDuringPlanning(t *testing.T) {
	o := New("../..")
	res, err := o.Route(RouteRequest{Intent: "help troubleshoot error", FlowHint: "support"})
	if err != nil {
		t.Fatalf("route orchestration: %v", err)
	}
	if res.Flow != "support" {
		t.Fatalf("expected support flow, got %s", res.Flow)
	}
	if len(res.Warnings) != 0 {
		t.Fatalf("expected no side-effect warnings, got %v", res.Warnings)
	}
}
