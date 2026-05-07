package tui

import (
	"context"
	"fmt"
	"strings"

	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/workflows"
)

func renderScreen(m model) string {
	switch m.screen {
	case screenInstall:
		return renderInstall(m)
	case screenSystem:
		return renderSystem(m)
	case screenAgents:
		return renderAgents(m)
	case screenComponents:
		return renderComponents(m)
	case screenCapabilities:
		return renderCapabilities(m)
	case screenProfiles:
		return renderProfiles(m)
	case screenWorkflows:
		return renderWorkflows(m)
	case screenSettings:
		return renderSettings(m)
	default:
		return renderHome(m)
	}
}

func frame(m model, title string, body ...string) string {
	parts := []string{m.styles.title.Render("bitsentry-ai"), m.styles.muted.Render("Active profile: " + m.active), "", m.styles.section.Render(title)}
	parts = append(parts, body...)
	if m.errMsg != "" {
		parts = append(parts, "", m.styles.error.Render("Error: "+m.errMsg))
	}
	if m.message != "" {
		parts = append(parts, "", m.styles.hint.Render(m.message))
	}
	parts = append(parts, "", m.styles.muted.Render("Esc/backspace: home • q/Ctrl+C: quit"))
	return strings.Join(parts, "\n")
}

func renderHome(m model) string {
	lines := make([]string, 0, len(m.menu)+2)
	for i, item := range m.menu {
		cursor := "  "
		line := item.Title
		if i == m.selected {
			cursor = "➜ "
			line = m.styles.selected.Render(line)
		}
		lines = append(lines, cursor+line)
	}
	lines = append(lines, "", m.styles.muted.Render("Navigate: up/down or j/k • Enter: select • q/Ctrl+C: quit"))
	return frame(m, "Main Menu", lines...)
}

func renderInstall(m model) string {
	cfgDir, cfgPath := configPaths()
	openCodeStatus := "not detected"
	for _, a := range m.agents {
		if a.ID == "opencode" {
			if a.Found {
				openCodeStatus = "detected"
				if a.Path != "" {
					openCodeStatus = fmt.Sprintf("detected (%s)", a.Path)
				}
			} else if a.Hint != "" {
				openCodeStatus = "not detected"
			}
		}
	}
	return frame(m, "Install / Setup",
		fmt.Sprintf("- OS: %s", m.system.OS),
		fmt.Sprintf("- Arch: %s", m.system.Arch),
		fmt.Sprintf("- Shell: %s", m.shell),
		fmt.Sprintf("- Package manager: %s", packageManager(m.deps)),
		fmt.Sprintf("- Config path: %s", cfgPath),
		fmt.Sprintf("- Config dir: %s", cfgDir),
		fmt.Sprintf("- Active profile: %s", m.active),
		fmt.Sprintf("- Dependency summary: %s", dependencySummary(m.deps)),
		fmt.Sprintf("- OpenCode: %s", openCodeStatus),
		"",
		"Component installation is not implemented yet. This screen is prepared for Phase 5+.",
	)
}

func renderSystem(m model) string {
	lines := []string{
		fmt.Sprintf("- OS: %s", m.system.OS),
		fmt.Sprintf("- Arch: %s", m.system.Arch),
		fmt.Sprintf("- Shell: %s", m.shell),
		fmt.Sprintf("- Package manager: %s", packageManager(m.deps)),
		"- Dependencies:",
	}
	for _, d := range m.deps {
		status := "not found"
		if d.Found {
			status = "found"
		}
		required := "optional"
		if d.Mandatory {
			required = "required"
		}
		entry := fmt.Sprintf("  - %s: %s (%s)", d.Name, status, required)
		if d.Path != "" {
			entry += fmt.Sprintf(" [%s]", d.Path)
		}
		lines = append(lines, entry)
	}
	return frame(m, "System check", lines...)
}

func renderAgents(m model) string {
	lines := make([]string, 0, len(m.agents)+2)
	if len(m.agents) == 0 {
		lines = append(lines, "No agent detectors registered.")
	}
	for _, a := range m.agents {
		status := "not detected"
		if a.Found {
			status = "detected"
		}
		lines = append(lines, fmt.Sprintf("- %s (%s): %s", a.Name, a.ID, status))
		if a.Path != "" {
			lines = append(lines, fmt.Sprintf("  path: %s", a.Path))
		}
		if a.Version != "" {
			lines = append(lines, fmt.Sprintf("  version: %s", a.Version))
		}
	}
	return frame(m, "Detect AI agents", lines...)
}

func renderComponents(m model) string {
	entries := components.Registry()
	engramDetails := components.EngramRuntimeDetails{Status: components.StatusError, Notes: []string{"Unable to run Engram runtime detection."}}
	context7Details := components.Context7RuntimeDetails{Status: components.StatusError, Notes: []string{"Unable to run Context7 runtime detection."}}
	mcpSummary := components.MCPRegistrySummary{}
	skillsSummary := components.SkillsRegistrySummary{}
	if cfg, err := m.app.ConfigManager.Load(); err == nil {
		engramDetails = components.DetectEngramRuntime(context.Background(), cfg)
		context7Details = components.DetectContext7Runtime(context.Background(), cfg)
		_, mcpSummary = components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
		_, skillsSummary = components.SkillsRegistry(cfg)
	} else {
		engramDetails.Notes = append(engramDetails.Notes, fmt.Sprintf("config error: %v", err))
		context7Details.Notes = append(context7Details.Notes, fmt.Sprintf("config error: %v", err))
	}
	for i := range entries {
		if entries[i].ID == "engram" {
			entries[i].Status = engramDetails.Status
			entries[i].Notes = strings.Join(engramDetails.Notes, " ")
		}
		if entries[i].ID == "context7" {
			entries[i].Status = context7Details.Status
			entries[i].Notes = strings.Join(context7Details.Notes, " ")
		}
		if entries[i].ID == "mcps" {
			if mcpSummary.Configured {
				entries[i].Status = components.StatusConfigured
			} else if mcpSummary.Enabled {
				entries[i].Status = components.StatusDetected
			}
			entries[i].Notes = fmt.Sprintf("Configured=%s selected=%s", yesNoLabel(mcpSummary.Configured), valueOrDefault(strings.Join(mcpSummary.Selected, ", "), "none"))
		}
		if entries[i].ID == "skills" {
			if skillsSummary.Configured {
				entries[i].Status = components.StatusConfigured
			} else if skillsSummary.Enabled {
				entries[i].Status = components.StatusDetected
			}
			entries[i].Notes = fmt.Sprintf("Configured=%s selected=%s", yesNoLabel(skillsSummary.Configured), valueOrDefault(strings.Join(skillsSummary.Selected, ", "), "none"))
		}
	}

	lines := []string{"Components status (Engram + Context7 runtime-detected, others modeled):"}
	for _, c := range entries {
		req := "optional"
		if c.Required {
			req = "required"
		}
		lines = append(lines, fmt.Sprintf("- %s (%s)", c.Name, c.ID))
		lines = append(lines, fmt.Sprintf("  status: %s • category: %s • %s", c.StatusLabel(), c.Category, req))
		lines = append(lines, fmt.Sprintf("  desc: %s", c.Description))
		if c.Notes != "" {
			lines = append(lines, fmt.Sprintf("  note: %s", c.Notes))
		}
		if c.DocsURL != "" {
			lines = append(lines, fmt.Sprintf("  docs: %s", c.DocsURL))
		}
		if c.ID == "engram" {
			lines = append(lines, fmt.Sprintf("  runtime: binary=%s path=%s", yesNoLabel(engramDetails.BinaryFound), valueOrNA(engramDetails.BinaryPath)))
			lines = append(lines, fmt.Sprintf("  runtime: version=%s", valueOrDefault(engramDetails.Version, "unavailable")))
			lines = append(lines, fmt.Sprintf("  runtime: data_dir=%s path=%s", yesNoLabel(engramDetails.DataDirFound), valueOrNA(engramDetails.DataDirPath)))
			lines = append(lines, fmt.Sprintf("  runtime: config enabled=%s configured=%s", yesNoLabel(engramDetails.ConfigEnabled), yesNoLabel(engramDetails.ConfigConfigured)))
		}
		if c.ID == "context7" {
			lines = append(lines, fmt.Sprintf("  runtime: command detected=%s command=%s", yesNoLabel(context7Details.Detected), valueOrDefault(context7Details.DetectedCommand, "not found")))
			lines = append(lines, fmt.Sprintf("  runtime: path=%s", valueOrDefault(context7Details.DetectedPath, "not found")))
			lines = append(lines, fmt.Sprintf("  runtime: config enabled=%s configured=%s", yesNoLabel(context7Details.ConfigEnabled), yesNoLabel(context7Details.ConfigConfigured)))
			lines = append(lines, fmt.Sprintf("  runtime: config command=%s package=%s", valueOrDefault(context7Details.ConfigCommand, "n/a"), valueOrDefault(context7Details.ConfigPackage, "n/a")))
		}
		if c.ID == "mcps" {
			lines = append(lines, fmt.Sprintf("  metadata: configured=%s enabled=%s", yesNoLabel(mcpSummary.Configured), yesNoLabel(mcpSummary.Enabled)))
			lines = append(lines, fmt.Sprintf("  metadata: selected=%s", valueOrDefault(strings.Join(mcpSummary.Selected, ", "), "none")))
			lines = append(lines, "  note: no agent config has been written yet")
		}
		if c.ID == "skills" {
			lines = append(lines, fmt.Sprintf("  metadata: configured=%s enabled=%s", yesNoLabel(skillsSummary.Configured), yesNoLabel(skillsSummary.Enabled)))
			lines = append(lines, fmt.Sprintf("  metadata: selected=%s", valueOrDefault(strings.Join(skillsSummary.Selected, ", "), "none")))
			lines = append(lines, "  note: no external agent config has been written yet")
		}
	}
	lines = append(lines, "", "Install/configure actions are intentionally disabled in this batch.")
	lines = append(lines, "Runtime detection/config validation for MCPs/Skills and workflow runners is planned for a future batch.")
	return frame(m, "Components", lines...)
}

func yesNoLabel(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func valueOrNA(v string) string {
	return valueOrDefault(v, "n/a")
}

func valueOrDefault(v string, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func renderProfiles(m model) string {
	lines := []string{"Select a profile and press Enter to activate:"}
	for i, p := range m.profiles {
		marker := " "
		if p.IsActive {
			marker = "*"
		}
		prefix := "  "
		label := fmt.Sprintf("%s %s (%s) - %s", marker, p.Name, p.ID, p.Description)
		if i == m.selected {
			prefix = "➜ "
			label = m.styles.selected.Render(label)
		}
		lines = append(lines, prefix+label)
	}
	lines = append(lines, "", m.styles.muted.Render("* active profile • up/down or j/k to move • Enter to activate"))
	return frame(m, "Profiles", lines...)
}

func renderCapabilities(m model) string {
	preset := valueOrDefault(m.capPreset, "custom")
	target := valueOrDefault(m.capTarget, "opencode")
	mcps := valueOrDefault(strings.Join(sortedSelected(m.capMCPs), ", "), "(none)")
	skills := valueOrDefault(strings.Join(sortedSelected(m.capSkills), ", "), "(none)")
	flows := valueOrDefault(strings.Join(sortedSelected(m.capFlows), ", "), "(none)")

	plan := valueOrDefault(m.capLastPlan, "(none)")
	validation := valueOrDefault(m.capLastValidate, "(none)")
	draftState := "saved"
	if m.capDirty {
		draftState = "unsaved changes"
	}

	return frame(m, "Capabilities",
		"Capability selector MVP (OpenCode is the only real target today)",
		"",
		fmt.Sprintf("- Draft preset: %s  ([ / ] to cycle, s to save)", preset),
		fmt.Sprintf("- Draft target: %s (fixed to opencode in this phase)", target),
		fmt.Sprintf("- Draft state: %s", draftState),
		fmt.Sprintf("- Draft MCPs: %s  (m=engram, c=context7, p=postgres[modeled/skip])", mcps),
		fmt.Sprintf("- Draft skills: %s  (z=bitsentry-sdd, x=research-init, n=bugbounty-notes)", skills),
		fmt.Sprintf("- Draft flows: %s  (1=sdd, 2=sdr, 3=notes, 4=redteam/security-notes)", flows),
		"",
		fmt.Sprintf("- Validate result: %s (v)", validation),
		fmt.Sprintf("- Plan preview: %s (l)", plan),
		"",
		"- Dry-run apply: d (requires saved draft)",
		"- Real apply: a, then confirm with y",
		"- Skills/flows are declarative in this phase",
		"- Postgres MCP is modeled but skipped in real apply",
	)
}

func renderWorkflows(m model) string {
	lines := []string{}
	for _, w := range workflows.Registry() {
		lines = append(lines, fmt.Sprintf("- %s (%s): %s", w.Name, w.ID, w.Status))
	}
	lines = append(lines, "", "Workflow execution is not implemented yet.")
	return frame(m, "Workflows", lines...)
}

func renderSettings(m model) string {
	cfgDir, cfgPath := configPaths()
	dryRun := "off"
	if m.dryRun {
		dryRun = "on"
	}
	return frame(m, "Settings",
		fmt.Sprintf("- Config directory: %s", cfgDir),
		fmt.Sprintf("- Config file: %s", cfgPath),
		fmt.Sprintf("- Active profile: %s", m.active),
		fmt.Sprintf("- Dry-run: %s", dryRun),
	)
}
