package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/system"
	"github.com/spf13/cobra"
)

type mvpReadiness struct {
	OpenCodeDetected      bool
	OpenCodeConfigRoot    string
	BitsentryPackStatus   string
	NativeAgentStatus     string
	CommandsStatus        string
	SecurityFlowsStatus   string
	EditPermissionStatus  string
	MCPPreviewOnlyStatus  string
	Result                string
	Notes                 []string
}

func newDoctorCmd(rt *Runtime) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Inspect local environment and dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := rt.App.ConfigManager.Load()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			info := system.DetectSystem()
			shell := system.DetectShell()
			deps := system.CheckDependencies()

			pkgMgr := "not detected"
			for _, d := range deps {
				if !d.Found {
					continue
				}
				switch d.Name {
				case "brew", "apt", "yum", "pacman":
					pkgMgr = fmt.Sprintf("%s (%s)", d.Name, d.Path)
				}
			}

			readiness := detectMVPReadiness()

			out := cmd.OutOrStdout()
			fmt.Fprintln(out, "Doctor report")
			fmt.Fprintf(out, "- OS: %s\n", info.OS)
			fmt.Fprintf(out, "- Arch: %s\n", info.Arch)
			fmt.Fprintf(out, "- Shell: %s\n", shell)
			fmt.Fprintf(out, "- Package manager: %s\n", pkgMgr)
			fmt.Fprintf(out, "- Config dir: %s\n", filepath.Dir(rt.App.ConfigManager.Path()))
			fmt.Fprintf(out, "- Config file: %s\n", rt.App.ConfigManager.Path())
			fmt.Fprintf(out, "- Active profile: %s\n", cfg.ActiveProfile)
			fmt.Fprintln(out, "- Dependencies:")
			for _, d := range deps {
				status := "not found"
				if d.Found {
					status = "found"
				}
				mandatory := "optional"
				if d.Mandatory {
					mandatory = "required"
				}
				fmt.Fprintf(out, "  - %s: %s (%s)", d.Name, status, mandatory)
				if d.Path != "" {
					fmt.Fprintf(out, " [%s]", d.Path)
				}
				fmt.Fprintln(out)
			}

			fmt.Fprintln(out, "- MVP readiness (OpenCode/TUI-first, CLI for support/debug/status):")
			fmt.Fprintf(out, "  - OpenCode detected/config root: %s / %s\n", yesNoLabel(readiness.OpenCodeDetected), valueOrDefault(readiness.OpenCodeConfigRoot, "not resolved"))
			fmt.Fprintf(out, "  - Bitsentry pack status: %s\n", readiness.BitsentryPackStatus)
			fmt.Fprintf(out, "  - Native agent status: %s\n", readiness.NativeAgentStatus)
			fmt.Fprintf(out, "  - Commands status: %s\n", readiness.CommandsStatus)
			fmt.Fprintf(out, "  - Security flows availability: %s\n", readiness.SecurityFlowsStatus)
			fmt.Fprintf(out, "  - Edit permission deny: %s\n", readiness.EditPermissionStatus)
			fmt.Fprintf(out, "  - MCP config: %s\n", readiness.MCPPreviewOnlyStatus)
			if len(readiness.Notes) > 0 {
				fmt.Fprintln(out, "  - Notes:")
				for _, n := range readiness.Notes {
					fmt.Fprintf(out, "    - %s\n", n)
				}
			}
			fmt.Fprintf(out, "- Result: %s\n", readiness.Result)

			return nil
		},
	}
}

func detectMVPReadiness() mvpReadiness {
	status := detectOpenCodeInstallStatusForDoctor()
	r := mvpReadiness{
		OpenCodeDetected:     status.Detected,
		OpenCodeConfigRoot:   status.ConfigRoot,
		BitsentryPackStatus:  boolWord(status.PackInstalled, "installed", "not installed"),
		NativeAgentStatus:    "not installed",
		CommandsStatus:       "not installed",
		SecurityFlowsStatus:  "not available",
		EditPermissionStatus: "not verifiable",
		MCPPreviewOnlyStatus: "preview-only contract available",
		Result:               "PASS",
		Notes:                []string{},
	}

	if !status.Detected {
		r.Notes = append(r.Notes, "OpenCode binary was not detected on PATH.")
	}
	if strings.TrimSpace(status.ConfigRoot) == "" {
		r.Notes = append(r.Notes, "OpenCode config root is not resolved.")
	}

	if strings.TrimSpace(status.ConfigRoot) != "" {
		preview := capabilities.BuildOpenCodeMCPConfigPreview(status.ConfigRoot, nil)
		r.MCPPreviewOnlyStatus = fmt.Sprintf("PREVIEW ONLY (%s)", valueOrDefault(preview.CurrentConfigState, "unknown"))
		agentStatus, commandStatus, editStatus, parseErr := inspectOpenCodeNativeStatus(status.ConfigRoot)
		r.NativeAgentStatus = agentStatus
		r.CommandsStatus = commandStatus
		r.EditPermissionStatus = editStatus
		if parseErr != nil {
			r.Notes = append(r.Notes, fmt.Sprintf("Could not fully inspect opencode.json: %v", parseErr))
		}
	}

	if doctorFileExists(filepath.Join("assets", "flows", "source-security-review.yaml")) && doctorFileExists(filepath.Join("assets", "flows", "web-assessment.yaml")) {
		r.SecurityFlowsStatus = "declarative manifests available (manual/non-runtime in MVP)"
	} else {
		r.SecurityFlowsStatus = "one or more security flow manifests are missing"
		r.Notes = append(r.Notes, "Expected security flow manifests are not complete.")
	}

	r.Result = readinessVerdict(r)
	return r
}

func readinessVerdict(r mvpReadiness) string {
	if !r.OpenCodeDetected || strings.TrimSpace(r.OpenCodeConfigRoot) == "" {
		return "FAIL"
	}
	if r.BitsentryPackStatus != "installed" {
		return "PASS WITH NOTES"
	}
	if strings.HasPrefix(r.EditPermissionStatus, "missing") || strings.HasPrefix(r.EditPermissionStatus, "invalid") {
		return "PASS WITH NOTES"
	}
	if len(r.Notes) > 0 {
		return "PASS WITH NOTES"
	}
	return "PASS"
}

type openCodeInstallStatusDoctor struct {
	Detected      bool
	ConfigRoot    string
	PackInstalled bool
}

func detectOpenCodeInstallStatusForDoctor() openCodeInstallStatusDoctor {
	ctx := context.Background()
	res, _ := agents.OpenCodeDetector{}.Detect(ctx)
	wd, _ := os.Getwd()
	root := ""
	found := agents.ExistingOpenCodeConfigPaths(wd)
	if len(found) > 0 {
		root = found[0]
	} else {
		candidates := agents.OpenCodeConfigPathCandidates(wd)
		if len(candidates) > 0 {
			root = candidates[0]
		}
	}
	packInstalled := false
	if strings.TrimSpace(root) != "" {
		packRoot := filepath.Join(root, "bitsentry")
		packInstalled = doctorFileExists(filepath.Join(packRoot, "OPENCODE_USAGE.md")) && doctorFileExists(filepath.Join(packRoot, "skill-registry.md"))
	}
	return openCodeInstallStatusDoctor{Detected: res.Found, ConfigRoot: root, PackInstalled: packInstalled}
}

func inspectOpenCodeNativeStatus(root string) (agent string, commands string, editPermission string, parseErr error) {
	path := filepath.Join(root, "opencode.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return "not installed", "not installed", "not verifiable", err
	}
	var cfg map[string]any
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return "not installed", "not installed", "not verifiable", err
	}

	agent = "not installed"
	editPermission = "not verifiable"
	if agentsMap, ok := cfg["agent"].(map[string]any); ok {
		if bitsentry, ok := agentsMap["bitsentry"].(map[string]any); ok {
			agent = "installed"
			if perm, ok := bitsentry["permission"].(map[string]any); ok {
				if edit, ok := perm["edit"].(string); ok {
					if edit == "deny" {
						editPermission = "agent.bitsentry.permission.edit=deny"
					} else {
						editPermission = fmt.Sprintf("invalid value (%s)", edit)
					}
				} else {
					editPermission = "missing edit permission"
				}
			} else {
				editPermission = "missing permission block"
			}
		}
	}

	commands = "not installed"
	if cmdMap, ok := cfg["commands"].(map[string]any); ok {
		count := 0
		for k := range cmdMap {
			if strings.HasPrefix(k, "bit-") {
				count++
			}
		}
		if count > 0 {
			commands = fmt.Sprintf("installed (%d bit-* commands)", count)
		}
	}

	return agent, commands, editPermission, nil
}

func doctorFileExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !st.IsDir()
}

func boolWord(v bool, yes, no string) string {
	if v {
		return yes
	}
	return no
}

func yesNoLabel(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

func valueOrDefault(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
