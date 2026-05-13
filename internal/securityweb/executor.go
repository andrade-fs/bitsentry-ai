package securityweb

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
)

type OfflineControlledExecutor struct {
	transport HTTPTransport
	redactor  Redactor
	now       func() time.Time

	mu          sync.Mutex
	inFlight    bool
	usedByScope map[string]int
}

func NewOfflineControlledExecutor(transport *FakeTransport, redactor Redactor) *OfflineControlledExecutor {
	return &OfflineControlledExecutor{
		transport:   transport,
		redactor:    redactor,
		now:         time.Now,
		usedByScope: map[string]int{},
	}
}

func (e *OfflineControlledExecutor) ExecuteApproved(ctx AssessmentSessionContext, req PlannedRequest, approval *ExecutionApproval) ExecutionResult {
	result := ExecutionResult{
		RequestID:      req.RequestRef,
		EvidenceID:     "WEB-EV-" + req.RequestRef,
		Method:         req.Method,
		URL:            req.URL,
		MaxPreviewSize: ctx.MaxPreviewSizeBytes,
		FollowRedirects:false,
	}

	e.mu.Lock()
	if e.inFlight {
		e.mu.Unlock()
		return denyResult(result, violation("limiter_one_request_at_a_time", ErrLimiterConcurrent, "limiter"))
	}
	e.inFlight = true
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.inFlight = false
		e.mu.Unlock()
	}()

	if v := e.validateRequiredLimits(ctx); len(v) > 0 {
		return denyResult(result, v...)
	}
	if v := e.validateApproval(ctx, req, approval); len(v) > 0 {
		return denyResult(result, v...)
	}

	result.ApprovalID = approval.ApprovalID
	result.ApprovedBy = approval.ApprovedBy
	result.ApprovalExpiresAt = approval.ExpiresAt

	if req.Method != MethodGET && req.Method != MethodHEAD {
		return denyResult(result, violation("approval_post_denied", ErrMethodDenied, "method"))
	}

	if e.budgetExceeded(ctx.ScopeRef, effectiveBudget(ctx, approval)) {
		return denyResult(result, violation("limiter_budget_exceeded", ErrLimiterBudgetExceeded, "request_budget"))
	}

	if e.transport == nil {
		return denyResult(result, PolicyViolation{Code: "approval_transport_missing_fake_response", Reason: "fake transport is required", Field: "transport"})
	}
	resp, err := e.transport.Execute(req)
	if err != nil {
		return denyResult(result, PolicyViolation{Code: "approval_transport_missing_fake_response", Reason: err.Error(), Field: "transport"})
	}

	if resp.RedirectObserved {
		result.RedirectObserved = true
		result.RedirectLocation = resp.RedirectLocation
		host, rErr := safeRedirectHost(resp.RedirectLocation)
		if rErr != nil {
			return denyResult(result, violation("redirect_location_invalid", ErrRedirectLocationInvalid, "redirect_location"))
		}
		if !inScope(host, ctx.InScopeTargets) {
			return denyResult(result, violation("redirect_out_of_scope", ErrOutOfScopeRedirect, "redirect_location"))
		}
	}

	if e.redactor == nil {
		e.redactor = DefaultRedactor{}
	}
	result.StatusCode = resp.StatusCode
	result.FinalURL = firstNonEmpty(resp.FinalURL, req.URL)
	result.HeadersRedacted = e.redactor.RedactHeaders(resp.Headers)
	result.ResponseSize = int64(len(resp.Body))
	result.MaxPreviewSize = ctx.MaxPreviewSizeBytes
	preview, trunc := buildPreview(resp.Body, ctx.MaxPreviewSizeBytes)
	redPreview, applied := redactBodyPreviewConservative(preview)
	result.BodyPreviewRedacted = redPreview
	result.BodyTruncated = trunc
	result.SensitiveDataRedacted = true
	result.RedactionsApplied = applied
	result.SafetyNotes = append(result.SafetyNotes, "no full response stored by default")
	result.PolicyDecision = "allow"

	e.consumeBudget(ctx.ScopeRef)
	return result
}

func (e *OfflineControlledExecutor) validateRequiredLimits(ctx AssessmentSessionContext) []PolicyViolation {
	v := make([]PolicyViolation, 0)
	if ctx.RequestBudget <= 0 {
		v = append(v, violation("missing_request_budget", ErrMissingRequestBudget, "request_budget"))
	}
	if ctx.RateLimitPerMinute <= 0 {
		v = append(v, violation("missing_rate_limit", ErrMissingRateLimit, "rate_limit"))
	}
	if ctx.TimeoutSeconds <= 0 {
		v = append(v, violation("missing_timeout", ErrMissingTimeout, "timeout"))
	}
	if ctx.MaxResponseSizeBytes <= 0 {
		v = append(v, violation("missing_max_response_size", ErrMissingMaxResponseSize, "max_response_size"))
	}
	if ctx.MaxPreviewSizeBytes <= 0 {
		v = append(v, violation("missing_max_preview_size", ErrMissingMaxPreviewSize, "max_preview_size"))
	}
	if len(ctx.StopConditions) == 0 {
		v = append(v, violation("missing_stop_conditions", ErrMissingStopConditions, "stop_conditions"))
	}
	return v
}

func (e *OfflineControlledExecutor) validateApproval(ctx AssessmentSessionContext, req PlannedRequest, approval *ExecutionApproval) []PolicyViolation {
	v := make([]PolicyViolation, 0)
	if approval == nil {
		return []PolicyViolation{violation("approval_required", ErrApprovalRequired, "approval")}
	}
	if strings.TrimSpace(approval.ApprovalID) == "" {
		v = append(v, violation("approval_missing", ErrApprovalRequired, "approval_id"))
	}
	if strings.TrimSpace(approval.ApprovedBy) == "" {
		v = append(v, violation("approval_actor_missing", ErrApprovalActorMissing, "approved_by"))
	}
	if strings.TrimSpace(approval.ApprovalTextOrHash) == "" {
		v = append(v, violation("approval_proof_missing", ErrApprovalProofMissing, "approval_text_or_hash"))
	}
	if strings.TrimSpace(approval.ApprovedScopeRef) == "" {
		v = append(v, violation("approval_scope_missing", ErrApprovalScopeMissing, "approved_scope_ref"))
	}
	if strings.TrimSpace(string(approval.ApprovedExecutionMode)) == "" {
		v = append(v, violation("approval_execution_mode_missing", ErrApprovalExecutionModeMissing, "approved_execution_mode"))
	}
	if strings.TrimSpace(string(approval.ApprovedToolClass)) == "" {
		v = append(v, violation("approval_tool_class_missing", ErrApprovalToolClassMissing, "approved_tool_class"))
	}
	if strings.TrimSpace(string(approval.ApprovedIntensity)) == "" {
		v = append(v, violation("approval_intensity_missing", ErrApprovalIntensityMissing, "approved_intensity"))
	}
	if approval.ExpiresAt.IsZero() {
		v = append(v, violation("approval_missing_expires_at", ErrApprovalMissingExpires, "approval.expires_at"))
	}
	if !approval.ExpiresAt.IsZero() && approval.ExpiresAt.Before(e.now()) {
		v = append(v, violation("approval_expired", ErrApprovalExpired, "approval.expires_at"))
	}

	if approval.ApprovedRequestID != req.RequestRef {
		v = append(v, violation("approval_request_id_mismatch", ErrApprovalRequestMismatch, "approval.approved_request_id"))
	}
	if approval.ApprovedMethod != req.Method {
		v = append(v, violation("approval_method_mismatch", ErrApprovalMethodMismatch, "approval.approved_method"))
	}
	if approval.ApprovedURL != req.URL {
		v = append(v, violation("approval_url_mismatch", ErrApprovalURLMismatch, "approval.approved_url"))
	}

	if approval.ApprovedScopeRef != ctx.ScopeRef {
		v = append(v, violation("approval_scope_mismatch", ErrApprovalScopeMismatch, "approval.approved_scope_ref"))
	}
	if approval.ApprovedExecutionMode != ExecutionModeExecuteApproved {
		v = append(v, violation("approval_execution_mode_mismatch", ErrApprovalExecutionModeMismatch, "approval.approved_execution_mode"))
	}
	if approval.ApprovedToolClass != req.ToolClass {
		v = append(v, violation("approval_tool_class_mismatch", ErrApprovalToolClassMismatch, "approval.approved_tool_class"))
	}
	if approval.ApprovedIntensity != ctx.Intensity {
		v = append(v, violation("approval_intensity_mismatch", ErrApprovalIntensityMismatch, "approval.approved_intensity"))
	}

	if approval.RateLimitPerMinute > ctx.RateLimitPerMinute || approval.MaxRequests > ctx.RequestBudget {
		v = append(v, violation("approval_exceeds_context_limits", ErrApprovalExceedsContextLimits, "approval_limits"))
	}
	if len(approval.StopConditions) == 0 {
		v = append(v, violation("missing_stop_conditions", ErrMissingStopConditions, "approval.stop_conditions"))
	}

	return v
}

func (e *OfflineControlledExecutor) budgetExceeded(scope string, effective int) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.usedByScope[scope] >= effective
}

func (e *OfflineControlledExecutor) consumeBudget(scope string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.usedByScope[scope]++
}

func effectiveBudget(ctx AssessmentSessionContext, a *ExecutionApproval) int {
	if a == nil {
		return ctx.RequestBudget
	}
	if a.MaxRequests <= 0 {
		return 0
	}
	if ctx.RequestBudget <= 0 {
		return a.MaxRequests
	}
	if a.MaxRequests < ctx.RequestBudget {
		return a.MaxRequests
	}
	return ctx.RequestBudget
}

func denyResult(result ExecutionResult, violations ...PolicyViolation) ExecutionResult {
	result.PolicyDecision = "deny"
	result.Violations = append(result.Violations, violations...)
	result.StatusCode = 0
	result.BodyPreviewRedacted = ""
	if len(violations) > 0 {
		result.SafetyNotes = append(result.SafetyNotes, violations[0].Code)
	}
	return result
}

func buildPreview(body string, maxBytes int64) (string, bool) {
	if maxBytes <= 0 {
		return "", false
	}
	if int64(len(body)) <= maxBytes {
		return body, false
	}
	return body[:maxBytes], true
}

func redactBodyPreviewConservative(s string) (string, []string) {
	if strings.TrimSpace(s) == "" {
		return "", nil
	}
	redactions := []string{}
	out := s
	patterns := []string{"authorization:", "cookie:", "set-cookie:", "bearer ", "token=", "access_token=", "api_key=", "password=", "secret="}
	for _, p := range patterns {
		if strings.Contains(strings.ToLower(out), p) {
			redactions = append(redactions, p)
		}
	}
	if len(redactions) == 0 {
		return out, nil
	}
	return "[REDACTED]", redactions
}

func safeRedirectHost(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", fmt.Errorf("invalid redirect: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid redirect scheme")
	}
	h := strings.TrimSpace(u.Hostname())
	if h == "" {
		return "", fmt.Errorf("empty redirect host")
	}
	return strings.ToLower(h), nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
