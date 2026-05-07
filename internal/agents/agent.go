package agents

import "context"

type AgentDetectionResult struct {
	ID      string
	Name    string
	Found   bool
	Path    string
	Version string
	Hint    string
}

type AgentDetector interface {
	ID() string
	Name() string
	Detect(ctx context.Context) (AgentDetectionResult, error)
}
