package tui

import (
	"context"
	"fmt"
	"strings"

	"bitsentry-ai/internal/capabilities"
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
	stepTitle := []string{
		"Step 1 of 6 — Confirm OpenCode target",
		"Step 2 of 6 — Choose install mode",
		"Step 3 of 6 — Install intent summary",
		"Step 4 of 6 — Review install plan",
		"Step 5 of 6 — Install / Update",
		"Step 6 of 6 — Done / Control panel summary",
	}
	preset, _ := capabilities.PresetByID(presetForMode(m.install.InstallMode), capabilities.DefaultPresets())
	lines := []string{stepTitle[m.install.CurrentStep], ""}

	switch m.install.CurrentStep {
	case installStepTarget:
		targetLabel := "[ ] OpenCode"
		if m.install.TargetSelected {
			targetLabel = "[x] OpenCode"
		}
		cursor := "  "
		if m.install.Cursor == 0 {
			cursor = "➜ "
		}
		lines = append(lines,
			"Detected target:",
			cursor+targetLabel,
			fmt.Sprintf("     binary: %s", valueOrDefault(m.install.OpenCodeBinary, "not found")),
			fmt.Sprintf("     config root: %s", valueOrDefault(m.install.OpenCodeConfig, "not resolved")),
			fmt.Sprintf("     pack root: %s", valueOrDefault(m.install.BitsentryPackRoot, "not resolved")),
			fmt.Sprintf("     pack installed: %s", openCodeStatusLine(m.install.PackInstalled)),
			"",
			"This installer is OpenCode-only in this phase.",
			"It will not execute flows/skills and will not mutate credentials.",
			"",
			"Enter: continue • Space: toggle • r: refresh • Esc: back",
		)
	case installStepMode:
		everything := "  [ ] Install Everything"
		packOnly := "  [ ] Install Bitsentry Pack"
		update := "  [ ] Update/Reinstall Bitsentry Pack"
		if m.install.Cursor == installModeEverything {
			everything = "➜ [x] Install Everything"
		}
		if m.install.Cursor == installModePackOnly {
			packOnly = "➜ [x] Install Bitsentry Pack"
		}
		if m.install.Cursor == installModeUpdateReinstall {
			update = "➜ [x] Update/Reinstall Bitsentry Pack"
		}
		lines = append(lines,
			everything,
			"     Includes pack export + native OpenCode integration + /bit-* commands.",
			packOnly,
			"     Exports pack files only. No agent/command/native skill registration.",
			update,
			"     Re-runs install using existing OpenCode root with backup safeguards.",
			"",
			"Advanced granularity (manual preset/skill/flow tuning) is moved out of this main path.",
			"Use Capabilities screen only when you need debug/plumbing parity.",
			"",
			"Enter: continue • up/down: move • Space: select • Esc: back",
		)
	case installStepReview:
		selectedMode := installModeLabel(m.install.InstallMode)
		registerAgent, installCommands, installNativeSkills := nativeOptionsForMode(m.install.InstallMode)
		lines = append(lines,
			"Detected state:",
			fmt.Sprintf("- OpenCode detected: %s", boolLabel(m.install.OpenCodeDetected, "yes", "no")),
			fmt.Sprintf("- config root: %s", valueOrDefault(m.install.OpenCodeConfig, "not resolved")),
			fmt.Sprintf("- managed pack path: %s", valueOrDefault(m.install.BitsentryPackRoot, "not resolved")),
			fmt.Sprintf("- current pack status: %s", boolLabel(m.install.PackInstalled, "installed", "not installed")),
			"",
			"Selected path:",
			fmt.Sprintf("- mode: %s", selectedMode),
			fmt.Sprintf("- preset used by installer: %s", preset.ID),
			fmt.Sprintf("- flows: %s", valueOrDefault(strings.Join(preset.Flows, ", "), "none")),
			fmt.Sprintf("- skills: %s", valueOrDefault(strings.Join(preset.Skills, ", "), "none")),
			"",
			"Detected MCP readiness:",
			fmt.Sprintf("- Engram: %s", m.install.MCPReadiness["engram"].Status),
			fmt.Sprintf("- Context7: %s", m.install.MCPReadiness["context7"].Status),
			"",
			"Will do:",
			"- export Bitsentry capability pack",
			"- verify OPENCODE_USAGE.md",
			"- verify skill-registry.md",
			fmt.Sprintf("- register bitsentry agent in opencode.json: %s", boolLabel(registerAgent, "yes", "no")),
			fmt.Sprintf("- install /bit-* commands: %s", boolLabel(installCommands, "yes", "no")),
			fmt.Sprintf("- install native OpenCode skills: %s", boolLabel(installNativeSkills, "yes", "no")),
			"- create backups before overwrite",
			"",
			"Will NOT do:",
			"- delete user config",
			"- mutate MCP credentials",
			"- activate autonomous runtime",
			"- execute flows",
			"- execute skills",
			"- change agent.bitsentry.permission.edit=deny contract",
		)
		lines = append(lines,
			"",
			"MCP config preview (PREVIEW ONLY):",
			fmt.Sprintf("- current_config_state: %s", valueOrDefault(m.install.MCPConfigPreview.CurrentConfigState, "unknown")),
			fmt.Sprintf("- exists: %s", yesNoLabel(m.install.MCPConfigPreview.Exists)),
			fmt.Sprintf("- readable: %s", yesNoLabel(m.install.MCPConfigPreview.Readable)),
			fmt.Sprintf("- current_mcp_config_detected: %s", yesNoLabel(m.install.MCPConfigPreview.CurrentMCPConfigDetected)),
			fmt.Sprintf("- mcp_readiness_state: %s", valueOrDefault(m.install.MCPConfigPreview.MCPReadinessState, "unknown")),
			fmt.Sprintf("- would_write: %s", yesNoLabel(m.install.MCPConfigPreview.WouldWrite)),
			fmt.Sprintf("- requires_confirmation: %s", yesNoLabel(m.install.MCPConfigPreview.RequiresConfirmation)),
			fmt.Sprintf("- backup_required: %s", yesNoLabel(m.install.MCPConfigPreview.BackupRequired)),
		)
		lines = append(lines, readinessSummaryLines("Engram", m.install.MCPReadiness["engram"])...)
		lines = append(lines, readinessSummaryLines("Context7", m.install.MCPReadiness["context7"])...)
		if len(m.install.MCPConfigPreview.ProposedSafeChanges) > 0 {
			lines = append(lines, "", "Preview proposed_safe_changes:")
			for _, c := range m.install.MCPConfigPreview.ProposedSafeChanges {
				lines = append(lines, fmt.Sprintf("- %s", c))
			}
		}
		if len(m.install.MCPConfigPreview.PreservedKeys) > 0 {
			lines = append(lines, fmt.Sprintf("- preserved_keys: %s", strings.Join(m.install.MCPConfigPreview.PreservedKeys, ", ")))
		}
		if len(m.install.MCPConfigPreview.PreservedMCPEntries) > 0 {
			lines = append(lines, fmt.Sprintf("- preserved_mcp_entries: %s", strings.Join(m.install.MCPConfigPreview.PreservedMCPEntries, ", ")))
		}
		if len(m.install.MCPConfigPreview.Warnings) > 0 {
			lines = append(lines, "Preview warnings:")
			for _, w := range m.install.MCPConfigPreview.Warnings {
				lines = append(lines, fmt.Sprintf("- %s", w))
			}
		}
		if len(m.install.MCPConfigPreview.ManualSteps) > 0 {
			lines = append(lines, "Preview manual_steps:")
			for _, s := range m.install.MCPConfigPreview.ManualSteps {
				lines = append(lines, fmt.Sprintf("- %s", s))
			}
		}
		if !m.install.TargetSelected {
			lines = append(lines,
				"",
				"Cannot install yet:",
				"- Select at least one target agent.",
			)
		}
		lines = append(lines, "", "Enter: continue • Esc: back")
	case installStepInstall:
		action := installModeLabel(m.install.InstallMode)
		lines = append(lines,
			action,
			"",
			"Press Enter to run.",
			"Esc: back",
		)
	case installStepDone:
		prompt := valueOrDefault(m.install.NextPrompt, "Install did not complete.")
		lines = append(lines,
			fmt.Sprintf("Result: %s", valueOrDefault(m.install.ResultStatus, "FAIL")),
			"",
			"Install summary:",
			fmt.Sprintf("- OpenCode detected: %s", boolLabel(m.install.OpenCodeDetected, "yes", "no")),
			fmt.Sprintf("- OpenCode config root: %s", valueOrDefault(m.install.OpenCodeConfig, "not resolved")),
			fmt.Sprintf("- Bitsentry pack root: %s", valueOrDefault(m.install.BitsentryPackRoot, "not resolved")),
			fmt.Sprintf("- Pack status: %s", boolLabel(m.install.PackInstalled, "installed", "not installed")),
			fmt.Sprintf("- Native integration status: %s", boolLabel(m.install.NativeIntegrationOK, "installed", "not installed")),
			fmt.Sprintf("- Backup path (config): %s", valueOrDefault(m.install.ConfigBackupPath, "none")),
			fmt.Sprintf("- Backup path (native skills): %s", valueOrDefault(m.install.NativeBackupPath, "none")),
			"",
			"MCP config preview contract (PREVIEW ONLY):",
			fmt.Sprintf("- state: %s", valueOrDefault(m.install.MCPConfigPreview.CurrentConfigState, "unknown")),
			fmt.Sprintf("- would_write: %s", yesNoLabel(m.install.MCPConfigPreview.WouldWrite)),
			fmt.Sprintf("- requires_confirmation: %s", yesNoLabel(m.install.MCPConfigPreview.RequiresConfirmation)),
			fmt.Sprintf("- backup_required: %s", yesNoLabel(m.install.MCPConfigPreview.BackupRequired)),
		)
		if len(m.install.MCPConfigPreview.ManualSteps) > 0 {
			lines = append(lines, "- manual steps:")
			for _, step := range m.install.MCPConfigPreview.ManualSteps {
				lines = append(lines, fmt.Sprintf("  - %s", step))
			}
		}
		if len(m.install.ResultNotes) > 0 {
			lines = append(lines, "", "Manual steps / notes:")
			for _, n := range m.install.ResultNotes {
				lines = append(lines, fmt.Sprintf("- %s", n))
			}
		}
		lines = append(lines, "", "MCP readiness summary:")
		lines = append(lines, readinessSummaryLines("Engram", m.install.MCPReadiness["engram"])...)
		lines = append(lines, readinessSummaryLines("Context7", m.install.MCPReadiness["context7"])...)
		lines = append(lines,
			"",
			"OpenCode prompt:",
		)
		lines = append(lines, strings.Split(prompt, "\n")...)
		lines = append(lines,
			"",
			"Esc: back",
		)
	}

	return frame(m, "Install / Setup", lines...)
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

func readinessSummaryLines(name string, readiness components.MCPReadiness) []string {
	lines := []string{fmt.Sprintf("- %s status: %s (safe_usable=%s)", name, readiness.Status, yesNoLabel(readiness.SafeUsable))}
	if len(readiness.DetectedEvidence) > 0 {
		lines = append(lines, fmt.Sprintf("  evidence: %s", strings.Join(readiness.DetectedEvidence, "; ")))
	}
	if len(readiness.Blockers) > 0 {
		lines = append(lines, fmt.Sprintf("  blockers: %s", strings.Join(readiness.Blockers, "; ")))
	}
	if len(readiness.ManualHints) > 0 {
		lines = append(lines, fmt.Sprintf("  manual: %s", strings.Join(readiness.ManualHints, "; ")))
	}
	return lines
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
