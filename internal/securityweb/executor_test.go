package securityweb

import (
	"strings"
	"testing"
	"time"
)

func validApproval(req PlannedRequest) *ExecutionApproval {
	return &ExecutionApproval{
		ApprovalID:         "ap-1",
		ApprovedRequestID:  req.RequestRef,
		ApprovedMethod:     req.Method,
		ApprovedURL:        req.URL,
		ApprovedScopeRef:   "scope-1",
		ApprovedAt:         time.Now().Add(-1 * time.Minute),
		ApprovedBy:         "reviewer",
		ExpiresAt:          time.Now().Add(10 * time.Minute),
		TTLSeconds:         600,
		ApprovalTextOrHash: "hash",
		MaxRequests:        1,
		MaxDurationSeconds: 10,
		RateLimitPerMinute: 10,
		StopConditions:     []string{"max-errors"},
	}
}

func TestExecuteApprovedRequiresApproval(t *testing.T) {
	ex := NewOfflineControlledExecutor(NewFakeTransport(), DefaultRedactor{})
	res := ex.ExecuteApproved(baseContext(), baseRequest(), nil)
	if res.PolicyDecision != "deny" || !hasViolation(res.Violations, "approval_required") {
		t.Fatalf("expected approval_required deny")
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

func TestFollowRedirectsFalseAndOutOfScopeRedirectDenied(t *testing.T) {
	ctx := baseContext()
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 302, RedirectObserved: true, RedirectLocation: "https://evil.example/path", Body: "redirect"})
	res := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, req, validApproval(req))
	if res.FollowRedirects {
		t.Fatalf("follow_redirects must default false")
	}
	if !hasViolation(res.Violations, "redirect_out_of_scope_requires_new_approval") {
		t.Fatalf("expected out-of-scope redirect denial")
	}
}

func TestMaxResponseSizeTruncatesBodyPreview(t *testing.T) {
	ctx := baseContext()
	ctx.MaxResponseSizeBytes = 5
	req := baseRequest()
	tr := NewFakeTransport()
	tr.Register(req.RequestRef, req.Method, req.URL, FakeTransportResponse{StatusCode: 200, Body: "123456789"})
	res := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, req, validApproval(req))
	if !res.BodyTruncated || res.BodyPreviewRedacted != "12345" || res.ResponseSize != 9 {
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
			"Authorization": "Bearer top",
			"Cookie":        "sid=1",
			"Set-Cookie":    "sid=2",
			"X-Test":        "ok",
		},
		Body: "Authorization: Bearer secret\nCookie: sid=1\n",
	})
	res := NewOfflineControlledExecutor(tr, DefaultRedactor{}).ExecuteApproved(ctx, req, approval)
	if res.EvidenceID == "" {
		t.Fatalf("evidence id required")
	}
	for _, h := range []string{"Authorization", "Cookie", "Set-Cookie"} {
		if res.HeadersRedacted[h] != "[REDACTED]" {
			t.Fatalf("expected redacted %s", h)
		}
	}
	if !strings.Contains(res.BodyPreviewRedacted, "[REDACTED]") {
		t.Fatalf("expected body preview redacted")
	}
	if res.MaxPreviewSize != ctx.MaxResponseSizeBytes {
		t.Fatalf("max preview metadata mismatch")
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
