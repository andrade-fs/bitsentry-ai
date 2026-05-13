package securityweb

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type OfflineControlledExecutor struct {
	transport *FakeTransport
	redactor  Redactor
	now       func() time.Time

	mu       sync.Mutex
	inFlight bool
}

func NewOfflineControlledExecutor(transport *FakeTransport, redactor Redactor) *OfflineControlledExecutor {
	return &OfflineControlledExecutor{transport: transport, redactor: redactor, now: time.Now}
}

func (e *OfflineControlledExecutor) ExecuteApproved(ctx AssessmentSessionContext, req PlannedRequest, approval *ExecutionApproval) ExecutionResult {
	result := ExecutionResult{
		RequestID:       req.RequestRef,
		EvidenceID:      "WEB-EV-" + req.RequestRef,
		Method:          req.Method,
		URL:             req.URL,
		MaxPreviewSize:  ctx.MaxResponseSizeBytes,
		FollowRedirects: false,
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

	if approval == nil {
		return denyResult(result, violation("approval_required", ErrApprovalRequired, "approval"))
	}
	if approval.ExpiresAt.IsZero() {
		return denyResult(result, violation("approval_missing_expires_at", ErrApprovalMissingExpires, "approval.expires_at"))
	}
	if approval.ExpiresAt.Before(e.now()) {
		return denyResult(result, violation("approval_expired", ErrApprovalExpired, "approval.expires_at"))
	}
	if approval.ApprovedRequestID != req.RequestRef {
		return denyResult(result, violation("approval_request_id_mismatch", ErrApprovalRequestMismatch, "approval.approved_request_id"))
	}
	if approval.ApprovedMethod != req.Method {
		return denyResult(result, violation("approval_method_mismatch", ErrApprovalMethodMismatch, "approval.approved_method"))
	}
	if approval.ApprovedURL != req.URL {
		return denyResult(result, violation("approval_url_mismatch", ErrApprovalURLMismatch, "approval.approved_url"))
	}
	if req.Method == MethodPOST {
		return denyResult(result, violation("approval_post_denied", ErrMethodDenied, "method"))
	}

	resp, err := e.transport.Execute(req)
	if err != nil {
		return denyResult(result, PolicyViolation{Code: "approval_transport_missing_fake_response", Reason: err.Error(), Field: "transport"})
	}

	if resp.RedirectObserved {
		if !inScope(hostnameOf(resp.RedirectLocation), ctx.InScopeTargets) {
			result.RedirectObserved = true
			result.RedirectLocation = resp.RedirectLocation
			return denyResult(result, violation("redirect_out_of_scope_requires_new_approval", ErrOutOfScopeRedirect, "redirect_location"))
		}
	}

	if e.redactor == nil {
		e.redactor = DefaultRedactor{}
	}
	result.StatusCode = resp.StatusCode
	result.FinalURL = firstNonEmpty(resp.FinalURL, req.URL)
	result.RedirectObserved = resp.RedirectObserved
	result.RedirectLocation = resp.RedirectLocation
	result.HeadersRedacted = e.redactor.RedactHeaders(resp.Headers)
	result.ResponseSize = int64(len(resp.Body))
	result.BodyPreviewRedacted, result.BodyTruncated = buildPreview(resp.Body, ctx.MaxResponseSizeBytes)
	result.BodyPreviewRedacted = redactBodyPreview(result.BodyPreviewRedacted)
	result.SensitiveDataRedacted = true
	result.PolicyDecision = "allow"
	return result
}

func denyResult(result ExecutionResult, violations ...PolicyViolation) ExecutionResult {
	result.PolicyDecision = "deny"
	result.Violations = append(result.Violations, violations...)
	result.StatusCode = 0
	result.BodyPreviewRedacted = ""
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

func redactBodyPreview(s string) string {
	out := s
	for _, key := range []string{"Authorization:", "Cookie:", "Set-Cookie:", "Bearer ", "token=", "access_token=", "api_key=", "password=", "secret="} {
		if strings.Contains(strings.ToLower(out), strings.ToLower(key)) {
			out = "[REDACTED]"
			break
		}
	}
	return out
}

func hostnameOf(raw string) string {
	u, err := parseURL(raw)
	if err != nil {
		return ""
	}
	return u
}

func parseURL(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("empty url")
	}
	idx := strings.Index(trimmed, "://")
	if idx < 0 {
		return "", fmt.Errorf("invalid url")
	}
	hostPath := trimmed[idx+3:]
	slash := strings.Index(hostPath, "/")
	if slash < 0 {
		return hostPath, nil
	}
	return hostPath[:slash], nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
