package components

import (
	"context"
	"os/exec"
	"sort"
	"strings"

	"bitsentry-ai/internal/config"
)

type MCP struct {
	ID          string
	Name        string
	Description string
	Category    string
	Command     string
	Args        []string
	Package     string
	Enabled     bool
	Configured  bool
	Required    bool
	Status      Status
	Readiness   MCPReadiness
	Notes       string
}

type MCPRegistrySummary struct {
	Enabled    bool
	Configured bool
	Selected   []string
	Total      int
}

func MCPRegistry(_ context.Context, cfg config.Config, engramDetails EngramRuntimeDetails, context7Details Context7RuntimeDetails) ([]MCP, MCPRegistrySummary) {
	mcpCfg := cfg.Components.MCPs
	selected := make([]string, 0, len(mcpCfg.Selected))
	for _, v := range mcpCfg.Selected {
		trimmed := strings.TrimSpace(v)
		if trimmed != "" {
			selected = append(selected, trimmed)
		}
	}

	selectedSet := map[string]bool{}
	for _, id := range selected {
		selectedSet[id] = true
	}

	entries := []MCP{
		{
			ID:          "context7",
			Name:        "Context7",
			Description: "Docs/research MCP metadata entry.",
			Category:    "docs/research",
			Command:     "npx",
			Args:        []string{"-y", "@upstash/context7-mcp"},
			Package:     "@upstash/context7-mcp",
			Required:    false,
			Status:      context7Details.Status,
			Readiness:   context7Details.Readiness,
			Notes:       "Status derived from Context7 component config/runtime when available.",
		},
		{
			ID:          "engram",
			Name:        "Engram",
			Description: "Memory MCP metadata entry.",
			Category:    "memory",
			Command:     "engram",
			Required:    false,
			Status:      engramDetails.Status,
			Readiness:   engramDetails.Readiness,
			Notes:       "Status derived from Engram component config/runtime when available.",
		},
		{
			ID:          "filesystem",
			Name:        "Filesystem",
			Description: "Local filesystem MCP metadata entry.",
			Category:    "local",
			Status:      StatusModeledOnly,
			Readiness:   BuildMCPReadiness(StatusModeledOnly, []string{"catalog entry only"}, []string{"runtime probing not implemented"}, []string{"Use manual local validation in target agent"}, false),
			Notes:       "Modeled only; no agent config is written.",
		},
		{
			ID:          "git",
			Name:        "Git",
			Description: "Git tooling MCP metadata entry.",
			Category:    "development",
			Command:     "git",
			Status:      detectPathCommandStatus("git"),
			Readiness:   BuildMCPReadiness(detectPathCommandStatus("git"), []string{"git runtime check via PATH"}, nil, []string{"No credential validation is performed"}, detectPathCommandStatus("git") == StatusDetected),
			Notes:       "Status is based on whether git is present in PATH.",
		},
		{
			ID:          "github",
			Name:        "GitHub",
			Description: "GitHub MCP metadata entry.",
			Category:    "development",
			Status:      StatusModeledOnly,
			Readiness:   BuildMCPReadiness(StatusModeledOnly, []string{"catalog entry only"}, []string{"token/config validation not implemented"}, []string{"Configure manually in OpenCode when needed"}, false),
			Notes:       "Modeled only; token/config validation is future work.",
		},
		{
			ID:          "postgres",
			Name:        "Postgres",
			Description: "PostgreSQL MCP metadata entry.",
			Category:    "database",
			Status:      StatusModeledOnly,
			Readiness:   BuildMCPReadiness(StatusModeledOnly, []string{"capability planning entry"}, []string{"adapter write support not implemented"}, []string{"Manual MCP setup required outside this phase"}, false),
			Notes:       "Modeled for capability planning; adapter write support is not implemented yet.",
		},
		{
			ID:          "browser",
			Name:        "Browser",
			Description: "Browser automation MCP metadata entry.",
			Category:    "automation",
			Status:      StatusNotImplemented,
			Readiness:   BuildMCPReadiness(StatusNotImplemented, nil, []string{"not implemented"}, []string{"Wait for future phase implementation"}, false),
			Notes:       "Reserved for future implementation.",
		},
		{
			ID:          "firecrawl",
			Name:        "Firecrawl",
			Description: "Research crawling MCP metadata entry.",
			Category:    "research",
			Status:      StatusModeledOnly,
			Readiness:   BuildMCPReadiness(StatusModeledOnly, []string{"catalog entry only"}, []string{"API key validation not implemented"}, []string{"Manual credential setup required; no secrets are read/written here"}, false),
			Notes:       "Modeled only; API key validation is not implemented yet.",
		},
	}

	for i := range entries {
		entries[i].Enabled = mcpCfg.Enabled && selectedSet[entries[i].ID]
		entries[i].Configured = mcpCfg.Configured && selectedSet[entries[i].ID]
	}

	sort.Strings(selected)
	summary := MCPRegistrySummary{
		Enabled:    mcpCfg.Enabled,
		Configured: mcpCfg.Configured,
		Selected:   selected,
		Total:      len(entries),
	}

	return entries, summary
}

func detectPathCommandStatus(command string) Status {
	if strings.TrimSpace(command) == "" {
		return StatusUnknown
	}
	if _, err := exec.LookPath(command); err == nil {
		return StatusDetected
	}
	return StatusMissing
}
