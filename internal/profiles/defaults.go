package profiles

func DefaultProfiles() []Profile {
	return []Profile{
		{ID: "default", Name: "Default", Description: "General-purpose balanced"},
		{ID: "minimal", Name: "Minimal", Description: "Minimal setup required components"},
		{ID: "development", Name: "Development", Description: "SDD/coding oriented"},
		{ID: "research", Name: "Research", Description: "SDR/research notes oriented"},
		{ID: "blog", Name: "Blog", Description: "Article drafting/validation/publishing prep"},
		{ID: "oscp", Name: "OSCP", Description: "OSCP notes/labs/methodology learning"},
		{ID: "bug-bounty", Name: "Bug Bounty", Description: "Authorized target research, recon notes, hypothesis tracking, report prep"},
		{ID: "redteam", Name: "Red Team", Description: "Authorized red team research, operation notes, tradecraft documentation, detection-aware thinking"},
	}
}
