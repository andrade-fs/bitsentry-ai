package securityweb

import (
	"fmt"
	"net/url"
	"strings"
)

type ManualPreflightInput struct {
	RequestRef         string
	Method             string
	URL                string
	ScopeHost          string
	ApprovalToken      string
	TimeoutSeconds     int
	MaxResponseSize    int64
	MaxPreviewSize     int64
	RequestBudget      int
	RateLimitPerMinute int
	StopConditions     []string
}

type ManualPreflightResult struct {
	ExecutionBackendAvailable bool     `json:"execution_backend_available"`
	EntrypointAvailable       bool     `json:"entrypoint_available"`
	RequestRef                string   `json:"request_ref"`
	Method                    string   `json:"method"`
	URL                       string   `json:"url"`
	ScopeHost                 string   `json:"scope_host"`
	ScopeValid                bool     `json:"scope_valid"`
	ApprovalValid             bool     `json:"approval_valid"`
	LimitsComplete            bool     `json:"limits_complete"`
	WouldExecute              bool     `json:"would_execute"`
	PolicyDecision            string   `json:"policy_decision"`
	Violations                []string `json:"violations"`
	RequiredLimits            []string `json:"required_limits"`
	SafetyNotes               []string `json:"safety_notes"`
	ExactApprovalRequired     bool     `json:"exact_approval_required"`
	NextStep                  string   `json:"next_step"`
}

func ManualPreflight(in ManualPreflightInput) ManualPreflightResult {
	method := strings.ToUpper(strings.TrimSpace(in.Method))
	reqRef := strings.TrimSpace(in.RequestRef)
	rawURL := strings.TrimSpace(in.URL)
	scopeHost := strings.ToLower(strings.TrimSpace(in.ScopeHost))
	token := strings.TrimSpace(in.ApprovalToken)

	violations := []string{}
	requiredLimits := []string{
		"timeout_seconds > 0",
		"max_response_size_bytes > 0",
		"max_preview_size_bytes > 0",
		"request_budget == 1",
		"rate_limit_per_minute > 0",
		"stop_conditions not empty",
	}

	if reqRef == "" {
		violations = append(violations, "request_ref_required")
	}
	if method == "" {
		violations = append(violations, "method_required")
	}
	if rawURL == "" {
		violations = append(violations, "url_required")
	}
	if scopeHost == "" {
		violations = append(violations, "missing_scope_host")
	}
	if token == "" {
		violations = append(violations, "approval_token_required")
	}

	if method != string(MethodHEAD) {
		violations = append(violations, "method_not_allowed_only_HEAD")
	}

	scopeValid := false
	if rawURL != "" {
		u, err := url.Parse(rawURL)
		if err != nil || strings.TrimSpace(u.Scheme) == "" || strings.TrimSpace(u.Host) == "" {
			violations = append(violations, "url_invalid")
		} else if strings.TrimSpace(u.Path) != "/" {
			violations = append(violations, "path_not_allowed_only_root")
		} else {
			urlHost := strings.ToLower(strings.TrimSpace(u.Hostname()))
			scopeValid = scopeHost != "" && urlHost == scopeHost
			if !scopeValid {
				violations = append(violations, "scope_host_mismatch")
			}
		}
	}

	if in.TimeoutSeconds <= 0 {
		violations = append(violations, "missing_timeout_seconds")
	}
	if in.MaxResponseSize <= 0 {
		violations = append(violations, "missing_max_response_size_bytes")
	}
	if in.MaxPreviewSize <= 0 {
		violations = append(violations, "missing_max_preview_size_bytes")
	}
	if in.RequestBudget != 1 {
		violations = append(violations, "request_budget_must_be_1")
	}
	if in.RateLimitPerMinute <= 0 {
		violations = append(violations, "missing_rate_limit_per_minute")
	}
	if len(in.StopConditions) == 0 {
		violations = append(violations, "missing_stop_conditions")
	}

	expectedToken := fmt.Sprintf("APPROVE request_ref=%s method=%s url=%s", reqRef, method, rawURL)
	approvalValid := token == expectedToken
	if !approvalValid {
		violations = append(violations, "approval_token_mismatch")
	}

	limitsComplete := in.TimeoutSeconds > 0 && in.MaxResponseSize > 0 && in.MaxPreviewSize > 0 && in.RequestBudget == 1 && in.RateLimitPerMinute > 0 && len(in.StopConditions) > 0
	decision := "allow"
	if len(violations) > 0 {
		decision = "deny"
	}

	return ManualPreflightResult{
		ExecutionBackendAvailable: true,
		EntrypointAvailable:       true,
		RequestRef:                reqRef,
		Method:                    method,
		URL:                       rawURL,
		ScopeHost:                 scopeHost,
		ScopeValid:                scopeValid,
		ApprovalValid:             approvalValid,
		LimitsComplete:            limitsComplete,
		WouldExecute:              false,
		PolicyDecision:            decision,
		Violations:                violations,
		RequiredLimits:            requiredLimits,
		SafetyNotes: []string{
			"preflight_only_no_transport_invocation",
			"no_network_requests_in_7_18A",
			"exact_approval_required",
			"no_redirects_followed",
			"no_auth_or_cookies",
			"no_POST",
		},
		ExactApprovalRequired: true,
		NextStep:              "Phase 7.18B manual execute design or one-request robots gate preflight",
	}
}
