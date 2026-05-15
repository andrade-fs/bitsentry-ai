package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestSecuritywebManualPreflight_HumanOutputIncludesEntrypointAvailable(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-preflight",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://bitsentry.xyz/",
		"--scope-host", "bitsentry.xyz",
		"--approval", "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/",
		"--timeout-seconds", "10",
		"--max-response-size-bytes", "4096",
		"--max-preview-size-bytes", "64",
		"--request-budget", "1",
		"--rate-limit-per-minute", "1",
		"--stop-condition", "timeout",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute manual-preflight: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "entrypoint_available: true") {
		t.Fatalf("expected entrypoint_available in output, got:\n%s", got)
	}
	if !strings.Contains(got, "would_execute: false") {
		t.Fatalf("expected would_execute false in output, got:\n%s", got)
	}
	if !strings.Contains(got, "scope_host: bitsentry.xyz") {
		t.Fatalf("expected scope_host in output, got:\n%s", got)
	}
	if !strings.Contains(got, "scope_valid: true") {
		t.Fatalf("expected scope_valid true in output, got:\n%s", got)
	}
}

func TestSecuritywebManualPreflight_JSONIncludesScopeAndWouldExecuteFalse(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-preflight",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://bitsentry.xyz/",
		"--scope-host", "bitsentry.xyz",
		"--approval", "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/",
		"--timeout-seconds", "10",
		"--max-response-size-bytes", "4096",
		"--max-preview-size-bytes", "64",
		"--request-budget", "1",
		"--rate-limit-per-minute", "1",
		"--stop-condition", "timeout",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute manual-preflight --json: %v", err)
	}
	var payload struct {
		EntrypointAvailable bool     `json:"entrypoint_available"`
		WouldExecute        bool     `json:"would_execute"`
		ScopeHost           string   `json:"scope_host"`
		ScopeValid          bool     `json:"scope_valid"`
		PolicyDecision      string   `json:"policy_decision"`
		Violations          []string `json:"violations"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if !payload.EntrypointAvailable {
		t.Fatalf("expected entrypoint_available=true")
	}
	if payload.WouldExecute {
		t.Fatalf("expected would_execute=false")
	}
	if payload.ScopeHost != "bitsentry.xyz" || !payload.ScopeValid {
		t.Fatalf("expected scope_host/scope_valid, got %+v", payload)
	}
	if payload.PolicyDecision != "allow" || len(payload.Violations) != 0 {
		t.Fatalf("expected allow with no violations, got %+v", payload)
	}
}

func TestSecuritywebManualPreflight_JSONMissingScopeHostDenied(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-preflight",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://bitsentry.xyz/",
		"--approval", "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/",
		"--timeout-seconds", "10",
		"--max-response-size-bytes", "4096",
		"--max-preview-size-bytes", "64",
		"--request-budget", "1",
		"--rate-limit-per-minute", "1",
		"--stop-condition", "timeout",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute manual-preflight --json: %v", err)
	}
	var payload struct {
		PolicyDecision string   `json:"policy_decision"`
		Violations     []string `json:"violations"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if payload.PolicyDecision != "deny" || !hasViolation(payload.Violations, "missing_scope_host") {
		t.Fatalf("expected deny + missing_scope_host, got %+v", payload)
	}
}

func TestSecuritywebManualPreflight_JSONScopeHostMismatchDenied(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{
		"manual-preflight",
		"--request-ref", "first-owned-head-root",
		"--method", "HEAD",
		"--url", "https://bitsentry.xyz/",
		"--scope-host", "example.com",
		"--approval", "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/",
		"--timeout-seconds", "10",
		"--max-response-size-bytes", "4096",
		"--max-preview-size-bytes", "64",
		"--request-budget", "1",
		"--rate-limit-per-minute", "1",
		"--stop-condition", "timeout",
		"--json",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute manual-preflight --json: %v", err)
	}
	var payload struct {
		PolicyDecision string   `json:"policy_decision"`
		Violations     []string `json:"violations"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, out.String())
	}
	if payload.PolicyDecision != "deny" || !hasViolation(payload.Violations, "scope_host_mismatch") {
		t.Fatalf("expected deny + scope_host_mismatch, got %+v", payload)
	}
}

func TestSecuritywebNoExecutePathExists(t *testing.T) {
	cmd := newSecuritywebCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"execute"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected unknown command error for execute path")
	}
	if !strings.Contains(out.String(), "unknown command") {
		t.Fatalf("expected unknown command output, got:\n%s", out.String())
	}
}

func hasViolation(vs []string, target string) bool {
	for _, v := range vs {
		if v == target {
			return true
		}
	}
	return false
}
