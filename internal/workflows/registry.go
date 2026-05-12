package workflows

func Registry() []Workflow {
	return []Workflow{
		{ID: "sdd", Name: "SDD", Status: "not yet implemented. Coming in Phase 2+"},
		{ID: "sdr", Name: "SDR", Status: "not yet implemented. Coming in Phase 2+"},
		{ID: "source-security-review", Name: "Source Security Review", Status: "not yet implemented. Declarative/read-only-first in Phase 7.1"},
		{ID: "red-team", Name: "Red Team", Status: "not yet implemented. Coming in Phase 2+"},
	}
}
