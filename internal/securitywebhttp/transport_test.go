package securitywebhttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
