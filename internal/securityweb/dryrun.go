package securityweb

type DefaultDryRunPlanner struct{}

func (p DefaultDryRunPlanner) BuildPlan(_ AssessmentSessionContext, req PlannedRequest) PlannedRequest {
	if req.RequestRef == "" {
		req.RequestRef = "planned-request"
	}
	return req
}
