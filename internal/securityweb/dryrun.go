package securityweb

import "fmt"

type DefaultDryRunPlanner struct{}

var passiveDiscoveryMVP = []struct {
	ref    string
	method RequestMethod
	path   string
}{
	{ref: "disc-req-001", method: MethodHEAD, path: "/"},
	{ref: "disc-req-002", method: MethodGET, path: "/"},
	{ref: "disc-req-003", method: MethodGET, path: "/robots.txt"},
	{ref: "disc-req-004", method: MethodGET, path: "/sitemap.xml"},
	{ref: "disc-req-005", method: MethodGET, path: "/.well-known/security.txt"},
}

func (p DefaultDryRunPlanner) BuildPlan(_ AssessmentSessionContext, req PlannedRequest) PlannedRequest {
	if req.RequestRef == "" {
		req.RequestRef = "planned-request"
	}
	return req
}

func (a *OfflineAdapter) BuildDiscoveryPlan(ctx AssessmentSessionContext, target string) DiscoveryPlan {
	plan := DiscoveryPlan{
		PlanRef:       "discovery-plan-mvp-v1",
		Target:        target,
		ExecutionMode: ctx.ExecutionMode,
		WouldExecute:  false,
		Items:         make([]DiscoveryPlanItem, 0, len(passiveDiscoveryMVP)),
	}

	for idx, step := range passiveDiscoveryMVP {
		req := PlannedRequest{
			RequestRef: step.ref,
			Target:     target,
			URL:        fmt.Sprintf("https://%s%s", target, step.path),
			Method:     step.method,
			ToolClass:  ToolClassSafeProbe,
		}

		decision, violations := a.Validate(ctx, req)
		entry := EvidenceEntry{
			EvidenceID:        fmt.Sprintf("WEB-EV-DISC-%03d", idx+1),
			SessionMode:       fmt.Sprintf("%s / %s", ctx.SessionID, ctx.ExecutionMode),
			AuthorizationRef:  ctx.AuthorizationRef,
			ScopeRef:          ctx.ScopeRef,
			PlannedRequestRef: req.RequestRef,
			PolicyDecision:    fmt.Sprintf("%t (%s)", decision.Allowed, decision.Reason),
			RequestURL:        req.URL,
			RequestHeaders:    req.Headers,
		}

		plan.Items = append(plan.Items, DiscoveryPlanItem{
			Request:          req,
			PolicyDecision:   decision,
			PolicyViolations: violations,
			EvidenceTemplate: entry,
		})
	}

	return plan
}
