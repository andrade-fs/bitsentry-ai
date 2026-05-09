package orchestrator

import (
	"fmt"
	"strings"

	"bitsentry-ai/internal/capabilities"
)

type Facade struct {
	root    string
	router  Router
	planner Planner
}

func New(root string) *Facade {
	return &Facade{
		root:    root,
		router:  NewRouter(),
		planner: NewPlanner(),
	}
}

func (o *Facade) Route(req RouteRequest) (RouteResult, error) {
	assets, err := capabilities.DiscoverAssets(o.root)
	if err != nil {
		return RouteResult{}, fmt.Errorf("discover assets: %w", err)
	}

	route, err := o.router.RouteIntentToFlow(req, assets.Flows)
	if err != nil {
		return RouteResult{}, err
	}

	plan, err := o.planner.BuildExecutionPlan(route.FlowID, assets.Flows)
	if err != nil {
		return RouteResult{}, err
	}

	result := RouteResult{
		Flow:     route.FlowID,
		Plan:     plan,
		Warnings: []string{},
	}
	if len(plan.Stages) > 0 {
		result.InitialSkill = strings.TrimSpace(plan.Stages[0].Skill)
	}
	return result, nil
}
