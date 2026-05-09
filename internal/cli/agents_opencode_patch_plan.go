package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
	"github.com/spf13/cobra"
)

func newAgentsOpenCodePatchPlanCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "patch-plan",
		Short: "Preview schema-aware OpenCode opencode.json patch without writing",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusDetails, err := buildOpenCodeStatus(context.Background(), rt)
			if err != nil {
				return err
			}

			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
			context7Details := components.DetectContext7Runtime(context.Background(), cfg)
			_, mcpSummary := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
			_, skillsSummary := components.SkillsRegistry(cfg)

			targetFile, targetWarning := resolveOpenCodeJSONTarget(statusDetails)
			plan, planWarnings := buildOpenCodeJSONPatchPlan(targetFile, mcpSummary.Selected, skillsSummary.Selected, cfg)

			warnings := append([]string{}, statusDetails.Warnings...)
			if strings.TrimSpace(targetWarning) != "" {
				warnings = append(warnings, targetWarning)
			}
			warnings = append(warnings, planWarnings...)
			warnings = append(warnings,
				"Read-only command: no directory creation and no file writes.",
				"MCP servers are not executed.",
				"Skills are not copied to OpenCode.",
			)

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve user home directory: %w", err)
			}
			futureBackup := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode", time.Now().UTC().Format("20060102T150405Z"))

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "OpenCode opencode.json patch plan")
			fmt.Fprintf(out, "- target config file: %s\n", fallback(plan.TargetFile, "none"))
			fmt.Fprintf(out, "- parse status: %s\n", plan.ParseStatus)

			fmt.Fprintf(out, "- existing top-level keys: %s\n", fallback(strings.Join(plan.TopLevelKeys, ", "), "none"))
			fmt.Fprintf(out, "- existing mcp keys: %s\n", fallback(strings.Join(plan.ExistingMCPKeys, ", "), "none"))

			fmt.Fprintln(out, "- proposed mcp additions/updates:")
			if len(plan.ProposedMCPChanges) == 0 {
				fmt.Fprintln(out, "  - none")
			} else {
				for _, change := range plan.ProposedMCPChanges {
					fmt.Fprintf(out, "  - %s (%s)\n", change.ID, change.Action)
					if strings.TrimSpace(change.Reason) != "" {
						fmt.Fprintf(out, "    reason: %s\n", change.Reason)
					}
					if strings.TrimSpace(change.Note) != "" {
						fmt.Fprintf(out, "    note: %s\n", change.Note)
					}
					if payload, marshalErr := json.Marshal(change.ProposedValue); marshalErr == nil {
						fmt.Fprintf(out, "    proposed_value: %s\n", string(payload))
					}
				}
			}

			fmt.Fprintf(out, "- skill handling decision: %s\n", plan.SkillsDecision)
			if strings.TrimSpace(plan.SkillsDecisionNote) != "" {
				fmt.Fprintf(out, "  note: %s\n", plan.SkillsDecisionNote)
			}

			if len(warnings) > 0 {
				fmt.Fprintln(out, "- warnings:")
				for _, warning := range warnings {
					fmt.Fprintf(out, "  - %s\n", warning)
				}
			}

			fmt.Fprintf(out, "- future backup path: %s\n", futureBackup)
			fmt.Fprintln(out, "Patch plan only. No OpenCode files were modified.")
			return nil
		},
	}
}

type openCodePatchPlan struct {
	TargetFile         string
	ParseStatus        string
	TopLevelKeys       []string
	ExistingMCPKeys    []string
	ProposedMCPChanges []mcpChange
	SkillsDecision     string
	SkillsDecisionNote string
}

type mcpChange struct {
	ID            string
	Action        string
	Reason        string
	Note          string
	ProposedValue map[string]any
}

func resolveOpenCodeJSONTarget(details openCodeStatusDetails) (string, string) {
	if strings.TrimSpace(details.TargetConfigCandidate) == "" {
		return "", "No OpenCode config directory candidate resolved."
	}
	return filepath.Join(details.TargetConfigCandidate, "opencode.json"), ""
}

func buildOpenCodeJSONPatchPlan(targetFile string, selectedMCPs []string, selectedSkills []string, cfg config.Config) (openCodePatchPlan, []string) {
	plan := openCodePatchPlan{
		TargetFile:         targetFile,
		ParseStatus:        "not-parsed",
		TopLevelKeys:       []string{},
		ExistingMCPKeys:    []string{},
		ProposedMCPChanges: []mcpChange{},
		SkillsDecision:     "skills deferred to patch-plan notes",
	}
	warnings := make([]string, 0)

	if strings.TrimSpace(targetFile) == "" {
		plan.ParseStatus = "missing target file"
		warnings = append(warnings, "Cannot parse opencode.json because target path was not resolved.")
		return plan, warnings
	}

	raw, err := os.ReadFile(targetFile)
	if err != nil {
		plan.ParseStatus = fmt.Sprintf("read failed: %v", err)
		warnings = append(warnings, "opencode.json could not be read; patch values are proposed from bitsentry config only.")
	} else {
		var root map[string]any
		if err := json.Unmarshal(raw, &root); err != nil {
			plan.ParseStatus = fmt.Sprintf("json parse failed: %v", err)
			warnings = append(warnings, "opencode.json JSON parse failed; schema-aware merge cannot be validated against existing document.")
		} else {
			plan.ParseStatus = "ok"
			for k := range root {
				plan.TopLevelKeys = append(plan.TopLevelKeys, k)
			}
			sort.Strings(plan.TopLevelKeys)

			if existing, ok := root["mcp"].(map[string]any); ok {
				for k := range existing {
					plan.ExistingMCPKeys = append(plan.ExistingMCPKeys, k)
				}
				sort.Strings(plan.ExistingMCPKeys)
			} else if _, exists := root["mcp"]; exists {
				warnings = append(warnings, "Top-level 'mcp' exists but is not an object; patch-plan can only propose object-style updates.")
			}

			hasSkills := false
			if _, ok := root["skills"]; ok {
				hasSkills = true
			}
			if _, ok := root["skill"]; ok {
				hasSkills = true
			}
			if hasSkills {
				plan.SkillsDecision = "skill/skills key exists; future schema-aware skill patch can be proposed"
				plan.SkillsDecisionNote = "This command remains read-only and does not emit direct skill writes in this phase."
			} else {
				plan.SkillsDecision = "no skill/skills key found; keep skills in patch-plan notes only"
				plan.SkillsDecisionNote = fmt.Sprintf("Selected skills from bitsentry config: %s", fallback(strings.Join(selectedSkills, ", "), "none"))
			}
		}
	}

	selectedSet := map[string]bool{}
	for _, id := range selectedMCPs {
		selectedSet[strings.TrimSpace(id)] = true
	}

	context7Value := map[string]any{
		"selected": selectedSet["context7"],
		"enabled":  selectedSet["context7"],
		"command":  "npx",
		"args":     []string{"-y", "@upstash/context7-mcp"},
	}
	context7Action := "add"
	if contains(plan.ExistingMCPKeys, "context7") {
		context7Action = "update"
	}
	plan.ProposedMCPChanges = append(plan.ProposedMCPChanges, mcpChange{
		ID:            "context7",
		Action:        context7Action,
		Reason:        "Managed MCP from bitsentry selected/configured set.",
		Note:          "Unknown MCP entries are preserved; only this managed entry would be added/updated.",
		ProposedValue: context7Value,
	})

	engramShouldBeManaged := selectedSet["engram"] || cfg.Components.Engram.Enabled || cfg.Components.Engram.Configured
	if engramShouldBeManaged {
		engramValue := map[string]any{
			"selected":   selectedSet["engram"],
			"enabled":    cfg.Components.Engram.Enabled,
			"configured": cfg.Components.Engram.Configured,
			"command":    "engram",
		}
		if p := strings.TrimSpace(cfg.Components.Engram.BinaryPath); p != "" {
			engramValue["binary_path"] = p
		}
		if p := strings.TrimSpace(cfg.Components.Engram.Project); p != "" {
			engramValue["project"] = p
		}

		engramAction := "add"
		if contains(plan.ExistingMCPKeys, "engram") {
			engramAction = "update"
		}

		plan.ProposedMCPChanges = append(plan.ProposedMCPChanges, mcpChange{
			ID:            "engram",
			Action:        engramAction,
			Reason:        "Managed MCP from bitsentry engram component settings.",
			Note:          "Unknown MCP entries are preserved; only this managed entry would be added/updated.",
			ProposedValue: engramValue,
		})
	}

	if len(plan.TopLevelKeys) == 0 && plan.ParseStatus == "ok" {
		warnings = append(warnings, "opencode.json root is empty object; patch-plan proposes creating/using top-level mcp object while preserving other keys.")
	}

	return plan, warnings
}
