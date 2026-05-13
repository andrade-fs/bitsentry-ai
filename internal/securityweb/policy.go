package securityweb

import (
	"net/url"
	"slices"
	"strings"
)

type DefaultPolicyEvaluator struct{}

func (p DefaultPolicyEvaluator) Evaluate(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation) {
	violations := make([]PolicyViolation, 0)

	if ctx.ExecutionMode == ExecutionModePlanningOnly {
		violations = append(violations, violation("planning_only_never_executes", ErrExecutionModeDenied, "execution_mode"))
	}
	if ctx.ExecutionMode == ExecutionModeDryRun {
		violations = append(violations, violation("dry_run_never_executes", ErrExecutionModeDenied, "execution_mode"))
	}

	if ctx.ExecutionMode == ExecutionModeExecuteApproved && !ctx.ExplicitApproval {
		violations = append(violations, violation("execute_approved_requires_explicit_approval", ErrMissingExplicitApproval, "explicit_approval"))
	}

	if ctx.ExecutionMode == ExecutionModeRetest && ctx.ExistingFindingID == "" && ctx.ExistingCheckID == "" {
		violations = append(violations, violation("retest_requires_existing_context", ErrExecutionModeDenied, "retest_context"))
	}

	u, err := url.Parse(req.URL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		violations = append(violations, violation("scheme_must_be_http_https", ErrSchemeDenied, "url"))
	}

	if !inScope(req.Target, ctx.InScopeTargets) {
		violations = append(violations, violation("target_out_of_scope", ErrScopeViolation, "target"))
	}

	if req.Method == MethodPOST {
		violations = append(violations, violation("post_denied_by_default", ErrMethodDenied, "method"))
	}
	if req.Method == MethodPUT || req.Method == MethodPATCH || req.Method == MethodDELETE || req.Method == MethodTRACE || req.Method == MethodCONNECT {
		violations = append(violations, violation("unsafe_method_denied_by_default", ErrMethodDenied, "method"))
	}

	if isProhibitedTool(req.ToolClass, ctx.ProhibitedToolClasses) {
		violations = append(violations, violation("prohibited_tool_class_denied", ErrToolClassNotAllowed, "tool_class"))
	}

	if ctx.ExecutionMode == ExecutionModeExecuteApproved || ctx.ExecutionMode == ExecutionModeRetest {
		if ctx.RateLimitPerMinute <= 0 {
			violations = append(violations, violation("missing_rate_limit", ErrMissingRateLimit, "rate_limit"))
		}
		if ctx.RequestBudget <= 0 {
			violations = append(violations, violation("missing_request_budget", ErrMissingRequestBudget, "request_budget"))
		}
		if ctx.TimeoutSeconds <= 0 {
			violations = append(violations, violation("missing_timeout", ErrMissingTimeout, "timeout"))
		}
		if ctx.MaxResponseSizeBytes <= 0 {
			violations = append(violations, violation("missing_max_response_size", ErrMissingMaxResponseSize, "max_response_size"))
		}
	}

	if len(ctx.StopConditions) == 0 {
		violations = append(violations, violation("missing_stop_conditions", ErrMissingStopConditions, "stop_conditions"))
	}
	if strings.TrimSpace(ctx.EvidencePlanRef) == "" {
		violations = append(violations, violation("missing_evidence_plan", ErrMissingEvidencePlan, "evidence_plan"))
	}

	if hasOutOfScopeRedirect(req.URL, ctx.InScopeTargets) {
		violations = append(violations, violation("out_of_scope_redirect_denied", ErrOutOfScopeRedirect, "url"))
	}

	if len(violations) > 0 {
		return PolicyDecision{Allowed: false, Reason: "denied by policy", Violations: violations}, violations
	}

	return PolicyDecision{Allowed: true, Reason: "allowed by policy"}, nil
}

func (a *OfflineAdapter) Plan(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation) {
	if a.planner != nil {
		req = a.planner.BuildPlan(ctx, req)
	}
	return a.Validate(ctx, req)
}

func (a *OfflineAdapter) Validate(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation) {
	if a.evaluator == nil {
		a.evaluator = DefaultPolicyEvaluator{}
	}
	return a.evaluator.Evaluate(ctx, req)
}

func inScope(target string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

func isProhibitedTool(tc ToolClass, prohibited []ToolClass) bool {
	if tc == ToolClassProhibited {
		return true
	}
	return slices.Contains(prohibited, tc)
}

func hasOutOfScopeRedirect(raw string, allowed []string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	redirect := u.Query().Get("redirect")
	if redirect == "" {
		return false
	}
	ru, err := url.Parse(redirect)
	if err != nil {
		return true
	}
	host := ru.Hostname()
	for _, a := range allowed {
		if strings.EqualFold(a, host) {
			return false
		}
	}
	return true
}

func violation(code string, err error, field string) PolicyViolation {
	return PolicyViolation{Code: code, Reason: err.Error(), Field: field}
}
