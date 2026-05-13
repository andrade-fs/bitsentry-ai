package securityweb

import "errors"

var (
	ErrExecutionModeDenied     = errors.New("execution mode denied")
	ErrMissingExplicitApproval = errors.New("explicit approval required")
	ErrScopeViolation          = errors.New("target out of scope")
	ErrSchemeDenied            = errors.New("scheme must be http/https")
	ErrMethodDenied            = errors.New("request method denied")
	ErrMissingRateLimit        = errors.New("rate limit required")
	ErrMissingRequestBudget    = errors.New("request budget required")
	ErrMissingTimeout          = errors.New("timeout required")
	ErrMissingMaxResponseSize  = errors.New("max response size required")
	ErrMissingStopConditions   = errors.New("stop conditions required")
	ErrMissingEvidencePlan     = errors.New("evidence plan required")
	ErrToolClassNotAllowed     = errors.New("tool class not allowed")
	ErrOutOfScopeRedirect      = errors.New("out-of-scope redirects denied")
	ErrApprovalRequired        = errors.New("valid execution approval required")
	ErrApprovalRequestMismatch = errors.New("approved request id mismatch")
	ErrApprovalMethodMismatch  = errors.New("approved method mismatch")
	ErrApprovalURLMismatch     = errors.New("approved url mismatch")
	ErrApprovalMissingExpires  = errors.New("approval expires_at required")
	ErrApprovalExpired         = errors.New("approval expired")
	ErrLimiterConcurrent       = errors.New("only one request at a time is allowed")
)
