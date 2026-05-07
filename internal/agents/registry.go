package agents

import "context"

type Registry struct {
	detectors []AgentDetector
}

func NewRegistry(detectors ...AgentDetector) *Registry {
	return &Registry{detectors: detectors}
}

func (r *Registry) List(ctx context.Context) ([]AgentDetectionResult, error) {
	results := make([]AgentDetectionResult, 0, len(r.detectors))
	for _, d := range r.detectors {
		result, err := d.Detect(ctx)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}
