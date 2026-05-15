package securityweb

import (
	"strings"
	"time"
)

type ManualExecutionState string

const (
	StatePreflightDenied             ManualExecutionState = "PREFLIGHT_DENIED"
	StateApprovalDenied              ManualExecutionState = "APPROVAL_DENIED"
	StateExecutionDeniedPreTransport ManualExecutionState = "EXECUTION_DENIED_PRE_TRANSPORT"
	StateTransportError              ManualExecutionState = "TRANSPORT_ERROR"
	StateExecuted                    ManualExecutionState = "EXECUTED"
	StateCompletedWithCandidates     ManualExecutionState = "COMPLETED_WITH_CANDIDATES"
	StateCompletedNoCandidates       ManualExecutionState = "COMPLETED_NO_CANDIDATES"
)

type ManualExecuteInput struct {
	SessionID          string
	RequestRef         string
	Method             string
	URL                string
	ScopeHost          string
	ApprovalToken      string
	ConfirmExecute     bool
	TimeoutSeconds     int
	MaxResponseSize    int64
	MaxPreviewSize     int64
	RequestBudget      int
	RateLimitPerMinute int
	StopConditions     []string
	Now                func() time.Time
	TransportFactory   func(timeout time.Duration, maxBodyBytes int64) (HTTPTransport, error)
}

type ManualExecuteResult struct {
	State                     ManualExecutionState `json:"state"`
	Executed                  bool                 `json:"executed"`
	ExecutionBackendAvailable bool                 `json:"execution_backend_available"`
	EntrypointAvailable       bool                 `json:"entrypoint_available"`
	RequestRef                string               `json:"request_ref"`
	Method                    string               `json:"method"`
	URL                       string               `json:"url"`
	ApprovalID                string               `json:"approval_id"`
	EvidenceID                string               `json:"evidence_id"`
	StatusCode                int                  `json:"status_code"`
	RedirectObserved          bool                 `json:"redirect_observed"`
	RedirectLocation          string               `json:"redirect_location"`
	DurationMS                int64                `json:"duration_ms"`
	BodyTruncated             bool                 `json:"body_truncated"`
	RedactionsApplied         []string             `json:"redactions_applied"`
	SafetyNotes               []string             `json:"safety_notes"`
	Limitations               []string             `json:"limitations"`
	Violations                []string             `json:"violations"`
	PassiveHeaderCheck        *PassiveCheckResult  `json:"passive_header_check,omitempty"`
	CandidateFindings         []CandidateFinding   `json:"candidate_findings,omitempty"`
	PassFail                  string               `json:"pass_fail"`
}

func ExecuteManualHead(in ManualExecuteInput) ManualExecuteResult {
	method := strings.ToUpper(strings.TrimSpace(in.Method))
	res := ManualExecuteResult{
		ExecutionBackendAvailable: true,
		EntrypointAvailable:       true,
		RequestRef:                strings.TrimSpace(in.RequestRef),
		Method:                    method,
		URL:                       strings.TrimSpace(in.URL),
	}

	if !in.ConfirmExecute {
		res.State = StateApprovalDenied
		res.Executed = false
		res.Violations = []string{"confirm_execute_required"}
		res.PassFail = "FAIL"
		return res
	}

	pf := ManualPreflight(ManualPreflightInput{
		RequestRef:         in.RequestRef,
		Method:             method,
		URL:                in.URL,
		ScopeHost:          in.ScopeHost,
		ApprovalToken:      in.ApprovalToken,
		TimeoutSeconds:     in.TimeoutSeconds,
		MaxResponseSize:    in.MaxResponseSize,
		MaxPreviewSize:     in.MaxPreviewSize,
		RequestBudget:      in.RequestBudget,
		RateLimitPerMinute: in.RateLimitPerMinute,
		StopConditions:     in.StopConditions,
	})

	if !pf.ApprovalValid {
		res.State = StateApprovalDenied
		res.Executed = false
		res.Violations = append([]string{}, pf.Violations...)
		res.SafetyNotes = append(res.SafetyNotes, "approval_token_invalid_or_mismatch")
		res.PassFail = "FAIL"
		return res
	}

	if pf.PolicyDecision != "allow" {
		res.State = StatePreflightDenied
		res.Executed = false
		res.Violations = append([]string{}, pf.Violations...)
		res.SafetyNotes = append(res.SafetyNotes, pf.SafetyNotes...)
		res.PassFail = "FAIL"
		return res
	}

	if in.TransportFactory == nil {
		res.State = StateTransportError
		res.Executed = false
		res.Violations = []string{"transport_factory_missing"}
		res.PassFail = "FAIL"
		return res
	}

	transport, err := in.TransportFactory(time.Duration(in.TimeoutSeconds)*time.Second, in.MaxResponseSize)
	if err != nil {
		res.State = StateTransportError
		res.Executed = false
		res.Violations = []string{"transport_init_error"}
		res.SafetyNotes = []string{err.Error()}
		res.PassFail = "FAIL"
		return res
	}

	now := in.Now
	if now == nil {
		now = time.Now
	}
	ctx := AssessmentSessionContext{
		SessionID:            strings.TrimSpace(in.SessionID),
		AuthorizationRef:     "manual-auth-" + strings.TrimSpace(in.RequestRef),
		ScopeRef:             "manual-scope-" + strings.TrimSpace(in.RequestRef),
		InScopeTargets:       []string{strings.TrimSpace(in.ScopeHost)},
		ExecutionMode:        ExecutionModeExecuteApproved,
		Intensity:            IntensityLow,
		ExplicitApproval:     true,
		RateLimitPerMinute:   in.RateLimitPerMinute,
		RequestBudget:        in.RequestBudget,
		TimeoutSeconds:       in.TimeoutSeconds,
		MaxResponseSizeBytes: in.MaxResponseSize,
		MaxPreviewSizeBytes:  in.MaxPreviewSize,
		StopConditions:       append([]string{}, in.StopConditions...),
		EvidencePlanRef:      "manual-evidence-" + strings.TrimSpace(in.RequestRef),
	}
	req := PlannedRequest{RequestRef: in.RequestRef, Target: in.ScopeHost, URL: in.URL, Method: RequestMethod(method), ToolClass: ToolClassSafeProbe, Headers: map[string]string{}}
	approval := &ExecutionApproval{
		ApprovalID:            "approval-" + strings.TrimSpace(in.RequestRef),
		ApprovedRequestID:     strings.TrimSpace(in.RequestRef),
		ApprovedMethod:        RequestMethod(method),
		ApprovedURL:           strings.TrimSpace(in.URL),
		ApprovedScopeRef:      ctx.ScopeRef,
		ApprovedExecutionMode: ExecutionModeExecuteApproved,
		ApprovedToolClass:     req.ToolClass,
		ApprovedIntensity:     ctx.Intensity,
		ApprovedAt:            now().Add(-1 * time.Minute),
		ApprovedBy:            "manual-owner",
		ExpiresAt:             now().Add(10 * time.Minute),
		TTLSeconds:            600,
		ApprovalTextOrHash:    in.ApprovalToken,
		MaxRequests:           1,
		MaxDurationSeconds:    30,
		RateLimitPerMinute:    in.RateLimitPerMinute,
		StopConditions:        append([]string{}, in.StopConditions...),
	}

	exec := NewOfflineControlledExecutor(transport, DefaultRedactor{})
	start := now()
	execRes := exec.ExecuteApproved(ctx, req, approval)
	dur := now().Sub(start).Milliseconds()

	res.ApprovalID = execRes.ApprovalID
	res.EvidenceID = execRes.EvidenceID
	res.StatusCode = execRes.StatusCode
	res.RedirectObserved = execRes.RedirectObserved
	res.RedirectLocation = execRes.RedirectLocation
	res.DurationMS = dur
	res.BodyTruncated = execRes.BodyTruncated
	res.RedactionsApplied = append([]string{}, execRes.RedactionsApplied...)
	res.SafetyNotes = append([]string{}, execRes.SafetyNotes...)

	if execRes.PolicyDecision != "allow" {
		for _, v := range execRes.Violations {
			res.Violations = append(res.Violations, v.Code)
		}
		res.Executed = false
		state := StateExecutionDeniedPreTransport
		for _, v := range execRes.Violations {
			if v.Code == "approval_transport_missing_response" {
				state = StateTransportError
				break
			}
		}
		res.State = state
		res.PassFail = "FAIL"
		return res
	}

	res.Executed = true
	res.State = StateExecuted
	phr := EvaluatePassiveHeaders(HeaderCheckInput{ExecutionResult: execRes, RequestedMethod: MethodHEAD})
	res.PassiveHeaderCheck = &phr
	res.CandidateFindings = append([]CandidateFinding{}, phr.CandidateFindings...)
	res.Limitations = append([]string{}, phr.Limitations...)

	if len(phr.CandidateFindings) > 0 {
		res.State = StateCompletedWithCandidates
	} else {
		res.State = StateCompletedNoCandidates
	}
	res.PassFail = "PASS"
	return res
}
