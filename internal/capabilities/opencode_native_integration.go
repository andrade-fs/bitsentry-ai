package capabilities

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type OpenCodeNativeOptions struct {
	RegisterAgent      bool
	InstallCommands    bool
	InstallNativeSkills bool
	ConfigureMCP       bool
}

type OpenCodeNativeResult struct {
	ConfigFile          string
	AgentPromptFile     string
	EntrypointFile      string
	CommandFiles        []string
	NativeSkillFiles    []string
	ConfigBackupPath    string
	NativeBackupPath    string
	Warnings            []string
}

func ExecuteOpenCodeNativeIntegration(projection OpenCodeExportProjection, configRoot string, opts OpenCodeNativeOptions) (OpenCodeNativeResult, error) {
	root := strings.TrimSpace(configRoot)
	if root == "" {
		return OpenCodeNativeResult{}, fmt.Errorf("opencode config root is required")
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return OpenCodeNativeResult{}, err
	}
	res := OpenCodeNativeResult{Warnings: []string{}}

	agentPromptPath := filepath.Join(root, "bitsentry", "agents", "bitsentry.md")
	entrypointPath := filepath.Join(root, "bitsentry", "opencode-entrypoint.md")
	if err := os.MkdirAll(filepath.Dir(agentPromptPath), 0o755); err != nil {
		return res, err
	}
	if err := os.WriteFile(agentPromptPath, []byte(buildBitsentryAgentPrompt()), 0o644); err != nil {
		return res, err
	}
	if err := os.WriteFile(entrypointPath, []byte(buildBitsentryEntrypoint()), 0o644); err != nil {
		return res, err
	}
	res.AgentPromptFile = agentPromptPath
	res.EntrypointFile = entrypointPath

	if opts.InstallCommands {
		cmdFiles, err := writeNativeBitCommands(root)
		if err != nil {
			return res, err
		}
		res.CommandFiles = cmdFiles
	}

	if opts.InstallNativeSkills {
		backupPath, err := backupNativeSkillsRoot(root)
		if err != nil {
			return res, err
		}
		res.NativeBackupPath = backupPath
		skillFiles, err := writeNativeSkills(root, projection)
		if err != nil {
			return res, err
		}
		res.NativeSkillFiles = skillFiles
	}

	if opts.ConfigureMCP {
		res.Warnings = append(res.Warnings, "MCP config mutation is disabled in this phase; skipping MCP updates.")
	}

	if opts.RegisterAgent || opts.InstallCommands {
		configFile, backupPath, err := mergeOpenCodeConfig(root, opts)
		if err != nil {
			return res, err
		}
		res.ConfigFile = configFile
		res.ConfigBackupPath = backupPath
	}

	return res, nil
}

func mergeOpenCodeConfig(root string, opts OpenCodeNativeOptions) (string, string, error) {
	path := filepath.Join(root, "opencode.json")
	obj := map[string]any{}
	backupPath := ""
	if raw, err := os.ReadFile(path); err == nil {
		if len(strings.TrimSpace(string(raw))) > 0 {
			if err := json.Unmarshal(raw, &obj); err != nil {
				return path, "", fmt.Errorf("opencode.json is unparsable; refusing overwrite: %w", err)
			}
		}
		b, err := backupOpenCodeConfigFile(raw)
		if err != nil {
			return path, "", err
		}
		backupPath = b
	}

	instructions := toStringSlice(obj["instructions"])
	entry := "bitsentry/opencode-entrypoint.md"
	if !containsString(instructions, entry) {
		instructions = append(instructions, entry)
	}
	obj["instructions"] = instructions

	agentObj := map[string]any{}
	if v, ok := obj["agent"]; ok {
		if existing, ok := v.(map[string]any); ok {
			agentObj = cloneAnyMap(existing)
		}
	}
	if opts.RegisterAgent {
		agentObj["bitsentry"] = buildBitsentryAgentConfig()
	}
	obj["agent"] = agentObj

	if opts.InstallCommands {
		commands, err := loadAndMigrateCommandMap(obj)
		if err != nil {
			return path, backupPath, err
		}
		for _, cmd := range bitCommandSpecs() {
			commands[cmd.Name] = map[string]any{
				"description": cmd.Description,
				"agent":       "bitsentry",
				"template":    fmt.Sprintf("{file:bitsentry/commands/%s.md}", cmd.Action),
			}
		}
		obj["command"] = commands
	}
	delete(obj, "commands")

	rawOut, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return path, backupPath, err
	}
	rawOut = append(rawOut, '\n')
	if err := os.WriteFile(path, rawOut, 0o644); err != nil {
		return path, backupPath, err
	}
	return path, backupPath, nil
}

func backupOpenCodeConfigFile(raw []byte) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".bitsentry-ai", "backups", "opencode-config", time.Now().UTC().Format("20060102T150405Z"))
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(d, "opencode.json"), raw, 0o644); err != nil {
		return "", err
	}
	return d, nil
}

type bitCommandSpec struct{ Name, Action, Description string }

func bitCommandSpecs() []bitCommandSpec {
	return []bitCommandSpec{
		{"bit-sdd-init", "bit-sdd-init", "Use bitsentry agent with SDD flow at init stage."},
		{"bit-sdd-explore", "bit-sdd-explore", "Use bitsentry agent with SDD flow at explore stage."},
		{"bit-sdd-propose", "bit-sdd-propose", "Use bitsentry agent with SDD flow at propose stage."},
		{"bit-sdd-spec", "bit-sdd-spec", "Use bitsentry agent with SDD flow at spec stage."},
		{"bit-sdd-design", "bit-sdd-design", "Use bitsentry agent with SDD flow at design stage."},
		{"bit-sdd-verify", "bit-sdd-verify", "Use bitsentry agent with SDD flow at verify stage."},
		{"bit-sdr-capture", "bit-sdr-capture", "Use bitsentry agent with SDR capture stage."},
		{"bit-sdr-triage", "bit-sdr-triage", "Use bitsentry agent with SDR triage guidance."},
		{"bit-sdr-enrich", "bit-sdr-enrich", "Use bitsentry agent with SDR enrichment guidance."},
		{"bit-support-triage", "bit-support-triage", "Use bitsentry agent with support triage guidance."},
		{"bit-support-handoff", "bit-support-handoff", "Use bitsentry agent with support handoff guidance."},
		{"bit-pack-status", "bit-pack-status", "Report Bitsentry pack installation status."},
		{"bit-install-check", "bit-install-check", "Check Bitsentry OpenCode installation."},
	}
}

func writeNativeBitCommands(root string) ([]string, error) {
	dstDir := filepath.Join(root, "bitsentry", "commands")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return nil, err
	}
	files := []string{}
	for _, c := range bitCommandSpecs() {
		if !strings.HasPrefix(c.Name, "bit-") {
			return nil, fmt.Errorf("invalid command name %q", c.Name)
		}
		if strings.Contains(c.Name, "..") || strings.Contains(c.Name, "/") || strings.Contains(c.Action, "..") || strings.Contains(c.Action, "/") {
			return nil, fmt.Errorf("unsafe command/action name %q/%q", c.Name, c.Action)
		}
		if strings.Contains(c.Name, "orchestrator") {
			return nil, fmt.Errorf("forbidden orchestrator command %q", c.Name)
		}
		content := strings.Join([]string{
			"# /" + c.Name,
			"",
			"Use the Bitsentry agent and managed pack:",
			"- bitsentry/OPENCODE_USAGE.md",
			"- bitsentry/skill-registry.md",
			"",
			c.Description,
			"",
			"Constraints:",
			"- Do not modify opencode.json.",
			"- Do not execute runtime flows.",
			"- Return structured output (flow, skills, brief, goals, non-goals, handoff, risks, verdict).",
		}, "\n") + "\n"
		p := filepath.Join(dstDir, c.Action+".md")
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			return nil, err
		}
		files = append(files, p)
	}
	sort.Strings(files)
	return files, nil
}

func backupNativeSkillsRoot(root string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(home, ".bitsentry-ai", "backups", "opencode-native", time.Now().UTC().Format("20060102T150405Z"))
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	skillsRoot := filepath.Join(root, "skills")
	if st, err := os.Stat(skillsRoot); err == nil && st.IsDir() {
		_ = copyDir(skillsRoot, filepath.Join(d, "skills"))
	}
	return d, nil
}

func writeNativeSkills(root string, projection OpenCodeExportProjection) ([]string, error) {
	mapping := map[string]string{
		"bitsentry-sdd-init":      "sdd/sdd-init",
		"bitsentry-sdd-explore":   "sdd/sdd-explore",
		"bitsentry-sdd-propose":   "sdd/sdd-propose",
		"bitsentry-sdd-spec":      "sdd/sdd-spec",
		"bitsentry-sdd-design":    "sdd/sdd-design",
		"bitsentry-sdd-apply":     "sdd/sdd-apply",
		"bitsentry-sdd-verify":    "sdd/sdd-verify",
		"bitsentry-sdr-capture":   "sdr/sdr-capture",
		"bitsentry-sdr-triage":    "sdr/sdr-questions",
		"bitsentry-sdr-enrich":    "sdr/sdr-research",
		"bitsentry-support-triage": "support/issue-creation",
		"bitsentry-support-handoff": "support/branch-pr",
	}
	srcByID := map[string]ProjectedSkill{}
	for _, s := range projection.IncludedSkills {
		srcByID[s.ID] = s
	}
	out := []string{}
	for native, srcID := range mapping {
		if strings.Contains(native, "..") || strings.Contains(native, "/") {
			return nil, fmt.Errorf("unsafe native skill name %q", native)
		}
		src, ok := srcByID[srcID]
		if !ok {
			continue
		}
		raw, err := os.ReadFile(src.SourcePath)
		if err != nil {
			return nil, err
		}
		body := string(raw)
		description := fmt.Sprintf("Bitsentry native skill for %s. References bitsentry/OPENCODE_USAGE.md and bitsentry/skill-registry.md.", srcID)
		content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n> Uses managed pack references:\n> - bitsentry/OPENCODE_USAGE.md\n> - bitsentry/skill-registry.md\n\n%s", native, description, body)
		target := filepath.Join(root, "skills", native, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
			return nil, err
		}
		out = append(out, target)
	}
	sort.Strings(out)
	return out, nil
}

func buildBitsentryEntrypoint() string {
	return strings.Join([]string{
		"# Bitsentry OpenCode Entrypoint",
		"",
		"Bitsentry capability pack is installed.",
		"Native agent `bitsentry` is available.",
		"Use bitsentry for SDD/SDR/support/capability-pack workflows.",
		"Use /bit-* commands for direct actions.",
		"Read bitsentry/OPENCODE_USAGE.md and bitsentry/skill-registry.md when needed.",
	}, "\n") + "\n"
}

func buildBitsentryAgentPrompt() string {
	return strings.Join([]string{
		"# Bitsentry Agent",
		"",
		"You are the Bitsentry orchestrator agent.",
		"References:",
		"- bitsentry/OPENCODE_USAGE.md",
		"- bitsentry/skill-registry.md",
		"- bitsentry/flows/",
		"- bitsentry/skills/",
		"",
		"Routing:",
		"- SDD requests -> SDD flow",
		"- SDR/security/research requests -> SDR flow",
		"- support/ops/help requests -> support flow",
		"- install/config/capabilities -> pack/install guidance",
		"",
		"Note: sdd-orchestrator/sdr-orchestrator are internal routing concepts, not direct slash commands.",
		"",
		"Boundaries:",
		"- do not modify opencode.json unless explicitly requested",
		"- do not execute runtime flows",
		"- do not run autonomous actions",
		"- do not modify code unless explicitly requested",
		"",
		"Return structured outputs:",
		"- selected flow",
		"- selected skills",
		"- brief",
		"- goals",
		"- non-goals",
		"- handoff sequence",
		"- risks",
		"- verdict",
	}, "\n") + "\n"
}

func toStringSlice(v any) []string {
	if v == nil {
		return []string{}
	}
	if s, ok := v.([]string); ok {
		return append([]string{}, s...)
	}
	if arr, ok := v.([]any); ok {
		out := make([]string, 0, len(arr))
		for _, x := range arr {
			if t, ok := x.(string); ok {
				out = append(out, t)
			}
		}
		return out
	}
	return []string{}
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := map[string]any{}
	for k, v := range in {
		out[k] = v
	}
	return out
}

func containsString(values []string, target string) bool {
	for _, v := range values {
		if strings.TrimSpace(v) == target {
			return true
		}
	}
	return false
}

func buildBitsentryAgentConfig() map[string]any {
	return map[string]any{
		"description": "Bitsentry orchestrator for SDD/SDR/support/capability-pack workflows.",
		"mode":        "primary",
		"prompt":      "{file:bitsentry/agents/bitsentry.md}",
		"permission": map[string]any{
			"edit": "ask",
			"bash": "ask",
		},
	}
}

func isBitsentryCommandKey(k string) bool {
	t := strings.TrimSpace(k)
	t = strings.TrimPrefix(t, "/")
	return strings.HasPrefix(t, "bit-")
}

func loadAndMigrateCommandMap(obj map[string]any) (map[string]any, error) {
	command := map[string]any{}
	if v, ok := obj["command"]; ok {
		if existing, ok := v.(map[string]any); ok {
			command = cloneAnyMap(existing)
		}
	}
	if v, ok := obj["commands"]; ok {
		existing, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("top-level commands exists but is not an object")
		}
		for k := range existing {
			if !isBitsentryCommandKey(k) {
				return nil, fmt.Errorf("top-level commands contains non-bitsentry entry %q; refusing silent migration", k)
			}
		}
		for k, v := range existing {
			command[k] = v
		}
	}
	migrated := map[string]any{}
	for k, v := range command {
		nk := strings.TrimPrefix(strings.TrimSpace(k), "/")
		entry := map[string]any{}
		if m, ok := v.(map[string]any); ok {
			entry = cloneAnyMap(m)
		} else {
			migrated[nk] = v
			continue
		}
		if p, ok := entry["prompt"].(string); ok {
			entry["template"] = normalizeTemplateFromPrompt(p)
			delete(entry, "prompt")
		}
		migrated[nk] = entry
	}
	return migrated, nil
}

func normalizeTemplateFromPrompt(prompt string) string {
	p := strings.TrimSpace(prompt)
	if strings.HasPrefix(p, "{file:") {
		return p
	}
	if strings.HasPrefix(p, "bitsentry/commands/") {
		return fmt.Sprintf("{file:%s}", p)
	}
	return p
}
