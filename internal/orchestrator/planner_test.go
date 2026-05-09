package orchestrator

import (
	"testing"

	"bitsentry-ai/internal/capabilities"
)

func TestBuildExecutionPlanFromFlowManifest(t *testing.T) {
	p := NewPlanner()
	flow := capabilities.DiscoveredFlow{
		ID: "sdd",
		Stages: []map[string]any{
			{"id": "init", "skill": "sdd/sdd-init", "description": "init stage"},
			{"id": "apply", "skill": "sdd/sdd-apply", "description": "apply stage"},
		},
	}

	plan, err := p.BuildExecutionPlan("sdd", []capabilities.DiscoveredFlow{flow})
	if err != nil {
		t.Fatalf("build execution plan: %v", err)
	}
	if plan.FlowID != "sdd" {
		t.Fatalf("expected flow sdd, got %s", plan.FlowID)
	}
	if len(plan.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(plan.Stages))
	}
	if plan.Stages[0].Order != 1 || plan.Stages[1].Order != 2 {
		t.Fatalf("expected stable stage order 1..2")
	}
}

func TestExecutionPlanDoesNotExecuteAgents(t *testing.T) {
	p := NewPlanner()
	flow := capabilities.DiscoveredFlow{
		ID: "support",
		Stages: []map[string]any{{"id": "quality", "skill": "support/judgment-day", "description": "review"}},
	}

	plan, err := p.BuildExecutionPlan("support", []capabilities.DiscoveredFlow{flow})
	if err != nil {
		t.Fatalf("build execution plan: %v", err)
	}
	if len(plan.Stages) != 1 {
		t.Fatalf("expected single stage plan")
	}
	if plan.Stages[0].Skill != "support/judgment-day" {
		t.Fatalf("unexpected stage skill: %s", plan.Stages[0].Skill)
	}
}
