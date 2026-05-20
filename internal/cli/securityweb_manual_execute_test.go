package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSecuritywebManualExecuteExists(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"manual-execute", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("manual-execute should exist: %v", err)
	}
}

func TestSecuritywebManualExecuteMissingConfirmDenied(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-execute",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://example.com/",
		"--scope-host", "example.com",
		"--approval", "APPROVE request_ref=first-owned-head-root method=HEAD url=https://example.com/",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var payload struct {
		State    string `json:"state"`
		Executed bool   `json:"executed"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload.State != "APPROVAL_DENIED" || payload.Executed {
		t.Fatalf("expected approval denied not executed, got %+v", payload)
	}
}

func TestSecuritywebManualExecuteJSONHTTptestExecution(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()
	u, _ := url.Parse(ts.URL)
	host := u.Hostname()

	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	approval := "APPROVE request_ref=first-owned-head-root method=HEAD url=" + ts.URL + "/"
	cmd.SetArgs([]string{
		"manual-execute",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", ts.URL + "/",
		"--scope-host", host,
		"--approval", approval,
		"--confirm-execute",
		"--timeout-seconds", "10",
		"--max-response-size-bytes", "4096",
		"--max-preview-size-bytes", "64",
		"--request-budget", "1",
		"--rate-limit-per-minute", "1",
		"--stop-condition", "timeout",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var payload struct {
		Executed bool   `json:"executed"`
		State    string `json:"state"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v\n%s", err, out.String())
	}
	if !payload.Executed {
		t.Fatalf("expected executed=true on httptest run")
	}
	if payload.State == "" {
		t.Fatalf("expected state")
	}
}

func TestSecuritywebManualExecuteHumanOutputIncludesStateAndEvidence(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-execute",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://example.com/",
		"--scope-host", "example.com",
		"--approval", "approve all",
		"--confirm-execute",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "state:") || !strings.Contains(got, "evidence_id:") {
		t.Fatalf("expected state and evidence in output: %s", got)
	}
}

func TestSecuritywebManualExecuteDenyOutputExecutedFalse(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-execute",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://example.com/",
		"--scope-host", "example.com",
		"--approval", "approve all",
		"--confirm-execute",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var payload struct {
		Executed bool `json:"executed"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("json: %v", err)
	}
	if payload.Executed {
		t.Fatalf("expected executed=false on deny")
	}
}
