package agents

type OpenCodeExportPlan struct {
	TargetAgent      string              `json:"target_agent"`
	TargetConfigPath string              `json:"target_config_path"`
	MCPs             []OpenCodeMCPPlan   `json:"mcps"`
	Skills           []OpenCodeSkillPlan `json:"skills"`
	Memory           OpenCodeMemoryPlan  `json:"memory"`
	Warnings         []string            `json:"warnings"`
	Actions          []string            `json:"actions"`
}

type OpenCodeMCPPlan struct {
	ID         string `json:"id"`
	Selected   bool   `json:"selected"`
	Configured bool   `json:"configured"`
	Status     string `json:"status"`
}

type OpenCodeSkillPlan struct {
	ID         string `json:"id"`
	Selected   bool   `json:"selected"`
	Configured bool   `json:"configured"`
	Status     string `json:"status"`
}

type OpenCodeMemoryPlan struct {
	EngramEnabled    bool   `json:"engram_enabled"`
	EngramConfigured bool   `json:"engram_configured"`
	EngramBinaryPath string `json:"engram_binary_path"`
	EngramProject    string `json:"engram_project"`
	Context7Enabled  bool   `json:"context7_enabled"`
	Context7Command  string `json:"context7_command"`
	Context7Package  string `json:"context7_package"`
}
