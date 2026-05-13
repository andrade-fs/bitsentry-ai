package securityweb

import "time"

type ExecutionMode string

const (
	ExecutionModePlanningOnly   ExecutionMode = "planning_only"
	ExecutionModeDryRun         ExecutionMode = "dry_run"
	ExecutionModeExecuteApproved ExecutionMode = "execute_approved"
	ExecutionModeRetest         ExecutionMode = "retest"
)

type Intensity string

const (
	IntensityLow    Intensity = "low"
	IntensityMedium Intensity = "medium"
	IntensityHigh   Intensity = "high"
)

type ToolClass string

const (
	ToolClassManual      ToolClass = "manual"
	ToolClassSafeProbe   ToolClass = "safe_probe"
	ToolClassProhibited  ToolClass = "prohibited"
)

type RequestMethod string

const (
	MethodGET     RequestMethod = "GET"
	MethodHEAD    RequestMethod = "HEAD"
	MethodPOST    RequestMethod = "POST"
	MethodPUT     RequestMethod = "PUT"
	MethodPATCH   RequestMethod = "PATCH"
	MethodDELETE  RequestMethod = "DELETE"
	MethodOPTIONS RequestMethod = "OPTIONS"
	MethodTRACE   RequestMethod = "TRACE"
	MethodCONNECT RequestMethod = "CONNECT"
)

type ControlledWebRequestAdapter interface {
	Plan(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation)
	Validate(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation)
	RenderEvidenceTemplate(entry EvidenceEntry) string
	RedactEvidence(entry EvidenceEntry) EvidenceEntry
}

type AssessmentSessionContext struct {
	SessionID             string
	AuthorizationRef      string
	ScopeRef              string
	InScopeTargets        []string
	ExecutionMode         ExecutionMode
	Intensity             Intensity
	ExplicitApproval      bool
	RateLimitPerMinute    int
	RequestBudget         int
	TimeoutSeconds        int
	MaxResponseSizeBytes  int64
	StopConditions        []string
	EvidencePlanRef       string
	AllowedToolClasses    []ToolClass
	ProhibitedToolClasses []ToolClass
	ExistingFindingID     string
	ExistingCheckID       string
}

type PlannedRequest struct {
	RequestRef string
	Target     string
	URL        string
	Method     RequestMethod
	ToolClass  ToolClass
	Headers    map[string]string
}

type DiscoveryPlanItem struct {
	Request          PlannedRequest
	PolicyDecision   PolicyDecision
	PolicyViolations []PolicyViolation
	EvidenceTemplate EvidenceEntry
}

type DiscoveryPlan struct {
	PlanRef       string
	Target        string
	ExecutionMode ExecutionMode
	WouldExecute  bool
	Items         []DiscoveryPlanItem
}

type PolicyDecision struct {
	Allowed    bool
	Reason     string
	Violations []PolicyViolation
}

type PolicyViolation struct {
	Code   string
	Reason string
	Field  string
}

type EvidenceEntry struct {
	EvidenceID         string
	SessionMode        string
	AuthorizationRef   string
	ScopeRef           string
	PlannedRequestRef  string
	PolicyDecision     string
	RedactionApplied   bool
	LinkedFindingIDs   []string
	NotesAssumptions   string
	RequestURL         string
	RequestHeaders     map[string]string
}

type PolicyEvaluator interface {
	Evaluate(ctx AssessmentSessionContext, req PlannedRequest) (PolicyDecision, []PolicyViolation)
}

type DryRunPlanner interface {
	BuildPlan(ctx AssessmentSessionContext, req PlannedRequest) PlannedRequest
}

type Redactor interface {
	RedactHeaders(headers map[string]string) map[string]string
	RedactURL(rawURL string) string
}

type EvidenceRecorder interface {
	Render(entry EvidenceEntry) string
	Redact(entry EvidenceEntry) EvidenceEntry
}

type OfflineAdapter struct {
	evaluator PolicyEvaluator
	planner   DryRunPlanner
	recorder  EvidenceRecorder
}

type ExecutionApproval struct {
	ApprovalID            string
	ApprovedRequestID     string
	ApprovedMethod        RequestMethod
	ApprovedURL           string
	ApprovedScopeRef      string
	ApprovedExecutionMode ExecutionMode
	ApprovedToolClass     ToolClass
	ApprovedIntensity     Intensity
	ApprovedAt            time.Time
	ApprovedBy            string
	ExpiresAt             time.Time
	TTLSeconds            int
	ApprovalTextOrHash    string
	MaxRequests           int
	MaxDurationSeconds    int
	RateLimitPerMinute    int
	StopConditions        []string
}

type ExecutionResult struct {
	RequestID            string
	EvidenceID           string
	Method               RequestMethod
	URL                  string
	StatusCode           int
	FinalURL             string
	RedirectObserved     bool
	RedirectLocation     string
	HeadersRedacted      map[string]string
	BodyPreviewRedacted  string
	BodyTruncated        bool
	ResponseSize         int64
	MaxPreviewSize       int64
	PolicyDecision       string
	Violations           []PolicyViolation
	LinkedFindingIDs     []string
	FollowRedirects      bool
	SensitiveDataRedacted bool
}

func NewOfflineAdapter(evaluator PolicyEvaluator, planner DryRunPlanner, recorder EvidenceRecorder) *OfflineAdapter {
	return &OfflineAdapter{evaluator: evaluator, planner: planner, recorder: recorder}
}
