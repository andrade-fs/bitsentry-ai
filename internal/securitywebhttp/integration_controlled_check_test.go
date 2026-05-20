package securitywebhttp_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"bitsentry-ai/internal/securityweb"
	"bitsentry-ai/internal/securitywebhttp"
)

func TestIntegrationControlledCheck_HeadersHappyPathHTTptest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		w.Header().Set("Content-Security-Policy", "default-src 'self'; frame-ancestors 'none'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	plan := mkWebPlan("plan-item-headers", "hyp-headers", securityweb.NextHeadersHardeningReview)
	bridgeCtx := mkBridgeCtx()
	ccp := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, bridgeCtx, "example.com")
	if ccp.ExecutionMode != securityweb.ExecutionModeDryRun || ccp.WouldExecute {
		t.Fatalf("controlled check plan must remain dry_run / would_execute=false")
	}
	req := ccp.PlannedRequests[0]
	req = remapRequestToHTTptest(t, req, ts.URL)

	approval := buildApprovalForRequest(req, execCtx(ts.URL))
	tr, _ := securitywebhttp.New(300*time.Millisecond, 1024)
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})
	res := exec.ExecuteApproved(execCtx(ts.URL), req, approval)

	passive := securityweb.EvaluatePassiveHeaders(securityweb.HeaderCheckInput{ExecutionResult: res, RequestedMethod: securityweb.MethodHEAD})
	if len(passive.CandidateFindings) != 0 {
		t.Fatalf("expected no candidate findings in happy path")
	}
	assertTraceability(t, req, approval, res, passive.EvidenceID)
}

func TestIntegrationControlledCheck_MissingCSPCandidateFinding(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	plan := mkWebPlan("plan-item-headers-missing-csp", "hyp-headers", securityweb.NextHeadersHardeningReview)
	ccp := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := remapRequestToHTTptest(t, ccp.PlannedRequests[0], ts.URL)

	approval := buildApprovalForRequest(req, execCtx(ts.URL))
	tr, _ := securitywebhttp.New(300*time.Millisecond, 1024)
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})
	res := exec.ExecuteApproved(execCtx(ts.URL), req, approval)

	passive := securityweb.EvaluatePassiveHeaders(securityweb.HeaderCheckInput{ExecutionResult: res, RequestedMethod: securityweb.MethodHEAD})
	if !hasFinding(passive, "hf-csp-missing") {
		t.Fatalf("expected conservative missing CSP candidate finding")
	}
	assertTraceability(t, req, approval, res, passive.EvidenceID)
}

func TestIntegrationControlledCheck_RobotsSensitivePathCandidateFinding(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/robots.txt" {
			t.Fatalf("expected /robots.txt path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /admin\n"))
	}))
	defer ts.Close()

	plan := mkWebPlan("plan-item-robots", "hyp-robots", securityweb.NextRobotsExposureReview)
	ccp := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := remapRequestToHTTptest(t, ccp.PlannedRequests[0], ts.URL)

	approval := buildApprovalForRequest(req, execCtx(ts.URL))
	tr, _ := securitywebhttp.New(300*time.Millisecond, 2048)
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})
	res := exec.ExecuteApproved(execCtx(ts.URL), req, approval)

	passive := securityweb.EvaluatePassiveRobots(res)
	if !hasFinding(passive, "robots-sensitive-paths-listed") {
		t.Fatalf("expected conservative robots sensitive-path finding")
	}
	assertTraceability(t, req, approval, res, passive.EvidenceID)
}

func TestIntegrationControlledCheck_InvalidApprovalNoTransportInvocation(t *testing.T) {
	plan := mkWebPlan("plan-item-headers-invalid", "hyp-headers", securityweb.NextHeadersHardeningReview)
	ccp := securityweb.BuildControlledCheckPlanFromWebTestPlan(plan, mkBridgeCtx(), "example.com")
	req := ccp.PlannedRequests[0]

	fake := &fakeCountingTransport{}
	exec := securityweb.NewOfflineControlledExecutor(fake, securityweb.DefaultRedactor{})
	ctx := execCtx("https://example.com")
	approval := buildApprovalForRequest(req, ctx)
	approval.ExpiresAt = time.Now().Add(-1 * time.Minute)

	res := exec.ExecuteApproved(ctx, req, approval)
	if res.PolicyDecision != "deny" {
		t.Fatalf("expected deny for expired approval")
	}
	if fake.calls != 0 {
		t.Fatalf("transport should not be invoked on invalid approval")
	}
}

func mkWebPlan(itemID, hypID string, next securityweb.SuggestedNextCheck) securityweb.WebTestPlan {
	return securityweb.WebTestPlan{
		PlanID: "web-test-plan-static-mvp",
		TestPlanItems: []securityweb.WebTestPlanItem{{
			ItemID:             itemID,
			HypothesisID:       hypID,
			SuggestedNextCheck: next,
			Status:             securityweb.WebTestPlanStatusPlanned,
		}},
	}
}

func mkBridgeCtx() securityweb.AssessmentSessionContext {
	return securityweb.AssessmentSessionContext{
		SessionID:            "sess-bridge",
		AuthorizationRef:     "auth-bridge",
		ScopeRef:             "scope-bridge",
		InScopeTargets:       []string{"example.com"},
		ExecutionMode:        securityweb.ExecutionModePlanningOnly,
		Intensity:            securityweb.IntensityLow,
		RateLimitPerMinute:   10,
		RequestBudget:        5,
		TimeoutSeconds:       10,
		MaxResponseSizeBytes: 4096,
		MaxPreviewSizeBytes:  1024,
		StopConditions:       []string{"user_stop"},
		EvidencePlanRef:      "ev-plan-bridge",
	}
}

func execCtx(rawURL string) securityweb.AssessmentSessionContext {
	u, _ := url.Parse(rawURL)
	host := u.Hostname()
	return securityweb.AssessmentSessionContext{
		SessionID:            "sess-exec",
		AuthorizationRef:     "auth-exec",
		ScopeRef:             "scope-exec",
		InScopeTargets:       []string{host},
		ExecutionMode:        securityweb.ExecutionModeExecuteApproved,
		Intensity:            securityweb.IntensityLow,
		ExplicitApproval:     true,
		RateLimitPerMinute:   20,
		RequestBudget:        3,
		TimeoutSeconds:       10,
		MaxResponseSizeBytes: 4096,
		MaxPreviewSizeBytes:  1024,
		StopConditions:       []string{"user_stop"},
		EvidencePlanRef:      "ev-plan-exec",
	}
}

func buildApprovalForRequest(req securityweb.PlannedRequest, ctx securityweb.AssessmentSessionContext) *securityweb.ExecutionApproval {
	return &securityweb.ExecutionApproval{
		ApprovalID:            "approval-" + req.RequestRef,
		ApprovedRequestID:     req.RequestRef,
		ApprovedMethod:        req.Method,
		ApprovedURL:           req.URL,
		ApprovedScopeRef:      ctx.ScopeRef,
		ApprovedExecutionMode: securityweb.ExecutionModeExecuteApproved,
		ApprovedToolClass:     req.ToolClass,
		ApprovedIntensity:     ctx.Intensity,
		ApprovedAt:            time.Now().Add(-1 * time.Minute),
		ApprovedBy:            "reviewer",
		ExpiresAt:             time.Now().Add(15 * time.Minute),
		TTLSeconds:            900,
		ApprovalTextOrHash:    "approval-proof",
		MaxRequests:           1,
		MaxDurationSeconds:    30,
		RateLimitPerMinute:    ctx.RateLimitPerMinute,
		StopConditions:        ctx.StopConditions,
	}
}

func remapRequestToHTTptest(t *testing.T, req securityweb.PlannedRequest, serverURL string) securityweb.PlannedRequest {
	t.Helper()
	uReq, err := url.Parse(req.URL)
	if err != nil {
		t.Fatalf("parse req url: %v", err)
	}
	uSrv, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}
	uReq.Scheme = uSrv.Scheme
	uReq.Host = uSrv.Host
	req.URL = uReq.String()
	req.Target = uSrv.Hostname()
	return req
}

func assertTraceability(t *testing.T, req securityweb.PlannedRequest, approval *securityweb.ExecutionApproval, res securityweb.ExecutionResult, passiveEvidenceID string) {
	t.Helper()
	if res.RequestID != req.RequestRef {
		t.Fatalf("request ref trace mismatch")
	}
	if res.ApprovalID != approval.ApprovalID {
		t.Fatalf("approval trace mismatch")
	}
	if res.EvidenceID == "" || passiveEvidenceID == "" {
		t.Fatalf("expected non-empty evidence ids")
	}
	if res.EvidenceID != passiveEvidenceID {
		t.Fatalf("passive evidence id must match execution evidence id")
	}
}

func hasFinding(r securityweb.PassiveCheckResult, id string) bool {
	for _, f := range r.CandidateFindings {
		if f.CandidateID == id {
			return true
		}
	}
	return false
}

type fakeCountingTransport struct {
	calls int
}

func (f *fakeCountingTransport) Execute(req securityweb.PlannedRequest) (securityweb.FakeTransportResponse, error) {
	f.calls++
	return securityweb.FakeTransportResponse{StatusCode: 200}, nil
}
