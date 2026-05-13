package securityweb

import "testing"

func baseContext() AssessmentSessionContext {
	return AssessmentSessionContext{
		SessionID:            "sess-1",
		AuthorizationRef:     "auth-1",
		ScopeRef:             "scope-1",
		InScopeTargets:       []string{"example.com"},
		ExecutionMode:        ExecutionModeExecuteApproved,
		Intensity:            IntensityLow,
		ExplicitApproval:     true,
		RateLimitPerMinute:   10,
		RequestBudget:        20,
		TimeoutSeconds:       5,
		MaxResponseSizeBytes: 1024,
		MaxPreviewSizeBytes:  256,
		StopConditions:       []string{"max-errors"},
		EvidencePlanRef:      "plan-1",
	}
}

func baseRequest() PlannedRequest {
	return PlannedRequest{
		RequestRef: "req-1",
		Target:     "example.com",
		URL:        "https://example.com/health",
		Method:     MethodGET,
		ToolClass:  ToolClassManual,
	}
}

func hasViolation(vs []PolicyViolation, code string) bool {
	for _, v := range vs {
		if v.Code == code {
			return true
		}
	}
	return false
}

func TestPlanningOnlyNeverExecutes(t *testing.T) {
	ctx := baseContext()
	ctx.ExecutionMode = ExecutionModePlanningOnly
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "planning_only_never_executes") {
		t.Fatalf("expected planning_only_never_executes violation")
	}
}

func TestDryRunNeverExecutes(t *testing.T) {
	ctx := baseContext()
	ctx.ExecutionMode = ExecutionModeDryRun
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "dry_run_never_executes") {
		t.Fatalf("expected dry_run_never_executes violation")
	}
}

func TestExecuteApprovedRequiresExplicitApproval(t *testing.T) {
	ctx := baseContext()
	ctx.ExplicitApproval = false
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "execute_approved_requires_explicit_approval") {
		t.Fatalf("expected explicit approval violation")
	}
}

func TestOutOfScopeTargetDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.Target = "other.example"
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, req)
	if !hasViolation(v, "target_out_of_scope") {
		t.Fatalf("expected target_out_of_scope violation")
	}
}

func TestNonHTTPSchemeDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.URL = "ftp://example.com/file"
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, req)
	if !hasViolation(v, "scheme_must_be_http_https") {
		t.Fatalf("expected scheme_must_be_http_https violation")
	}
}

func TestPOSTDeniedByDefault(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.Method = MethodPOST
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, req)
	if !hasViolation(v, "post_denied_by_default") {
		t.Fatalf("expected post_denied_by_default violation")
	}
}

func TestMissingRateLimitDeniedForExecuteApproved(t *testing.T) {
	ctx := baseContext()
	ctx.RateLimitPerMinute = 0
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "missing_rate_limit") {
		t.Fatalf("expected missing_rate_limit violation")
	}
}

func TestMissingRequestBudgetDeniedForExecuteApproved(t *testing.T) {
	ctx := baseContext()
	ctx.RequestBudget = 0
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "missing_request_budget") {
		t.Fatalf("expected missing_request_budget violation")
	}
}

func TestMissingStopConditionsDenied(t *testing.T) {
	ctx := baseContext()
	ctx.StopConditions = nil
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, baseRequest())
	if !hasViolation(v, "missing_stop_conditions") {
		t.Fatalf("expected missing_stop_conditions violation")
	}
}

func TestProhibitedToolClassDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.ToolClass = ToolClassProhibited
	_, v := DefaultPolicyEvaluator{}.Evaluate(ctx, req)
	if !hasViolation(v, "prohibited_tool_class_denied") {
		t.Fatalf("expected prohibited_tool_class_denied violation")
	}
}
