package components

type Status string

const (
	StatusUnknown        Status = "unknown"
	StatusAvailable      Status = "available"
	StatusDetected       Status = "detected"
	StatusInstalled      Status = "installed"
	StatusConfigured     Status = "configured"
	StatusMissing        Status = "missing"
	StatusError          Status = "error"
	StatusNotImplemented Status = "not_implemented"
)

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
