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
	RoleFiles           []string
	IntentFiles         []string
	ConfigBackupPath    string
	NativeBackupPath    string
	Warnings            []string
}

type OpenCodeMCPConfigPreview struct {
	CurrentConfigState       string   `json:"current_config_state"`
	Exists                   bool     `json:"exists"`
	Readable                 bool     `json:"readable"`
	InvalidError             string   `json:"invalid_error,omitempty"`
	CurrentMCPConfigDetected bool     `json:"current_mcp_config_detected"`
	MCPReadinessState        string   `json:"mcp_readiness_state"`
	ProposedSafeChanges      []string `json:"proposed_safe_changes"`
	PreservedKeys            []string `json:"preserved_keys"`
	PreservedMCPEntries      []string `json:"preserved_mcp_entries"`
	Warnings                 []string `json:"warnings"`
	ManualSteps              []string `json:"manual_steps"`
	WouldWrite               bool     `json:"would_write"`
	RequiresConfirmation     bool     `json:"requires_confirmation"`
	BackupRequired           bool     `json:"backup_required"`
}

func BuildOpenCodeMCPConfigPreview(configRoot string, selectedMCPs []string) OpenCodeMCPConfigPreview {
	preview := OpenCodeMCPConfigPreview{
		CurrentConfigState:       "missing",
		Exists:                   false,
		Readable:                 false,
		CurrentMCPConfigDetected: false,
		MCPReadinessState:        "manual_step_needed",
		ProposedSafeChanges:      []string{},
		PreservedKeys:            []string{},
		PreservedMCPEntries:      []string{},
		Warnings:                 []string{},
		ManualSteps:              []string{},
		WouldWrite:               false,
		RequiresConfirmation:     true,
		BackupRequired:           true,
	}

	root := strings.TrimSpace(configRoot)
	if root == "" {
		preview.CurrentConfigState = "missing_config_root"
		preview.Warnings = append(preview.Warnings, "OpenCode config root is not resolved.")
		preview.ManualSteps = append(preview.ManualSteps,
			"Resolve OpenCode config root before any MCP apply.",
			"Do not add credentials in this preview phase; configure sensitive values manually.",
		)
		return preview
	}

	path := filepath.Join(root, "opencode.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		preview.CurrentConfigState = "missing_opencode_json"
		preview.Warnings = append(preview.Warnings, "opencode.json not found or not readable.")
		preview.ManualSteps = append(preview.ManualSteps,
			"Create/repair opencode.json manually.",
			"Future apply must require explicit confirmation and backup before any write.",
		)
		preview.proposeSafeMCPChanges(selectedMCPs)
		return preview
	}

	preview.Exists = true
	preview.Readable = true

	obj := map[string]any{}
	if len(strings.TrimSpace(string(raw))) == 0 {
		preview.CurrentConfigState = "readable_empty"
		preview.Warnings = append(preview.Warnings, "opencode.json is empty.")
		preview.ManualSteps = append(preview.ManualSteps, "Add valid JSON object to opencode.json before any MCP apply.")
		preview.proposeSafeMCPChanges(selectedMCPs)
		return preview
	}

	if err := json.Unmarshal(raw, &obj); err != nil {
		preview.Readable = false
		preview.CurrentConfigState = "invalid"
		preview.InvalidError = err.Error()
		preview.Warnings = append(preview.Warnings, "Invalid opencode.json; preview cannot safely parse MCP section.")
		preview.ManualSteps = append(preview.ManualSteps,
			"Fix JSON syntax manually.",
			"Re-run preview and confirm backup requirement before future apply.",
		)
		preview.proposeSafeMCPChanges(selectedMCPs)
		return preview
	}

	preview.CurrentConfigState = "readable"
	for k := range obj {
		preview.PreservedKeys = append(preview.PreservedKeys, k)
	}
	sort.Strings(preview.PreservedKeys)

	if mcpObj, ok := obj["mcp"].(map[string]any); ok {
		preview.CurrentMCPConfigDetected = true
		for k := range mcpObj {
			preview.PreservedMCPEntries = append(preview.PreservedMCPEntries, k)
		}
		sort.Strings(preview.PreservedMCPEntries)
	} else if _, has := obj["mcp"]; has {
		preview.CurrentMCPConfigDetected = true
		preview.Warnings = append(preview.Warnings, "Top-level mcp exists but is not an object; manual normalization required.")
		preview.ManualSteps = append(preview.ManualSteps, "Normalize opencode.json.mcp to an object manually.")
	}

	if preview.CurrentMCPConfigDetected {
		preview.CurrentConfigState = "readable_with_mcp"
	} else {
		preview.CurrentConfigState = "readable_without_mcp"
	}

	preview.proposeSafeMCPChanges(selectedMCPs)
	preview.deriveReadinessState()
	if len(preview.ManualSteps) == 0 {
		preview.ManualSteps = append(preview.ManualSteps,
			"This is PREVIEW ONLY. Future apply must be explicit and create backup first.",
		)
	}
	return preview
}

func (p *OpenCodeMCPConfigPreview) proposeSafeMCPChanges(selectedMCPs []string) {
	set := map[string]bool{}
	for _, id := range selectedMCPs {
		t := strings.TrimSpace(id)
		if t != "" {
			set[t] = true
		}
	}
	if set["engram"] {
		p.ProposedSafeChanges = append(p.ProposedSafeChanges, "Ensure mcp.engram metadata entry exists without sensitive values (command-only metadata).")
		p.ManualSteps = append(p.ManualSteps, "Configure any Engram-sensitive values manually outside preview.")
	}
	if set["context7"] {
		p.ProposedSafeChanges = append(p.ProposedSafeChanges, "Ensure mcp.context7 metadata entry exists without sensitive values (command/args only).")
		p.ManualSteps = append(p.ManualSteps, "Configure Context7 token/API key manually outside preview.")
	}
	if len(p.ProposedSafeChanges) == 0 {
		p.ProposedSafeChanges = append(p.ProposedSafeChanges, "No MCP selected; no config changes proposed.")
	}
	sort.Strings(p.ProposedSafeChanges)
	p.ManualSteps = uniqueSortedPreviewStrings(p.ManualSteps)
}

func (p *OpenCodeMCPConfigPreview) deriveReadinessState() {
	if p.CurrentMCPConfigDetected && p.Readable {
		p.MCPReadinessState = "configured"
		return
	}
	if p.Readable {
		p.MCPReadinessState = "detected"
		return
	}
	if p.Exists {
		p.MCPReadinessState = "manual_step_needed"
		return
	}
	p.MCPReadinessState = "modeled_only"
}

func uniqueSortedPreviewStrings(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	seen := map[string]bool{}
	out := []string{}
	for _, v := range in {
		t := strings.TrimSpace(v)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
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

	roleFiles, err := writeNativeRoles(root, projection)
	if err != nil {
		return res, err
	}
	res.RoleFiles = roleFiles
	intentFiles, err := writeNativeIntents(root, projection)
	if err != nil {
		return res, err
	}
	res.IntentFiles = intentFiles

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
		for _, role := range knownRoleIDs() {
			agentObj[role] = buildBitsentrySubagentConfig(role)
		}
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
		content := buildNativeBitCommandContent(c)
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
		"This OpenCode setup includes the native `bitsentry` agent and managed Bitsentry capability pack.",
		"",
		"Start here:",
		"1) Activate agent: `@bitsentry`",
		"2) State intent: SDD (feature/design/implementation), SDR (research/debug), or support (handoff/status/close)",
		"3) Use `/bit-*` commands when you want guided shortcuts",
		"",
		"Primary references:",
		"- bitsentry/agents/bitsentry.md",
		"- bitsentry/commands/",
		"- bitsentry/OPENCODE_USAGE.md",
		"- bitsentry/skill-registry.md",
		"",
		"Safety boundaries:",
		"- BitsentryAI is installer/projector/validator/pack manager, NOT runtime flow executor.",
		"- Do not change repository code unless the user explicitly requests implementation.",
		"- Do not mutate opencode.json unless explicitly requested.",
		"- Do not claim memory persistence unless Engram (or another memory backend) is detected/configured.",
		"",
		"If tools/MCPs are unavailable:",
		"- report as AVAILABLE / CONFIGURED / MISSING CREDENTIALS / MANUAL STEP NEEDED / UNSUPPORTED",
		"- provide exact manual next step instead of pretending tool execution.",
	}, "\n") + "\n"
}

func buildBitsentryAgentPrompt() string {
	return strings.Join([]string{
		"# Bitsentry Agent",
		"",
		"## Non-negotiable first response rules",
		"- Apply Compact Direct Mode & Agent Token Efficiency contract before composing the first visible response.",
		"- For open/ambiguous/exploratory/strategic/planning/architecture/security/multi-file/sensitive requests, your FIRST visible response MUST be a Full Route Decision Preview.",
		"- For direct/atomic/concrete/already-intentional requests with clear implicit intent, use Compact Direct Mode: concise execution/analysis-oriented response WITHOUT full Route Decision Preview table and WITHOUT full Capability Preview table.",
		"- For direct requests that are sensitive/remote/destructive/ambiguous, keep Compact Direct Mode but require a brief explicit confirmation before mutation-risk actions.",
		"- Compact Direct Mode does NOT remove internal reasoning quality; it only reduces verbose route/capability verbalization when it adds no value.",
		"- Required first response template:",
		"  Route Decision Preview",
		"  - matched_intent:",
		"  - decision:",
		"  - matched_signals:",
		"  - primary_skills:",
		"  - secondary_skills:",
		"  - deferred_skills:",
		"  - primary_roles:",
		"  - secondary_roles:",
		"  - capability_reason:",
		"  - capability_gates:",
		"  - reason:",
		"  - requires_bounded_discovery:",
		"  - requires_confirmation:",
		"  - gates:",
		"    - no_edits_in_preview",
		"    - no_persistence_in_preview",
		"    - no_flow_execution_in_preview",
		"- The Route Decision Preview template is mandatory when Full Preview applies; it is intentionally omitted in Compact Direct Mode.",
		"- Only AFTER showing Route + Capability Decision Preview may you inspect files, analyze repo context, ask clarifying questions, recommend flows/capabilities, or propose changes.",
		"- Only AFTER showing Route + Capability Decision Preview (when Full Preview applies) may you inspect files, analyze repo context, ask clarifying questions, recommend flows/capabilities, or propose changes.",
		"- Exception: if the request is a clearly simple conceptual question mapped to direct_answer and does not need repository context, respond directly.",
		"- Do not confuse 'small scope' with direct_answer: small copy/product-messaging/UX changes may still require compact SDD when they affect core narrative, product perception, multi-component consistency, or implementation planning.",
		"",
		"You are the OpenCode-native Bitsentry orchestrator.",
		"BitsentryAI is a local capability installer/projector/verifier/pack manager.",
		"It is NOT a standalone runtime orchestrator.",
		"References:",
		"- bitsentry/OPENCODE_USAGE.md",
		"- bitsentry/skill-registry.md",
		"- bitsentry/flows/",
		"- bitsentry/skills/",
		"",
		"Intent routing:",
		"- Intent Decision Contract (choose exactly one first action): direct_answer | use_skill | use_role | use_flow_sdr | use_flow_support | use_flow_sdd | use_flow_source-security-review | use_flow_web-assessment | bounded_discovery_then_decide | ask_clarifying_question",
		"- SDD: feature work, implementation planning, product changes, architecture/design changes.",
		"- SDR: research, debugging, investigation, repo analysis, technical diagnosis.",
		"- Source Security Review: source-only AppSec review flow (read-only-first, no pentest/runtime).",
		"- Web Assessment: live/public web target assessment route ONLY when authorization + exact scope gates are explicitly satisfied; no runtime/tooling execution in this phase.",
		"- Support: handoff, closing session, summaries, memory-save blocks, status/install checks.",
		"- Capability pack: install/config/capabilities/tooling guidance.",
		"- Direct reasoning: trivial/small tasks where formal flow is overkill.",
		"- Questions simples/explanations => direct_answer (NO SDD by default).",
		"",
		"Route decision first (when no flow active):",
		"- OpenCode-native route decision preview is the PRIMARY route-selection UX inside @bitsentry conversations",
		"- CLI `bitsentry-ai route decide` is debug/plumbing parity, not the primary end-user workflow",
		"- Full Preview trigger classes: open, ambiguous, exploratory, planning, architecture, security/risk, multi-file change, sensitive/high-impact operations",
		"- Compact Direct Mode trigger classes: direct, atomic, concrete, clearly-intentional requests where expanded route/capability tables provide low value",
		"- briefly classify request into SDD / SDR / Support / Direct reasoning",
		"- explain why in 1-2 lines",
		"- recommend a route (compact SDD or direct) proportional to task size",
		"- if ambiguous, ask user to choose before entering a flow",
		"- do NOT force SDD automatically",
		"- do NOT activate any formal flow silently; announce selected flow first",
		"- bounded context discovery is allowed only AFTER a visible Route Decision Preview, when the preview sets requires_bounded_discovery=true",
		"- bounded discovery limits: read-only only, no edits, no todos/tasks, no persistence writes, no implementation decisions",
		"- preview gates are mandatory: no_edits_in_preview, no_persistence_in_preview, no_flow_execution_in_preview",
		"- for narrow direct requests (e.g., 'open file X and show Y'), inspection is allowed if explicitly requested",
		"- default broad landing/product-copy consistency route: SDR brief audit -> compact SDD corrections",
		"- Route decision MUST be visible in chat BEFORE any non-trivial discovery",
		"- After discovery, show an updated route decision only if findings change the route, capabilities, gates, or confirmation requirements.",
		"- post-discovery gate is mandatory: STOP after discovery and ask confirmation before route selection, edit planning, apply, or persistence",
		"- forbidden without approval: statements like 'procedo directamente con las ediciones' or equivalent apply-now phrasing",
		"- ask separate permission before apply/edit planning and before persistence",
		"- ask_clarifying_question is allowed only when route cannot be decided with high confidence",
		"- include matched_signals in route preview reasoning when available",
		"- for web-assessment recommendation, require explicit authorization + exact scope + environment + testing window + intensity + tool allowlist BEFORE any request/tooling",
		"",
		"## Visible Route Decision Preview Requirement",
		"- Before any non-trivial repository discovery, analysis, planning, flow recommendation, capability recommendation, or implementation proposal, you MUST show a concise Route Decision Preview visible to the user WHEN Full Preview applies.",
		"- The preview must include: matched_intent, decision, recommended_flow/recommended_roles/recommended_skills (when applicable), primary_skills/secondary_skills/deferred_skills, primary_roles/secondary_roles, capability_reason, capability_gates, matched_signals (when available), confidence (when available), reason, requires_confirmation, requires_bounded_discovery, and gates.",
		"- Required preview gates: no_edits_in_preview, no_persistence_in_preview, no_flow_execution_in_preview.",
		"- In Compact Direct Mode, do not print full Route Decision Preview/Capability Preview blocks; provide a short direct decision line and required confirmation gate when risk-sensitive.",
		"- After showing the preview, bounded read-only discovery is allowed only when the selected decision allows it.",
		"- Never skip visible route preview for non-trivial repository/product/code/documentation/UX/architecture/security/bug requests.",
		"- Exception: for simple conceptual questions that clearly map to direct_answer and do not require repository context, answer directly without activating flows.",
		"",
		"Note: sdd-orchestrator/sdr-orchestrator are internal routing concepts, not direct slash commands.",
		"",
		"Boundaries:",
		"- bitsentry is a non-mutating orchestrator by default",
		"- bitsentry must not edit files directly",
		"- implementation/apply requires an explicit approved apply path (subagent/future mechanism), not direct edits from bitsentry",
		"- no internal Bitsentry runtime/session execution",
		"- do not modify opencode.json unless explicitly requested",
		"- do not execute runtime flows",
		"- do not run autonomous actions",
		"- do not modify code unless explicitly requested",
		"- if user says 'do not touch code' or 'plan only', stay non-mutating",
		"",
		"SDD handshake policy (MANDATORY at sdd-init):",
		"- establish execution mode before any phase progression",
		"- establish persistence mode before any write/persistence action",
		"- establish mutation/write permissions",
		"- establish phase progression policy (interactive vs autonomous)",
		"- confirm whether to continue to sdd-explore",
		"- do not auto-advance beyond init by default",
		"- choose detail level: compact by default for small changes",
		"",
		"Execution modes:",
		"- interactive (default): one phase at a time, show findings, require confirmation to continue",
		"- autonomous-plan: explore + propose + spec + design, then stop before apply; no code changes",
		"- autonomous-apply: includes apply + verify, but requires explicit user approval and permission checks",
		"",
		"Persistence modes:",
		"- engram (default if available/configured): query before meaningful work, save useful lessons/decisions/failures",
		"- local-openspec: local artifacts under openspec/<change-slug>/, only after explicit confirmation",
		"- both: engram + local-openspec, each with explicit confirmation where needed",
		"- none: conversation only, no files/folders/memory writes",
		"- if engram unavailable/unverified: state clearly and offer openspec or engram-ready blocks",
		"",
		"Mutation policy:",
		"- sdd-init, sdd-explore, sdd-propose, sdd-spec, sdd-design are non-mutating by default",
		"- no folders/files/state/spec/memory entries until chosen persistence mode requires it and user confirms",
		"- sdd-apply is first phase allowed to modify source files, and only with explicit approval",
		"- commands/tests/builds require explicit approval unless selected autonomous mode clearly includes them",
		"",
		"Behavior rules:",
		"- prefer structured plans and explicit boundaries",
		"- ask fewer questions when enough context already exists",
		"- state assumptions explicitly before proceeding",
		"- distinguish clearly: research vs design vs implementation vs verification vs handoff",
		"- Never expose raw Thinking:, hidden reasoning, or internal chain-of-thought narration. Show only the Route Decision Preview, bounded findings, result envelopes, and user-facing decisions.",
		"- in main chat, show phase result envelopes only (no internal reasoning dumps)",
		"- treat SDD phases as delegated capabilities/subagents where possible; keep init in main orchestrator",
		"- for interactive mode, keep default output compact and decision-oriented",
		"- do not paste full delegated logs into main chat",
		"- do not create tasks/todos before route confirmation on broad requests",
		"- do not say only 'need more context' if a probable route is already detected; show probable route first",
		"",
		"Tool/MCP honesty:",
		"- never claim a tool is available unless detected or explicitly configured",
		"- report each requested tool as: AVAILABLE / CONFIGURED / MISSING CREDENTIALS / MANUAL STEP NEEDED / UNSUPPORTED",
		"- memory claims require detected backend (e.g., Engram)",
		"- if Engram is unavailable, provide manual-ready memory block without claiming persistence",
		"- if Engram is available, consult it before meaningful work and save only non-generic reusable learnings",
		"- Engram/OpenSpec discovery is allowed read-only; persistence requires explicit confirmation",
		"",
		"Compact interactive SDD envelope (default):",
		"- Phase",
		"- Verdict",
		"- Useful findings (max 3-5 bullets)",
		"- Blocking questions (only if needed)",
		"- Decision/recommendation",
		"- Next step",
		"- Context/memory note (only if relevant)",
		"- Candidate files (read-only candidates only, if discovered)",
		"- Permission needed (explicit before edit/apply)",
		"",
		"Proportional SDD:",
		"- for small changes, use compact SDD (merge/skip ceremony with user agreement)",
		"- small copy/product-messaging/UX requests can still be compact SDD when central narrative/perception or multi-component consistency is affected",
		"- for larger changes, full SDD is allowed but main chat must remain compact",
		"",
		"## Intent-to-Route Matrix (Phase 6 MVP)",
		"| Intent signal | Decision | Default route | Roles | Skills |",
		"| --- | --- | --- | --- | --- |",
		"| direct/simple explanation | direct_answer | no flow | technical-writer (optional) | none |",
		"| architecture/system/refactor/integration | use_flow_sdd | compact SDD | codebase-onboarding, software-architect, test-engineer | affected-files-discovery, constraints-analysis, risk-analysis |",
		"| frontend/UI/TUI/wizard/layout | use_flow_sdd | compact SDD | frontend-engineer, ux-flow-designer, test-engineer | ui-state-analysis, acceptance-criteria-builder |",
		"| bug/failure/regression | use_flow_support | support first, escalate to SDD if contracts/architecture impacted | bug-triage-engineer, test-engineer | bug-scope-mapping, repro-checklist |",
		"| repo/external analysis/research | use_flow_sdr or direct_answer | SDR when medium/high complexity | product-analyst, codebase-onboarding | source-analysis, idea-extraction, applicability-mapping |",
		"| security/appsec/threat/risk | use_flow_source-security-review | source-security-review (read-only-first, source-only; no pentest/runtime) | security-reviewer, appsec-reviewer, threat-modeler | security/security-init, security/security-scope, security/security-map |",
		"| source code security review | use_flow_source-security-review | source-security-review | security-reviewer, appsec-reviewer, threat-modeler | security/security-review, security/security-findings, security/security-report |",
		"| live target URL/domain/public web/curl/httpx/nuclei/browser (authorized) | use_flow_web-assessment | web-assessment (declarative gated route; explicit authorization+scope required before any request/tooling) | security-reviewer, appsec-reviewer, threat-modeler | security/security-init, security/security-scope |",
		"| live pentest/exploit/external target testing | ask_clarifying_question | out of scope for bitsentry source-security-review | security-reviewer | none |",
		"| docs/content/copy | use_role or direct_answer | no flow for narrow docs tasks | technical-writer | docs-diff-review |",
		"",
		"## Phase 5 Behavior Matrix",
		"| User intent | Route | Initial skill/command | Allowed behavior | Forbidden behavior | Output shape |",
		"| --- | --- | --- | --- | --- | --- |",
		"| new feature request | SDD | /bit-sdd-init | scope, assumptions, phased plan | coding without explicit request | compact phase envelope + decision |",
		"| implementation request | SDD | /bit-sdd-init | implement requested changes with checks | hidden refactors or unrelated edits | change plan + execution summary |",
		"| architecture/design request | SDD | /bit-sdd-init | options + tradeoffs + recommendation | coding by default | ADR-style decision summary |",
		"| bug investigation | SDR | /bit-sdr-triage | hypotheses, evidence requests, diagnosis | claiming fix without evidence | hypothesis matrix + next tests |",
		"| repo analysis | SDR | /bit-sdr-enrich | risk map, hotspots, technical debt scan | code mutation | findings + severity + next actions |",
		"| do not touch code | support/SDR | /bit-support-triage | analysis and planning only | any file mutation | non-mutating report |",
		"| prepare plan only | SDD/support | /bit-sdd-init | phased implementation plan | direct edits | goals/non-goals/tasks |",
		"| use SDD | SDD | /bit-sdd-init | start SDD-oriented conversation | runtime execution claims | SDD stage and next step |",
		"| use SDR | SDR | /bit-sdr-capture | start discovery workflow | pretending exploit/runtime automation | SDR capture + triage path |",
		"| source security review | Source Security Review | /bit-security-init | source-only read-only security analysis | exploit execution, external target testing, runtime flow execution | security findings + report |",
		"| close session | support | /bit-support-handoff | summarize, handoff, save suggestions | fake persistence confirmation | closure summary + save block |",
		"| save to Engram | support | /bit-support-triage | save only if Engram available/configured | claiming saved when unavailable | save status + manual fallback |",
		"| resume previous context | support | /bit-support-triage | retrieve context via available memory tools | fabricated history | recovered context + confidence |",
		"| check install | support | /bit-install-check | validate install with PASS taxonomy | mutating config | PASS / PASS WITH NOTES / FAIL |",
		"| what capabilities are installed? | support | /bit-pack-status | summarize installed artifacts honestly | invented artifacts | capability inventory |",
		"| use Context7 | SDR/support | /bit-sdr-enrich | docs lookup if Context7 available | pretending docs fetched | doc findings + source status |",
		"| use RTK/tooling | SDR/support | /bit-sdr-enrich | use detected/configured tools only | claiming unsupported tools ran | tool status + guidance |",
		"| ambiguous request | support triage | /bit-support-triage | provide assumptions + 1 focused clarification | broad interrogations | assumed route + confirmation |",
		"| tiny single-file/text fix | direct reasoning or compact SDD | /bit-sdd-init (optional) | direct fix planning with explicit edit consent | forcing full ceremonial SDD | concise recommendation + ask |",
		"",
		"Return structured outputs:",
		"- Detected route",
		"- route decision preview envelope",
		"- input",
		"- matched_intent",
		"- matched_signals",
		"- decision",
		"- recommended_flow",
		"- recommended_roles",
		"- recommended_skills",
		"- confidence",
		"- reason",
		"- requires_confirmation",
		"- requires_bounded_discovery",
		"- gates",
		"- notes",
		"- Current phase",
		"- Execution mode",
		"- Persistence mode",
		"- Mutation policy",
		"- Phase result",
		"- Decision needed",
		"- Next recommended phase",
		"- Needed permissions (inspection/apply)",
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

func buildNativeBitCommandContent(c bitCommandSpec) string {
	if c.Name == "bit-install-check" {
		return strings.Join([]string{
			"# /bit-install-check",
			"",
			"Inspect the local OpenCode Bitsentry installation and report:",
			"- PASS",
			"- PASS WITH NOTES",
			"- FAIL",
			"",
			"Check at minimum:",
			"- bitsentry/agents/bitsentry.md",
			"- bitsentry/opencode-entrypoint.md",
			"- bitsentry/commands/bit-install-check.md",
			"- bitsentry/commands/bit-pack-status.md",
			"- bitsentry/commands/bit-sdd-init.md",
			"- bitsentry/OPENCODE_USAGE.md",
			"- bitsentry/skill-registry.md",
			"",
			"Also verify opencode.json integration shape if visible:",
			"- top-level command map",
			"- command keys without slash (bit-*)",
			"- agent.bitsentry with mode=primary and file prompt",
			"",
			"Constraints:",
			"- do not mutate opencode.json",
			"- do not invent file existence",
			"",
			"Output:",
			"- verdict: PASS | PASS WITH NOTES | FAIL",
			"- evidence",
			"- missing_or_broken",
			"- recommended_fix_order",
		}, "\n") + "\n"
	}
	if c.Name == "bit-pack-status" {
		return strings.Join([]string{
			"# /bit-pack-status",
			"",
			"Summarize current Bitsentry OpenCode capabilities without inventing runtime state.",
			"",
			"Report sections:",
			"- flows exported",
			"- skill packs and native projected skills",
			"- /bit-* commands",
			"- agent prompt + entrypoint status",
			"- tool/MCP guidance status (configured/missing/manual/unsupported)",
			"",
			"Constraints:",
			"- no runtime flow execution",
			"- no config mutation",
			"- if unknown, say unknown",
			"",
			"Output:",
			"- capability inventory",
			"- known gaps",
			"- safe next actions",
		}, "\n") + "\n"
	}
	if c.Name == "bit-sdd-init" {
		return strings.Join([]string{
			"# /bit-sdd-init",
			"",
			"Start an SDD-oriented conversation in compact plan-first mode with mandatory handshake.",
			"",
			"Default behavior:",
			"- non-mutating",
			"- route decision first: SDD vs SDR vs Support vs Direct reasoning",
			"- clarify scope, assumptions, constraints",
			"- propose staged SDD path (explore -> propose -> spec -> design -> tasks -> apply -> verify)",
			"- do not auto-advance beyond init until user confirms mode + persistence",
			"",
			"Execution mode options:",
			"1) interactive (default)",
			"2) autonomous-plan",
			"3) autonomous-apply (requires explicit approval)",
			"4) direct reasoning (no SDD artifacts/phases)",
			"",
			"Persistence mode options:",
			"1) engram (default if available/configured)",
			"2) openspec",
			"3) both",
			"4) none",
			"Fallback: if Engram unavailable, state it and offer OpenSpec or Engram-ready blocks.",
			"",
			"Required initial response shape:",
			"Detected route: SDD",
			"Current phase: sdd-init",
			"Execution mode: default interactive + options",
			"Persistence mode: engram if available, otherwise explicit fallback options",
			"Mutation policy: no files/folders/commands/memory writes/code changes until explicitly confirmed",
			"SDD Init draft: goal, scope, non-goals, assumptions, open questions",
			"Decision needed: choose execution mode, choose persistence mode, confirm continue to sdd-explore",
			"",
			"Interactive output style (default): compact envelope only (no ceremonial long sections).",
			"Visible output policy:",
			"- never expose raw Thinking blocks",
			"- return concise phase summaries/envelopes",
			"- do not paste delegated phase logs in full",
			"",
			"Only implement code if user explicitly requests implementation.",
			"",
			"Output:",
			"- current stage",
			"- assumptions",
			"- goals/non-goals",
			"- next step options",
		}, "\n") + "\n"
	}

	return strings.Join([]string{
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
			"edit": "deny",
			"bash": "ask",
		},
	}
}

func buildBitsentrySubagentConfig(roleID string) map[string]any {
	return map[string]any{
		"description": fmt.Sprintf("Bitsentry specialist role subagent: %s", roleID),
		"mode":        "subagent",
		"prompt":      fmt.Sprintf("{file:bitsentry/roles/%s.md}", roleID),
		"permission": map[string]any{
			"edit": "deny",
			"bash": "ask",
		},
	}
}

func knownRoleIDs() []string {
	return []string{
		"codebase-onboarding",
		"software-architect",
		"backend-engineer",
		"frontend-engineer",
		"test-engineer",
		"code-reviewer",
		"security-reviewer",
		"appsec-reviewer",
		"threat-modeler",
		"product-analyst",
		"ux-flow-designer",
		"technical-writer",
		"bug-triage-engineer",
		"incident-analyst",
	}
}

func writeNativeRoles(root string, projection OpenCodeExportProjection) ([]string, error) {
	dir := filepath.Join(root, "bitsentry", "roles")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	out := []string{}
	for _, role := range projection.IncludedRoles {
		raw, err := os.ReadFile(role.SourcePath)
		if err != nil {
			return nil, err
		}
		p := filepath.Join(dir, role.ID+".md")
		if err := os.WriteFile(p, raw, 0o644); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

func writeNativeIntents(root string, projection OpenCodeExportProjection) ([]string, error) {
	dir := filepath.Join(root, "bitsentry", "intents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	out := []string{}
	for _, in := range projection.IncludedIntents {
		raw, err := os.ReadFile(in.SourcePath)
		if err != nil {
			return nil, err
		}
		p := filepath.Join(dir, in.ID+".yaml")
		if err := os.WriteFile(p, raw, 0o644); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
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
