package orchestrator

import (
	"fmt"
	"strings"

	"bitsentry-ai/internal/capabilities"
)

type Planner struct{}

func NewPlanner() Planner {
	return Planner{}
}

func (p Planner) BuildExecutionPlan(flowID string, flows []capabilities.DiscoveredFlow) (ExecutionPlan, error) {
	manifest, ok := findFlow(flowID, flows)
	if !ok {
		return ExecutionPlan{}, fmt.Errorf("flow manifest not found: %s", flowID)
	}

	plan := ExecutionPlan{FlowID: manifest.ID}
	for i, rawStage := range manifest.Stages {
		stage := ExecutionStage{
			ID:          asString(rawStage["id"]),
			Skill:       asString(rawStage["skill"]),
			Description: asString(rawStage["description"]),
			Order:       i + 1,
		}
		plan.Stages = append(plan.Stages, stage)
	}

	return plan, nil
}

func findFlow(flowID string, flows []capabilities.DiscoveredFlow) (capabilities.DiscoveredFlow, bool) {
	trimmed := strings.TrimSpace(flowID)
	for _, f := range flows {
		if f.ID == trimmed {
			return f, true
		}
	}
	return capabilities.DiscoveredFlow{}, false
}

func asString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}
