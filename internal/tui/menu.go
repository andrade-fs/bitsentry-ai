package tui

type menuItem struct {
	Title  string
	Screen screen
}

func defaultMenu() []menuItem {
	return []menuItem{
		{Title: "Install / Setup", Screen: screenInstall},
		{Title: "System check", Screen: screenSystem},
		{Title: "Detect AI agents", Screen: screenAgents},
		{Title: "Components", Screen: screenComponents},
		{Title: "Capabilities", Screen: screenCapabilities},
		{Title: "Profiles", Screen: screenProfiles},
		{Title: "Workflows", Screen: screenWorkflows},
		{Title: "Settings", Screen: screenSettings},
		{Title: "Exit", Screen: screenExit},
	}
}
