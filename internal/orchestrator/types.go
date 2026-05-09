package orchestrator

type Intent string

type RouteRequest struct {
	Intent   Intent
	FlowHint string
}

type FlowRoute struct {
	FlowID string
	Reason string
}

type ExecutionStage struct {
	ID          string
	Skill       string
	Description string
	Order       int
}

type ExecutionPlan struct {
	FlowID string
	Stages []ExecutionStage
}

type RouteResult struct {
	Flow         string
	InitialSkill string
	Plan         ExecutionPlan
	Warnings     []string
}

type Orchestrator interface {
	Route(RouteRequest) (RouteResult, error)
}
