package securitywebhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"bitsentry-ai/internal/securityweb"
)

func TestExecuteGETWithHTTPTestServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Header().Set("X-Test", "ok")
		_, _ = w.Write([]byte("hello"))
	}))
	defer ts.Close()

	tr, err := New(200*time.Millisecond, 1024)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}

	resp, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodGET, URL: ts.URL})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if resp.Body != "hello" {
		t.Fatalf("body = %q, want hello", resp.Body)
	}
}

func TestExecuteHEADWithHTTPTestServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	tr, _ := New(200*time.Millisecond, 1024)
	resp, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodHEAD, URL: ts.URL})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	if resp.Body != "" {
		t.Fatalf("head body = %q, want empty", resp.Body)
	}
}

func TestRedirectNotFollowedAndLocationCaptured(t *testing.T) {
	loc := "/next"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, loc, http.StatusFound)
	}))
	defer ts.Close()

	tr, _ := New(200*time.Millisecond, 1024)
	resp, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodGET, URL: ts.URL})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !resp.RedirectObserved {
		t.Fatalf("expected redirect observed")
	}
	if resp.RedirectLocation != loc {
		t.Fatalf("location = %q, want %q", resp.RedirectLocation, loc)
	}
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status = %d, want 302", resp.StatusCode)
	}
}

func TestBodyPreviewCappedAndTruncatedFlag(t *testing.T) {
	large := strings.Repeat("a", 30)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(large))
	}))
	defer ts.Close()

	tr, _ := New(200*time.Millisecond, 10)
	resp, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodGET, URL: ts.URL})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(resp.Body) != 10 {
		t.Fatalf("body len = %d, want 10", len(resp.Body))
	}
	if !resp.BodyTruncated {
		t.Fatalf("expected BodyTruncated=true")
	}
	if resp.Body == large {
		t.Fatalf("full body should not be stored by default")
	}
}

func TestTimeoutEnforcedWithSlowServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(80 * time.Millisecond)
		_, _ = w.Write([]byte("slow"))
	}))
	defer ts.Close()

	tr, _ := New(10*time.Millisecond, 1024)
	_, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodGET, URL: ts.URL})
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Fatalf("expected normalized timeout error, got %v", err)
	}
}

func TestTransportDoesNotEnforcePolicy(t *testing.T) {
	seen := ""
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Method
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	tr, _ := New(200*time.Millisecond, 1024)
	_, err := tr.Execute(securityweb.PlannedRequest{
		Method: securityweb.MethodPOST,
		URL:    ts.URL,
		Headers: map[string]string{
			"X-Arbitrary": "1",
		},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if seen != http.MethodPost {
		t.Fatalf("expected POST to be sent; transport should not enforce policy")
	}
}

func TestTimeoutIsMandatory(t *testing.T) {
	_, err := New(0, 10)
	if err == nil {
		t.Fatalf("expected timeout required error")
	}
	if !strings.Contains(err.Error(), "timeout required") {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestTransportErrorsNormalizedSafely(t *testing.T) {
	tr, _ := New(100*time.Millisecond, 64)
	_, err := tr.Execute(securityweb.PlannedRequest{Method: securityweb.MethodGET, URL: "http://127.0.0.1:1"})
	if err == nil {
		t.Fatalf("expected error")
	}
	msg := err.Error()
	if strings.Contains(msg, "127.0.0.1:1") {
		t.Fatalf("error should be normalized and not expose target details: %s", msg)
	}
	if !strings.Contains(msg, fmt.Sprintf("%v", ErrTransport)) {
		t.Fatalf("expected normalized transport error, got: %s", msg)
	}
}

func TestExecutorWithInjectedRealTransportGETAndHEAD(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok-get"))
		case http.MethodHead:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer ts.Close()

	tr, err := New(200*time.Millisecond, 1024)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}
	exec := securityweb.NewOfflineControlledExecutor(tr, securityweb.DefaultRedactor{})

	ctx := assessmentCtx(ts.URL)

	getReq := securityweb.PlannedRequest{RequestRef: "req-get", URL: ts.URL, Method: securityweb.MethodGET, ToolClass: securityweb.ToolClassManual}
	getRes := exec.ExecuteApproved(ctx, getReq, testApproval(ctx, getReq))
	if getRes.PolicyDecision != "allow" || getRes.StatusCode != http.StatusOK {
		t.Fatalf("expected approved GET allow via injected real transport")
	}

	headReq := securityweb.PlannedRequest{RequestRef: "req-head", URL: ts.URL, Method: securityweb.MethodHEAD, ToolClass: securityweb.ToolClassManual}
	headRes := exec.ExecuteApproved(ctx, headReq, testApproval(ctx, headReq))
	if headRes.PolicyDecision != "allow" || headRes.StatusCode != http.StatusNoContent {
		t.Fatalf("expected approved HEAD allow via injected real transport")
	}
}

func assessmentCtx(rawURL string) securityweb.AssessmentSessionContext {
	u, _ := url.Parse(rawURL)
	host := u.Hostname()
	return securityweb.AssessmentSessionContext{
		SessionID:            "s-1",
		ScopeRef:             "scope-1",
		InScopeTargets:       []string{host},
		ExecutionMode:        securityweb.ExecutionModeExecuteApproved,
		Intensity:            securityweb.IntensityLow,
		RateLimitPerMinute:   10,
		RequestBudget:        2,
		TimeoutSeconds:       5,
		MaxResponseSizeBytes: 1024,
		MaxPreviewSizeBytes:  256,
		StopConditions:       []string{"max-errors"},
	}
}

func testApproval(ctx securityweb.AssessmentSessionContext, req securityweb.PlannedRequest) *securityweb.ExecutionApproval {
	return &securityweb.ExecutionApproval{
		ApprovalID:            "ap-1",
		ApprovedRequestID:     req.RequestRef,
		ApprovedMethod:        req.Method,
		ApprovedURL:           req.URL,
		ApprovedScopeRef:      ctx.ScopeRef,
		ApprovedExecutionMode: securityweb.ExecutionModeExecuteApproved,
		ApprovedToolClass:     req.ToolClass,
		ApprovedIntensity:     ctx.Intensity,
		ApprovedAt:            time.Now().Add(-1 * time.Minute),
		ApprovedBy:            "reviewer",
		ExpiresAt:             time.Now().Add(10 * time.Minute),
		TTLSeconds:            600,
		ApprovalTextOrHash:    "hash",
		MaxRequests:           2,
		MaxDurationSeconds:    10,
		RateLimitPerMinute:    10,
		StopConditions:        []string{"max-errors"},
	}
}
