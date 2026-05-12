package components

type Status string

const (
	StatusUnknown        Status = "unknown"
	StatusAvailable      Status = "available"
	StatusDetected       Status = "detected"
	StatusInstalled      Status = "installed"
	StatusConfigured     Status = "configured"
	StatusMissing        Status = "missing"
	StatusModeledOnly    Status = "modeled_only"
	StatusManualStep     Status = "manual_step_needed"
	StatusError          Status = "error"
	StatusNotImplemented Status = "not_implemented"
)

type MCPReadiness struct {
	Status          Status
	DetectedEvidence []string
	Blockers        []string
	ManualHints     []string
	SafeUsable      bool
}

func BuildMCPReadiness(status Status, evidence []string, blockers []string, hints []string, safeUsable bool) MCPReadiness {
	return MCPReadiness{
		Status:           status,
		DetectedEvidence: append([]string{}, evidence...),
		Blockers:         append([]string{}, blockers...),
		ManualHints:      append([]string{}, hints...),
		SafeUsable:       safeUsable,
	}
}

type Component struct {
	ID                 string
	Name               string
	Description        string
	Category           string
	Status             Status
	Required           bool
	Optional           bool
	InstallSupported   bool
	ConfigureSupported bool
	DocsURL            string
	Notes              string
}

func (c Component) StatusLabel() string {
	return string(c.Status)
}
