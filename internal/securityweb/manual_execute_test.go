package securityweb

import (
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"
)

type countingTransportFactory struct{ calls int }

func (c *countingTransportFactory) factory(timeout time.Duration, maxBodyBytes int64) (HTTPTransport, error) {
	c.calls++
	return &fakeDenyTransport{}, nil
}

type fakeDenyTransport struct{}

func (f *fakeDenyTransport) Execute(req PlannedRequest) (FakeTransportResponse, error) {
	return FakeTransportResponse{StatusCode: 200, FinalURL: req.URL, Headers: map[string]string{}}, nil
}

type fakeAllowTransport struct{}

func (f *fakeAllowTransport) Execute(req PlannedRequest) (FakeTransportResponse, error) {
	return FakeTransportResponse{
		StatusCode: 200,
		FinalURL:   req.URL,
		Headers: map[string]string{
			"Content-Security-Policy": "default-src 'self'",
			"Referrer-Policy":         "strict-origin-when-cross-origin",
			"X-Content-Type-Options":  "nosniff",
			"X-Frame-Options":         "DENY",
		},
	}, nil
}

func validExecuteInput(url string) ManualExecuteInput {
	return ManualExecuteInput{
		SessionID:          "sess-1",
		RequestRef:         "first-owned-head-root",
		Method:             "HEAD",
		URL:                url,
		ScopeHost:          hostOnly(url),
		ApprovalToken:      "APPROVE request_ref=first-owned-head-root method=HEAD url=" + url,
		ConfirmExecute:     true,
		TimeoutSeconds:     10,
		MaxResponseSize:    4096,
		MaxPreviewSize:     64,
		RequestBudget:      1,
		RateLimitPerMinute: 1,
		StopConditions:     []string{"timeout"},
		Now:                time.Now,
	}
}

func hostOnly(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err == nil {
		if h := strings.TrimSpace(u.Hostname()); h != "" {
			return h
		}
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(raw, "https://"), "http://"))
}

func TestManualExecute_ValidApprovalExecutesOneHEADHTTptest(t *testing.T) {
	in := validExecuteInput("https://example.com/")
	in.TransportFactory = func(timeout time.Duration, maxBodyBytes int64) (HTTPTransport, error) {
		return &fakeAllowTransport{}, nil
	}
	res := ExecuteManualHead(in)
	if !res.Executed || (res.State != StateCompletedWithCandidates && res.State != StateCompletedNoCandidates) {
		t.Fatalf("expected executed completion state, got %+v", res)
	}
	if res.PassiveHeaderCheck == nil {
		t.Fatalf("expected passive header check on executed result")
	}
}

func TestManualExecute_MissingConfirmDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ConfirmExecute = false
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StateApprovalDenied || res.Executed {
		t.Fatalf("expected approval denied without execution")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_GenericApprovalDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ApprovalToken = "approve all"
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StateApprovalDenied || res.Executed {
		t.Fatalf("expected approval denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_RequestRefMismatchDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ApprovalToken = "APPROVE request_ref=other method=HEAD url=https://example.com/"
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StateApprovalDenied {
		t.Fatalf("expected approval denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_MethodMismatchDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=GET url=https://example.com/"
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StateApprovalDenied {
		t.Fatalf("expected approval denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_URLMismatchDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=HEAD url=https://other.com/"
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StateApprovalDenied {
		t.Fatalf("expected approval denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_MissingMaxPreviewPreflightDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.MaxPreviewSize = 0
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StatePreflightDenied {
		t.Fatalf("expected preflight denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_NonHEADPreflightDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.Method = "GET"
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=GET url=https://example.com/"
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StatePreflightDenied {
		t.Fatalf("expected preflight denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_NonRootPreflightDeniedNoTransport(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/path")
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StatePreflightDenied {
		t.Fatalf("expected preflight denied")
	}
	if cf.calls != 0 {
		t.Fatalf("transport factory should not be called")
	}
}

func TestManualExecute_TransportErrorState(t *testing.T) {
	in := validExecuteInput("https://example.com/")
	in.TransportFactory = func(timeout time.Duration, maxBodyBytes int64) (HTTPTransport, error) {
		return nil, errors.New("boom")
	}
	res := ExecuteManualHead(in)
	if res.State != StateTransportError || res.Executed {
		t.Fatalf("expected transport error non executed, got %+v", res)
	}
}

func TestManualExecute_ExecutedOnlyWhenExecutionResultExists(t *testing.T) {
	in := validExecuteInput("https://example.com/")
	in.TransportFactory = func(timeout time.Duration, maxBodyBytes int64) (HTTPTransport, error) {
		return &fakeDenyTransport{}, nil
	}
	res := ExecuteManualHead(in)
	if !res.Executed {
		t.Fatalf("expected executed true on allow result")
	}
}

func TestManualExecute_PassiveCheckOnlyOnExecuted(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.ConfirmExecute = false
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.PassiveHeaderCheck != nil {
		t.Fatalf("expected no passive check in deny path")
	}
}

func TestManualExecute_DeniedNoPassiveCheck(t *testing.T) {
	cf := &countingTransportFactory{}
	in := validExecuteInput("https://example.com/")
	in.RequestBudget = 2
	in.TransportFactory = cf.factory
	res := ExecuteManualHead(in)
	if res.State != StatePreflightDenied {
		t.Fatalf("expected preflight denied")
	}
	if res.PassiveHeaderCheck != nil || len(res.CandidateFindings) > 0 {
		t.Fatalf("denied path must not create passive findings")
	}
}
