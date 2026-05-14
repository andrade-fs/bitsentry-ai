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
	MaxPreviewSizeBytes   int64
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
	RequestID             string
	EvidenceID            string
	ApprovalID            string
	ApprovedBy            string
	ApprovalExpiresAt     time.Time
	Method                RequestMethod
	URL                   string
	StatusCode            int
	FinalURL              string
	RedirectObserved      bool
	RedirectLocation      string
	HeadersRedacted       map[string]string
	BodyPreviewRedacted   string
	BodyTruncated         bool
	ResponseSize          int64
	MaxPreviewSize        int64
	PolicyDecision        string
	Violations            []PolicyViolation
	LinkedFindingIDs      []string
	FollowRedirects       bool
	SensitiveDataRedacted bool
	SafetyNotes           []string
	RedactionsApplied     []string
}



type ControlledHTTPExecutor interface {
	ExecuteApproved(ctx AssessmentSessionContext, req PlannedRequest, approval *ExecutionApproval) ExecutionResult
}

type HTTPTransport interface {
	Execute(req PlannedRequest) (FakeTransportResponse, error)
}

func NewOfflineAdapter(evaluator PolicyEvaluator, planner DryRunPlanner, recorder EvidenceRecorder) *OfflineAdapter {
	return &OfflineAdapter{evaluator: evaluator, planner: planner, recorder: recorder}
}

type ObservationStatus string

const (
	ObservationStatusPresent       ObservationStatus = "present"
	ObservationStatusMissing       ObservationStatus = "missing"
	ObservationStatusWeak          ObservationStatus = "weak"
	ObservationStatusNotApplicable ObservationStatus = "not_applicable"
	ObservationStatusNeedsContext  ObservationStatus = "needs_context"
)

type PassiveCheckID string

const (
	PassiveCheckIDHeadersMVP     PassiveCheckID = "passive_headers_mvp"
	PassiveCheckIDRobotsMVP      PassiveCheckID = "passive_robots_mvp"
	PassiveCheckIDSitemapMVP     PassiveCheckID = "passive_sitemap_mvp"
	PassiveCheckIDSecurityTxtMVP PassiveCheckID = "passive_securitytxt_mvp"
)

type FindingCategory string

const (
	FindingCategoryConfiguration FindingCategory = "Configuration"
	FindingCategoryExposure      FindingCategory = "Exposure"
	FindingCategoryInformational FindingCategory = "Informational"
)

type SeverityHint string

const (
	SeverityCritical      SeverityHint = "Critical"
	SeverityHigh          SeverityHint = "High"
	SeverityMedium        SeverityHint = "Medium"
	SeverityLow           SeverityHint = "Low"
	SeverityInformational SeverityHint = "Informational"
)

type ConfidenceHint string

const (
	ConfidenceHigh   ConfidenceHint = "High"
	ConfidenceMedium ConfidenceHint = "Medium"
	ConfidenceLow    ConfidenceHint = "Low"
)

type PassiveObservation struct {
	ObservationID     string
	Title             string
	Status            ObservationStatus
	EvidenceID        string
	AffectedURL       string
	AffectedComponent string
	ObservedValue     string
	SeverityHint      SeverityHint
	ConfidenceHint    ConfidenceHint
	Notes             string
	SourceCheckID     PassiveCheckID
}

type CandidateFinding struct {
	CandidateID           string
	Title                 string
	Category              FindingCategory
	SeverityHint          SeverityHint
	ConfidenceHint        ConfidenceHint
	EvidenceID            string
	RelatedObservationIDs []string
	AffectedURL           string
	AffectedComponent     string
	Impact                string
	Likelihood            string
	Remediation           string
	Verification          string
	Limitations           []string
	SourceCheckID         PassiveCheckID
}

type PassiveCheckResult struct {
	CheckID           PassiveCheckID
	EvidenceID        string
	Observations      []PassiveObservation
	CandidateFindings []CandidateFinding
	Limitations       []string
}

type HeaderCheckResult = PassiveCheckResult

type SurfaceScopeStatus string

const (
	SurfaceScopeInScope      SurfaceScopeStatus = "in_scope"
	SurfaceScopeOutOfScope   SurfaceScopeStatus = "out_of_scope"
	SurfaceScopeUnknownScope SurfaceScopeStatus = "unknown_scope"
)

type SensitivityHint string

const (
	SensitivityNone    SensitivityHint = "none"
	SensitivityAdmin   SensitivityHint = "admin"
	SensitivityBackup  SensitivityHint = "backup"
	SensitivityPrivate SensitivityHint = "private"
	SensitivityStaging SensitivityHint = "staging"
	SensitivityAuth    SensitivityHint = "auth"
	SensitivityUnknown SensitivityHint = "unknown"
)

type SurfaceSource string

const (
	SurfaceSourceSitemap         SurfaceSource = "sitemap"
	SurfaceSourceRobots          SurfaceSource = "robots"
	SurfaceSourceSecurityTxt     SurfaceSource = "securitytxt"
	SurfaceSourceExecutionResult SurfaceSource = "execution_result"
	SurfaceSourceHeaders         SurfaceSource = "headers"
)

type SurfaceHost struct {
	Host        string
	ScopeStatus SurfaceScopeStatus
	EvidenceIDs []string
	Sources     []SurfaceSource
}

type SurfaceURL struct {
	URL         string
	Host        string
	Path        string
	ScopeStatus SurfaceScopeStatus
	Source      SurfaceSource
	EvidenceIDs []string
}

type SurfacePath struct {
	Path            string
	Source          SurfaceSource
	SensitivityHint SensitivityHint
	EvidenceIDs     []string
}

type SurfaceSignal struct {
	SignalID               string
	Title                  string
	Category               FindingCategory
	EvidenceID             string
	SourceCheckID          PassiveCheckID
	RelatedObservationIDs  []string
	Notes                  string
}

type SurfaceCandidateArea struct {
	AreaID              string
	Category            FindingCategory
	Reason              string
	EvidenceIDs         []string
	SuggestedNextCheck  string
	ConfidenceHint      ConfidenceHint
}

type SurfaceMap struct {
	MapID           string
	ScopeHosts      []string
	Hosts           []SurfaceHost
	URLs            []SurfaceURL
	Paths           []SurfacePath
	Signals         []SurfaceSignal
	CandidateAreas  []SurfaceCandidateArea
	EvidenceIDs     []string
	Limitations     []string
}
