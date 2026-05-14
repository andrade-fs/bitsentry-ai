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
}

func TestSecuritywebManualPreflight_JSONIncludesEntrypointAndWouldExecuteFalse(t *testing.T) {
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
		EntrypointAvailable bool `json:"entrypoint_available"`
		WouldExecute        bool `json:"would_execute"`
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
