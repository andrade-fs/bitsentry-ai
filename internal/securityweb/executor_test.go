package securityweb

import (
	"strings"
	"testing"
	"time"
)

func validApproval(req PlannedRequest) *ExecutionApproval {
	return &ExecutionApproval{
		ApprovalID:            "ap-1",
		ApprovedRequestID:     req.RequestRef,
		ApprovedMethod:        req.Method,
		ApprovedURL:           req.URL,
		ApprovedScopeRef:      "scope-1",
		ApprovedExecutionMode: ExecutionModeExecuteApproved,
		ApprovedToolClass:     req.ToolClass,
		ApprovedIntensity:     IntensityLow,
		ApprovedAt:            time.Now().Add(-1 * time.Minute),
		ApprovedBy:            "reviewer",
		ExpiresAt:             time.Now().Add(10 * time.Minute),
		TTLSeconds:            600,
		ApprovalTextOrHash:    "hash",
		MaxRequests:           1,
		MaxDurationSeconds:    10,
		RateLimitPerMinute:    10,
		StopConditions:        []string{"max-errors"},
	}
}

func TestExecuteApprovedRequiresApproval(t *testing.T) {
	ex := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{})
	res := ex.ExecuteApproved(baseContext(), baseRequest(), nil)
	if res.PolicyDecision != "deny" || !hasViolation(res.Violations, "approval_required") {
		t.Fatalf("expected approval_required deny")
	}
}

func TestApprovalMissingFieldsDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "ok"})
	ex := NewOfflineControlledExecutor(tr, DefaultRedactor{})
	a := validApproval(req)
	a.ApprovalID = ""
	a.ApprovedBy = ""
	a.ApprovalTextOrHash = ""
	a.ApprovedScopeRef = ""
	a.ApprovedExecutionMode = ""
	a.ApprovedToolClass = ""
	a.ApprovedIntensity = ""
	res := ex.ExecuteApproved(ctx, req, a)
	for _, code := range []string{"approval_missing", "approval_actor_missing", "approval_proof_missing", "approval_scope_missing", "approval_execution_mode_missing", "approval_tool_class_missing", "approval_intensity_missing"} {
		if !hasViolation(res.Violations, code) {
			t.Fatalf("expected %s", code)
		}
	}
}

func TestExecuteApprovedMismatchsDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "ok"})
	ex := NewOfflineControlledExecutor(tr, DefaultRedactor{})

	a1 := validApproval(req)
	a1.ApprovedRequestID = "other"
	if !hasViolation(ex.ExecuteApproved(ctx, req, a1).Violations, "approval_request_id_mismatch") {
		t.Fatalf("expected approval_request_id_mismatch")
	}

	a2 := validApproval(req)
	a2.ApprovedMethod = MethodHEAD
	if !hasViolation(ex.ExecuteApproved(ctx, req, a2).Violations, "approval_method_mismatch") {
		t.Fatalf("expected approval_method_mismatch")
	}

	a3 := validApproval(req)
	a3.ApprovedURL = "https://example.com/other"
	if !hasViolation(ex.ExecuteApproved(ctx, req, a3).Violations, "approval_url_mismatch") {
		t.Fatalf("expected approval_url_mismatch")
	}
}

func TestApprovalContextMismatchesDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.ToolClass = ToolClassSafeProbe
	ctx.Intensity = IntensityMedium
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "ok"})
	ex := NewOfflineControlledExecutor(tr, DefaultRedactor{})

	a := validApproval(req)
	a.ApprovedScopeRef = "scope-other"
	a.ApprovedExecutionMode = ExecutionModeDryRun
	a.ApprovedToolClass = ToolClassManual
	a.ApprovedIntensity = IntensityLow
	res := ex.ExecuteApproved(ctx, req, a)
	for _, code := range []string{"approval_scope_mismatch", "approval_execution_mode_mismatch", "approval_tool_class_mismatch", "approval_intensity_mismatch"} {
		if !hasViolation(res.Violations, code) {
			t.Fatalf("expected %s", code)
		}
	}
}

func TestExecuteApprovedExpiredAndMissingExpiresDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	ex := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{})

	am := validApproval(req)
	am.ExpiresAt = time.Time{}
	if !hasViolation(ex.ExecuteApproved(ctx, req, am).Violations, "approval_missing_expires_at") {
		t.Fatalf("expected approval_missing_expires_at")
	}

	ae := validApproval(req)
	ae.ExpiresAt = time.Now().Add(-1 * time.Minute)
	if !hasViolation(ex.ExecuteApproved(ctx, req, ae).Violations, "approval_expired") {
		t.Fatalf("expected approval_expired")
	}
}

func TestApprovalExceedsContextLimitsDenied(t *testing.T) {
	ctx := baseContext()
	ctx.RateLimitPerMinute = 5
	ctx.RequestBudget = 1
	req := baseRequest()
	ex := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{})
	a := validApproval(req)
	a.RateLimitPerMinute = 10
	a.MaxRequests = 2
	res := ex.ExecuteApproved(ctx, req, a)
	if !hasViolation(res.Violations, "approval_exceeds_context_limits") {
		t.Fatalf("expected approval_exceeds_context_limits")
	}
}

func TestHardRequiredLimitsForExecuteApproved(t *testing.T) {
	req := baseRequest()
	a := validApproval(req)
	ex := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{})
	ctx := baseContext()
	ctx.RequestBudget = 0
	ctx.RateLimitPerMinute = 0
	ctx.TimeoutSeconds = 0
	ctx.MaxResponseSizeBytes = 0
	ctx.MaxPreviewSizeBytes = 0
	ctx.StopConditions = nil
	res := ex.ExecuteApproved(ctx, req, a)
	for _, code := range []string{"missing_request_budget", "missing_rate_limit", "missing_timeout", "missing_max_response_size", "missing_max_preview_size", "missing_stop_conditions"} {
		if !hasViolation(res.Violations, code) {
			t.Fatalf("expected %s", code)
		}
	}
}

func TestValidGETAndHEADSucceedViaFakeTransport(t *testing.T) {
	ctx := baseContext()
	tr := NewFakeTransport()

	getReq := baseRequest()
	tr.Register(getReq.RequestRef, MethodGET, getReq.URL, FakeTransportResponse{StatusCode: 200, FinalURL: getReq.URL, Body: "ok-get"})
	resGet := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, getReq, validApproval(getReq))
	if resGet.PolicyDecision != "allow" || resGet.StatusCode != 200 {
		t.Fatalf("expected GET allow")
	}

	headReq := baseRequest()
	headReq.RequestRef = "req-head"
	headReq.Method = MethodHEAD
	headReq.URL = "https://example.com/head"
	tr.Register(headReq.RequestRef, MethodHEAD, headReq.URL, FakeTransportResponse{StatusCode: 204, FinalURL: headReq.URL, Body: ""})
	resHead := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, headReq, validApproval(headReq))
	if resHead.PolicyDecision != "allow" || resHead.StatusCode != 204 {
		t.Fatalf("expected HEAD allow")
	}
}

func TestPOSTDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.Method = MethodPOST
	approval := validApproval(req)
	res := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{}).ExecuteApproved(ctx, req, approval)
	if !hasViolation(res.Violations, "approval_post_denied") {
		t.Fatalf("expected approval_post_denied")
	}
}

func TestFollowRedirectsFalseAndRedirectGuards(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 302, RedirectObserved: true, RedirectLocation: ""})
	ex := NewOfflineControlledExecutor(tr, DefaultRedactor{})
	resInvalid := ex.ExecuteApproved(ctx, req, validApproval(req))
	if resInvalid.FollowRedirects {
		t.Fatalf("follow_redirects must default false")
	}
	if !hasViolation(resInvalid.Violations, "redirect_location_invalid") {
		t.Fatalf("expected redirect_location_invalid")
	}

	tr2 := NewFakeTransport()
	tr2.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 302, RedirectObserved: true, RedirectLocation: "https://evil.example/path"})
	resOut := NewOfflineControlledExecutor(tr2, DefaultRedactor{}).ExecuteApproved(ctx, req, validApproval(req))
	if !hasViolation(resOut.Violations, "redirect_out_of_scope") {
		t.Fatalf("expected redirect_out_of_scope")
	}
}

func TestBudgetExhaustedDenied(t *testing.T) {
	ctx := baseContext()
	ctx.RequestBudget = 1
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "ok"})
	ex := NewOfflineControlledExecutor(tr, DefaultRedactor{})
	a := validApproval(req)
	first := ex.ExecuteApproved(ctx, req, a)
	if first.PolicyDecision != "allow" {
		t.Fatalf("first request should pass")
	}
	second := ex.ExecuteApproved(ctx, req, a)
	if !hasViolation(second.Violations, "limiter_budget_exceeded") {
		t.Fatalf("expected limiter_budget_exceeded")
	}
}

func TestMaxResponseAndPreviewMetadata(t *testing.T) {
	ctx := baseContext()
	ctx.MaxResponseSizeBytes = 100
	ctx.MaxPreviewSizeBytes = 5
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "123456789"})
	res := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, req, validApproval(req))
	if !res.BodyTruncated || res.BodyPreviewRedacted != "12345" || res.ResponseSize != 9 || res.MaxPreviewSize != 5 {
		t.Fatalf("expected truncated preview with metadata")
	}
}

func TestRedactionCoverageAndMetadataAndEvidenceID(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	req.URL = "https://example.com/path?token=abc&api_key=1"
	approval := validApproval(req)
	approval.ApprovedURL = req.URL
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"authorization": "Bearer top",
			"Cookie":        "sid=1",
			"set-cookie":    "sid=2",
			"X-Test":        "ok",
		},
		Body: "authorization: Bearer secret\naccess_token=abc\n",
	})
	res := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, req, approval)
	if res.EvidenceID == "" || res.ApprovalID == "" || res.ApprovedBy == "" || res.ApprovalExpiresAt.IsZero() {
		t.Fatalf("traceability metadata required")
	}
	for _, h := range []string{"authorization", "Cookie", "set-cookie"} {
		if res.HeadersRedacted[h] != "[REDACTED]" {
			t.Fatalf("expected redacted %s", h)
		}
	}
	if !strings.Contains(res.BodyPreviewRedacted, "[REDACTED]") {
		t.Fatalf("expected body preview redacted")
	}
	if len(res.RedactionsApplied) == 0 {
		t.Fatalf("expected redactions applied metadata")
	}
}

func TestFakeTransportRequestRefDeterministicAndFallback(t *testing.T) {
	tr := NewFakeTransport()
	tr.Register("req-stable", MethodGET, "https://a", FakeTransportResponse{StatusCode: 201})
	tr.Register("", MethodGET, "https://a", FakeTransportResponse{StatusCode: 202})
	r1, _ := tr.Execute(PlannedRequest{RequestRef: "req-stable", Method: MethodGET, URL: "https://a"})
	if r1.StatusCode != 201 {
		t.Fatalf("must prefer request_ref mapping")
	}
	r2, _ := tr.Execute(PlannedRequest{Method: MethodGET, URL: "https://a"})
	if r2.StatusCode != 202 {
		t.Fatalf("must fallback to method+url when request_ref empty")
	}
}
