package securityweb

import "testing"

func validPreflightInput() ManualPreflightInput {
	return ManualPreflightInput{
		RequestRef:         "first-owned-head-root",
		Method:             "HEAD",
		URL:                "https://bitsentry.xyz/",
		ScopeHost:          "bitsentry.xyz",
		ApprovalToken:      "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/",
		TimeoutSeconds:     10,
		MaxResponseSize:    4096,
		MaxPreviewSize:     64,
		RequestBudget:      1,
		RateLimitPerMinute: 1,
		StopConditions:     []string{"timeout"},
	}
}

func hasViolationCode(vs []string, target string) bool {
	for _, v := range vs {
		if v == target {
			return true
		}
	}
	return false
}

func TestManualPreflight_ExactApprovalPass(t *testing.T) {
	res := ManualPreflight(validPreflightInput())
	if res.PolicyDecision != "allow" || !res.ApprovalValid || !res.LimitsComplete || !res.ScopeValid {
		t.Fatalf("expected allow/valid/limits complete/scope valid, got %+v", res)
	}
	if !res.ExecutionBackendAvailable || !res.EntrypointAvailable || !res.ExactApprovalRequired {
		t.Fatalf("expected backend/entrypoint/exact approval flags true")
	}
	if res.WouldExecute {
		t.Fatalf("preflight must never execute")
	}
}

func TestManualPreflight_GenericApprovalRejected(t *testing.T) {
	in := validPreflightInput()
	in.ApprovalToken = "approve all"
	res := ManualPreflight(in)
	if res.PolicyDecision != "deny" || !hasViolationCode(res.Violations, "approval_token_mismatch") {
		t.Fatalf("expected approval mismatch deny, got %+v", res)
	}
}

func TestManualPreflight_RequestRefMismatchRejected(t *testing.T) {
	in := validPreflightInput()
	in.ApprovalToken = "APPROVE request_ref=other method=HEAD url=https://bitsentry.xyz/"
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "approval_token_mismatch") {
		t.Fatalf("expected approval mismatch violation")
	}
}

func TestManualPreflight_MethodMismatchRejected(t *testing.T) {
	in := validPreflightInput()
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=GET url=https://bitsentry.xyz/"
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "approval_token_mismatch") {
		t.Fatalf("expected approval mismatch violation")
	}
}

func TestManualPreflight_URLMismatchRejected(t *testing.T) {
	in := validPreflightInput()
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/other"
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "approval_token_mismatch") {
		t.Fatalf("expected approval mismatch violation")
	}
}

func TestManualPreflight_MissingMaxPreviewRejected(t *testing.T) {
	in := validPreflightInput()
	in.MaxPreviewSize = 0
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "missing_max_preview_size_bytes") {
		t.Fatalf("expected missing max preview violation")
	}
}

func TestManualPreflight_NonHEADRejected(t *testing.T) {
	in := validPreflightInput()
	in.Method = "GET"
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=GET url=https://bitsentry.xyz/"
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "method_not_allowed_only_HEAD") {
		t.Fatalf("expected non-head violation")
	}
}

func TestManualPreflight_NonRootPathRejected(t *testing.T) {
	in := validPreflightInput()
	in.URL = "https://bitsentry.xyz/path"
	in.ApprovalToken = "APPROVE request_ref=first-owned-head-root method=HEAD url=https://bitsentry.xyz/path"
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "path_not_allowed_only_root") {
		t.Fatalf("expected non-root path violation")
	}
}

func TestManualPreflight_RequestBudgetMustBeOne(t *testing.T) {
	in := validPreflightInput()
	in.RequestBudget = 2
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "request_budget_must_be_1") {
		t.Fatalf("expected request budget violation")
	}
}

func TestManualPreflight_MissingStopConditionRejected(t *testing.T) {
	in := validPreflightInput()
	in.StopConditions = nil
	res := ManualPreflight(in)
	if !hasViolationCode(res.Violations, "missing_stop_conditions") {
		t.Fatalf("expected missing stop conditions violation")
	}
}

func TestManualPreflight_WouldExecuteAlwaysFalse(t *testing.T) {
	res := ManualPreflight(validPreflightInput())
	if res.WouldExecute {
		t.Fatalf("expected would_execute false")
	}
}


func TestManualPreflight_MissingScopeHostRejected(t *testing.T) {
	in := validPreflightInput()
	in.ScopeHost = ""
	res := ManualPreflight(in)
	if res.PolicyDecision != "deny" || !hasViolationCode(res.Violations, "missing_scope_host") {
		t.Fatalf("expected missing scope host deny, got %+v", res)
	}
}

func TestManualPreflight_ScopeHostMismatchRejected(t *testing.T) {
	in := validPreflightInput()
	in.ScopeHost = "example.com"
	res := ManualPreflight(in)
	if res.PolicyDecision != "deny" || !hasViolationCode(res.Violations, "scope_host_mismatch") {
		t.Fatalf("expected scope host mismatch deny, got %+v", res)
	}
}

func TestManualPreflight_ScopeHostMatchAllows(t *testing.T) {
	in := validPreflightInput()
	in.ScopeHost = "  BITSENTRY.XYZ  "
	res := ManualPreflight(in)
	if res.PolicyDecision != "allow" || !res.ScopeValid {
		t.Fatalf("expected allow with valid scope, got %+v", res)
	}
	if res.ScopeHost != "bitsentry.xyz" {
		t.Fatalf("expected normalized scope host, got %q", res.ScopeHost)
	}
}
