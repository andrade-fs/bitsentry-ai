package capabilities

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteOpenCodeNativeIntegration_CreatesConfigAndFiles(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	cat, err := DiscoverAssets("../..")
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	p, err := BuildOpenCodeExportProjection(cat, []string{"sdd", "support"})
	if err != nil {
		t.Fatalf("projection: %v", err)
	}
	res, err := ExecuteOpenCodeNativeIntegration(p, root, OpenCodeNativeOptions{RegisterAgent: true, InstallCommands: true, InstallNativeSkills: true})
	if err != nil {
		t.Fatalf("native integration: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bitsentry", "agents", "bitsentry.md")); err != nil {
		t.Fatalf("missing bitsentry agent prompt: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "bitsentry", "opencode-entrypoint.md")); err != nil {
		t.Fatalf("missing entrypoint: %v", err)
	}
	if len(res.CommandFiles) == 0 {
		t.Fatalf("expected command files")
	}
	if len(res.NativeSkillFiles) == 0 {
		t.Fatalf("expected native skill files")
	}
	if len(res.RoleFiles) < 14 {
		t.Fatalf("expected role files projected")
	}
	if len(res.IntentFiles) < 7 {
		t.Fatalf("expected intent files projected")
	}
	raw, err := os.ReadFile(filepath.Join(root, "opencode.json"))
	if err != nil {
		t.Fatalf("read opencode.json: %v", err)
	}
	obj := map[string]any{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("parse opencode.json: %v", err)
	}
	ins := toStringSlice(obj["instructions"])
	if !containsString(ins, "bitsentry/opencode-entrypoint.md") {
		t.Fatalf("missing entrypoint instruction")
	}
	a, ok := obj["agent"].(map[string]any)
	if !ok || a["bitsentry"] == nil {
		t.Fatalf("missing agent.bitsentry")
	}
	bits := a["bitsentry"].(map[string]any)
	if bits["mode"] != "primary" {
		t.Fatalf("expected bitsentry mode primary")
	}
	if bits["prompt"] != "{file:bitsentry/agents/bitsentry.md}" {
		t.Fatalf("unexpected bitsentry prompt: %v", bits["prompt"])
	}
	if _, hasName := bits["name"]; hasName {
		t.Fatalf("bitsentry agent must not include legacy name field")
	}
	perm, ok := bits["permission"].(map[string]any)
	if !ok {
		t.Fatalf("expected permission block")
	}
	if perm["edit"] != "deny" || perm["bash"] != "ask" {
		t.Fatalf("unexpected permission block: %#v", perm)
	}
	for _, role := range knownRoleIDs() {
		entry, ok := a[role].(map[string]any)
		if !ok {
			t.Fatalf("missing role subagent %s", role)
		}
		if entry["mode"] != "subagent" {
			t.Fatalf("role %s must be subagent", role)
		}
		if entry["prompt"] != "{file:bitsentry/roles/"+role+".md}" {
			t.Fatalf("unexpected role prompt for %s: %v", role, entry["prompt"])
		}
	}
	if _, exists := obj["commands"]; exists {
		t.Fatalf("unexpected top-level commands key")
	}
	cmd, ok := obj["command"].(map[string]any)
	if !ok {
		t.Fatalf("missing top-level command map")
	}
	for k, v := range cmd {
		if strings.HasPrefix(k, "/") {
			t.Fatalf("command key must not start with slash: %s", k)
		}
		if !strings.HasPrefix(k, "bit-") {
			continue
		}
		e, ok := v.(map[string]any)
		if !ok {
			t.Fatalf("command entry %s must be object", k)
		}
		if _, ok := e["prompt"]; ok {
			t.Fatalf("command entry %s must not use prompt", k)
		}
		template, _ := e["template"].(string)
		if !strings.HasPrefix(template, "{file:bitsentry/commands/") {
			t.Fatalf("command entry %s bad template: %s", k, template)
		}
	}
}

func TestBuildOpenCodeMCPConfigPreview_MissingOpenCodeJSON(t *testing.T) {
	root := t.TempDir()
	p := BuildOpenCodeMCPConfigPreview(root, []string{"engram", "context7"})
	if p.Exists {
		t.Fatalf("expected exists=false")
	}
	if p.CurrentConfigState != "missing_opencode_json" {
		t.Fatalf("unexpected state: %s", p.CurrentConfigState)
	}
	if p.WouldWrite {
		t.Fatalf("preview must be read-only")
	}
	if !p.RequiresConfirmation || !p.BackupRequired {
		t.Fatalf("future apply contract must require confirmation+backup")
	}
}

func TestBuildOpenCodeMCPConfigPreview_InvalidOpenCodeJSON(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), []byte("{"), 0o644)
	p := BuildOpenCodeMCPConfigPreview(root, []string{"engram"})
	if !p.Exists {
		t.Fatalf("expected exists=true")
	}
	if p.Readable {
		t.Fatalf("expected readable=false for invalid json")
	}
	if p.CurrentConfigState != "invalid" {
		t.Fatalf("unexpected state: %s", p.CurrentConfigState)
	}
	if strings.TrimSpace(p.InvalidError) == "" {
		t.Fatalf("expected invalid error message")
	}
}

func TestBuildOpenCodeMCPConfigPreview_PreservesExistingMCPConfig(t *testing.T) {
	root := t.TempDir()
	raw := []byte(`{"agent":{"custom":{}},"mcp":{"existing":{"enabled":true},"github":{"enabled":false}},"other":1}`)
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), raw, 0o644)
	p := BuildOpenCodeMCPConfigPreview(root, []string{"context7"})
	if !p.CurrentMCPConfigDetected {
		t.Fatalf("expected existing mcp config detected")
	}
	if p.CurrentConfigState != "readable_with_mcp" {
		t.Fatalf("unexpected state: %s", p.CurrentConfigState)
	}
	if !containsString(p.PreservedKeys, "agent") || !containsString(p.PreservedKeys, "mcp") {
		t.Fatalf("expected preserved top-level keys, got %#v", p.PreservedKeys)
	}
	if !containsString(p.PreservedMCPEntries, "existing") || !containsString(p.PreservedMCPEntries, "github") {
		t.Fatalf("expected preserved mcp entries, got %#v", p.PreservedMCPEntries)
	}
}

func TestBuildOpenCodeMCPConfigPreview_ProposedChangesExcludeCredentialsSecrets(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), []byte(`{"mcp":{}}`), 0o644)
	p := BuildOpenCodeMCPConfigPreview(root, []string{"engram", "context7"})
	joined := strings.ToLower(strings.Join(p.ProposedSafeChanges, " "))
	forbidden := []string{"token", "api_key", "apikey", "password", "secret", "credential"}
	for _, f := range forbidden {
		if strings.Contains(joined, f) {
			t.Fatalf("proposed changes must not include sensitive field %q: %s", f, joined)
		}
	}
}

func TestBuildOpenCodeMCPConfigPreview_NoWriteInPreview(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), []byte(`{"mcp":{}}`), 0o644)
	_ = BuildOpenCodeMCPConfigPreview(root, []string{"engram"})
	backupDir := filepath.Join(home, ".bitsentry-ai", "backups")
	if _, err := os.Stat(backupDir); err == nil {
		t.Fatalf("preview must not create backup directories")
	}
}

func TestExecuteOpenCodeNativeIntegration_PreservesExistingConfigAndNoDupInstruction(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	existing := map[string]any{
		"$schema":      "x",
		"instructions": []string{"keep.md", "bitsentry/opencode-entrypoint.md"},
		"agent": map[string]any{
			"plan":      map[string]any{"name": "plan"},
			"bitsentry": map[string]any{"name": "bitsentry", "prompt": "bitsentry/agents/bitsentry.md"},
		},
		"mcp":         map[string]any{"x": map[string]any{"enabled": true}},
		"permissions": map[string]any{"allow": []string{"*"}},
	}
	raw, _ := json.Marshal(existing)
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), raw, 0o644)

	cat, _ := DiscoverAssets("../..")
	p, _ := BuildOpenCodeExportProjection(cat, []string{"sdd"})
	if _, err := ExecuteOpenCodeNativeIntegration(p, root, OpenCodeNativeOptions{RegisterAgent: true}); err != nil {
		t.Fatalf("native integration: %v", err)
	}
	raw2, _ := os.ReadFile(filepath.Join(root, "opencode.json"))
	obj := map[string]any{}
	_ = json.Unmarshal(raw2, &obj)
	ins := toStringSlice(obj["instructions"])
	count := 0
	for _, v := range ins {
		if v == "bitsentry/opencode-entrypoint.md" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected one entrypoint instruction, got %d", count)
	}
	if _, ok := obj["mcp"]; !ok {
		t.Fatalf("expected preserve mcp")
	}
	if _, ok := obj["permissions"]; !ok {
		t.Fatalf("expected preserve permissions")
	}
	a, _ := obj["agent"].(map[string]any)
	if a["plan"] == nil {
		t.Fatalf("expected preserve existing agent.plan")
	}
	b := a["bitsentry"].(map[string]any)
	if _, hasName := b["name"]; hasName {
		t.Fatalf("expected legacy name removed from bitsentry agent")
	}
	if b["mode"] != "primary" {
		t.Fatalf("expected bitsentry mode primary")
	}
	if b["prompt"] != "{file:bitsentry/agents/bitsentry.md}" {
		t.Fatalf("expected bitsentry file prompt")
	}
	bPerm, ok := b["permission"].(map[string]any)
	if !ok || bPerm["edit"] != "deny" || bPerm["bash"] != "ask" {
		t.Fatalf("expected bitsentry permission block repaired: %#v", b["permission"])
	}
	if role, ok := a["software-architect"].(map[string]any); !ok || role["mode"] != "subagent" {
		t.Fatalf("expected role subagent preserved/installed")
	}
}

func TestExecuteOpenCodeNativeIntegration_RejectsUnparsableConfig(t *testing.T) {
	root := t.TempDir()
	_ = os.WriteFile(filepath.Join(root, "opencode.json"), []byte("{"), 0o644)
	cat, _ := DiscoverAssets("../..")
	p, _ := BuildOpenCodeExportProjection(cat, []string{"sdd"})
	_, err := ExecuteOpenCodeNativeIntegration(p, root, OpenCodeNativeOptions{RegisterAgent: true})
	if err == nil || !strings.Contains(err.Error(), "unparsable") {
		t.Fatalf("expected unparsable config error")
	}
}

func TestBitCommandsNamespaceAndNoOrchestrator(t *testing.T) {
	for _, c := range bitCommandSpecs() {
		if !strings.HasPrefix(c.Name, "bit-") {
			t.Fatalf("command without bit- prefix: %s", c.Name)
		}
		if strings.Contains(c.Name, "orchestrator") {
			t.Fatalf("orchestrator command forbidden: %s", c.Name)
		}
	}
}

func TestMigrateCommandsToCommandAndPromptToTemplate(t *testing.T) {
	obj := map[string]any{
		"commands": map[string]any{
			"/bit-sdd-init": map[string]any{"prompt": "bitsentry/commands/bit-sdd-init.md", "agent": "bitsentry"},
		},
		"command": map[string]any{
			"/bit-pack-status": map[string]any{"prompt": "{file:bitsentry/commands/bit-pack-status.md}", "agent": "bitsentry"},
			"custom-user":      map[string]any{"description": "keep"},
		},
	}
	m, err := loadAndMigrateCommandMap(obj)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, ok := m["bit-sdd-init"]; !ok {
		t.Fatalf("expected migrated bit-sdd-init key")
	}
	e := m["bit-sdd-init"].(map[string]any)
	if _, ok := e["prompt"]; ok {
		t.Fatalf("prompt should be migrated to template")
	}
	if e["template"] != "{file:bitsentry/commands/bit-sdd-init.md}" {
		t.Fatalf("unexpected template value: %v", e["template"])
	}
	if _, ok := m["custom-user"]; !ok {
		t.Fatalf("expected existing user command preserved")
	}
}

func TestMigrateCommandsFailsWhenUnknownEntryExists(t *testing.T) {
	obj := map[string]any{
		"commands": map[string]any{"custom": map[string]any{"x": true}},
	}
	if _, err := loadAndMigrateCommandMap(obj); err == nil {
		t.Fatalf("expected failure for unknown non-bitsentry commands entry")
	}
}

func TestBitsentryPromptContainsPhase5OrchestrationContract(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	required := []string{
		"OpenCode-native Bitsentry orchestrator",
		"Intent routing:",
		"SDD:",
		"SDR:",
		"Support:",
		"bitsentry is a non-mutating orchestrator by default",
		"bitsentry must not edit files directly",
		"implementation/apply requires an explicit approved apply path",
		"no internal Bitsentry runtime/session execution",
		"do not modify code unless explicitly requested",
		"Tool/MCP honesty:",
		"## Phase 5 Behavior Matrix",
		"ambiguous request",
		"save to Engram",
		"use Context7",
		"SDD handshake policy (MANDATORY at sdd-init)",
		"Execution modes:",
		"interactive (default)",
		"Persistence modes:",
		"none: conversation only, no files/folders/memory writes",
		"do not auto-advance beyond init by default",
		"Never expose raw Thinking:",
		"treat SDD phases as delegated capabilities/subagents",
		"Route decision first (when no flow active):",
		"OpenCode-native route decision preview is the PRIMARY route-selection UX",
		"CLI `bitsentry-ai route decide` is debug/plumbing parity, not the primary end-user workflow",
		"do NOT force SDD automatically",
		"Direct reasoning: trivial/small tasks where formal flow is overkill",
		"Compact interactive SDD envelope (default):",
		"do not paste full delegated logs into main chat",
		"if Engram is available, consult it before meaningful work",
		"save only non-generic reusable learnings",
		"if engram unavailable/unverified: state clearly and offer openspec or engram-ready blocks",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("agent prompt missing %q", token)
		}
	}
}

func TestBitsentryPromptClassifiesRoutesAndDirectOption(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	for _, token := range []string{
		"Intent Decision Contract (choose exactly one first action)",
		"direct_answer | use_skill | use_role | use_flow_sdr | use_flow_support | use_flow_sdd | use_flow_source-security-review | use_flow_web-assessment | bounded_discovery_then_decide | ask_clarifying_question",
		"Questions simples/explanations => direct_answer (NO SDD by default).",
		"do NOT activate any formal flow silently; announce selected flow first",
		"ask_clarifying_question is allowed only when route cannot be decided with high confidence",
		"include matched_signals in route preview reasoning when available",
		"## Intent-to-Route Matrix (Phase 6 MVP)",
		"architecture/system/refactor/integration",
		"frontend/UI/TUI/wizard/layout",
		"bug/failure/regression",
		"security/appsec/threat/risk",
		"source code security review",
		"live target URL/domain/public web/curl/httpx/nuclei/browser (authorized)",
		"SDD:",
		"SDR:",
		"Support:",
		"Direct reasoning:",
		"if ambiguous, ask user to choose before entering a flow",
		"bounded context discovery is allowed only AFTER a visible Route Decision Preview",
		"bounded discovery limits: read-only only",
		"for narrow direct requests",
		"open file X and show Y",
		"SDR brief audit -> compact SDD corrections",
		"web-assessment planning is allowed",
		"planning without live interaction",
		"authorized synthetic planning",
		"no over-refusal for planning",
		"allowed without live interaction",
		"gated live interaction",
		"prohibited by default",
		"Route decision MUST be visible in chat BEFORE any non-trivial discovery",
		"updated route decision only if findings change",
		"separate permission before apply/edit planning and before persistence",
		"do not create tasks/todos before route confirmation on broad requests",
		"require explicit authorization + exact scope + environment + testing window + intensity + tool allowlist BEFORE any request/tooling",
	} {
		if !strings.Contains(prompt, token) {
			t.Fatalf("route classification missing %q", token)
		}
	}
}

func TestBitsentryPromptWebAssessmentPlanningAllowedWithoutLiveInteraction(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"web-assessment planning is allowed",
		"planning without live interaction",
		"authorized synthetic planning",
		"assessment session context",
		"planning_only",
		"dry_run",
		"execute_approved",
		"retest",
		"no over-refusal for planning",
		"allowed without live interaction",
		"analysis of provided synthetic evidence",
		"scope/out-of-scope",
		"rate limits",
		"stop conditions",
		"evidence logging plan",
		"defensive checklist",
		"report template",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected web-assessment planning token %q", token)
		}
	}
}

func TestBitsentryPromptWebAssessmentExecutionRemainsGated(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"no execution by default",
		"explicit approval before requests",
		"no live requests without complete gates",
		"no tooling without explicit approval",
		"gated live interaction",
		"before any request/tooling",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected web-assessment gate token %q", token)
		}
	}
}

func TestBitsentryPromptWebAssessmentHardGuardrailsPreserved(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"prohibited by default",
		"exploit execution",
		"dos/load testing",
		"credential attacks",
		"brute force",
		"password spraying",
		"aggressive fuzzing",
		"mass scanning",
		"out-of-scope scanning",
		"destructive actions",
		"exfiltration",
		"secrets exposure",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected hard guardrail token %q", token)
		}
	}
}

func TestBitsentryPromptBroadRequestGatesToolUseBeforeRouteConfirmation(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	checks := []string{
		"bounded context discovery is allowed only AFTER a visible Route Decision Preview",
		"read-only only, no edits, no todos/tasks, no persistence writes, no implementation decisions",
		"Route decision MUST be visible in chat BEFORE any non-trivial discovery",
		"Engram/OpenSpec discovery is allowed read-only; persistence requires explicit confirmation",
	}
	for _, c := range checks {
		if !strings.Contains(prompt, c) {
			t.Fatalf("missing broad-request gating rule %q", c)
		}
	}
}

func TestBitsentryPromptRequiresVisibleRouteAfterDiscovery(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	for _, token := range []string{
		"Route decision MUST be visible in chat BEFORE any non-trivial discovery",
		"updated route decision only if findings change",
		"post-discovery gate is mandatory: STOP after discovery and ask confirmation",
		"before route selection, edit planning, apply, or persistence",
		"forbidden without approval: statements like 'procedo directamente con las ediciones'",
		"Never expose raw Thinking:",
		"do not say only 'need more context' if a probable route is already detected; show probable route first",
	} {
		if !strings.Contains(prompt, token) {
			t.Fatalf("prompt missing visible route enforcement %q", token)
		}
	}
}

func TestBitsentryPromptPostDiscoveryOutputRequiresPermissionAndNoApply(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	required := []string{
		"Route decision",
		"Decision/recommendation",
		"Candidate files (read-only candidates only, if discovered)",
		"Permission needed (explicit before edit/apply)",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("prompt missing post-discovery output token %q", token)
		}
	}
}

func TestBitsentryPromptContainsRequiredVisibleOutputShape(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	for _, token := range []string{
		"Detected route",
		"route decision preview envelope",
		"matched_intent",
		"matched_signals",
		"primary_skills",
		"secondary_skills",
		"deferred_skills",
		"primary_roles",
		"secondary_roles",
		"capability_reason",
		"capability_gates",
		"recommended_flow",
		"recommended_roles",
		"recommended_skills",
		"requires_bounded_discovery",
		"gates",
		"Current phase",
		"Execution mode",
		"Persistence mode",
		"Mutation policy",
		"Phase result",
		"Decision needed",
		"Next recommended phase",
	} {
		if !strings.Contains(prompt, token) {
			t.Fatalf("prompt missing visible output field %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_OpenCodeFirstRoutePreviewContract(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"opencode-native route decision preview is the primary route-selection ux",
		"cli `bitsentry-ai route decide` is debug/plumbing parity, not the primary end-user workflow",
		"route decision preview envelope",
		"matched_signals",
		"requires_confirmation",
		"requires_bounded_discovery",
		"no_flow_execution_in_preview",
		"no persistence",
		"no edits",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected prompt to contain %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_CompactDirectModeContractPresent(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"compact direct mode & agent token efficiency",
		"for direct/atomic/concrete/already-intentional requests with clear implicit intent",
		"without full route decision preview table",
		"without full capability preview table",
		"compact direct mode does not remove internal reasoning quality",
		"for direct requests that are sensitive/remote/destructive/ambiguous",
		"require a brief explicit confirmation",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected compact direct mode token %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_FullPreviewKeptForOpenEndedAndSensitiveScopes(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"for open/ambiguous/exploratory/strategic/planning/architecture/security/multi-file/sensitive requests",
		"first visible response must be a full route decision preview",
		"full preview trigger classes: open, ambiguous, exploratory, planning, architecture, security/risk, multi-file change, sensitive/high-impact operations",
		"the route decision preview template is mandatory when full preview applies",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected full-preview trigger token %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_DoesNotPresentCLIAsPrimaryUX(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	forbidden := []string{
		"users should run bitsentry-ai route decide",
		"ask the user to run bitsentry-ai route decide",
		"primary workflow is bitsentry-ai route decide",
		"use the cli as the main workflow",
		"run bitsentry-ai route decide manually",
	}
	for _, phrase := range forbidden {
		if strings.Contains(prompt, phrase) {
			t.Fatalf("prompt must not present CLI as primary UX: found %q", phrase)
		}
	}
}

func TestBitsentryAgentPrompt_RequiresVisibleRoutePreviewBeforeDiscovery(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	required := []string{
		"before any non-trivial repository discovery",
		"must show a concise route decision preview",
		"visible to the user",
		"only after a visible route decision preview",
		"requires_bounded_discovery=true",
		"before any non-trivial discovery",
		"updated route decision only if findings change",
		"never expose raw thinking:",
		"matched_intent",
		"decision",
		"matched_signals",
		"requires_confirmation",
		"requires_bounded_discovery",
		"gates",
		"no_edits_in_preview",
		"no_persistence_in_preview",
		"no_flow_execution_in_preview",
		"bounded read-only discovery",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected prompt to contain %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_DoesNotAllowDiscoveryBeforeRouteConfirmation(t *testing.T) {
	prompt := strings.ToLower(buildBitsentryAgentPrompt())
	forbidden := []string{
		"discovery is allowed before route confirmation",
		"bounded context discovery is allowed before route confirmation",
	}
	for _, token := range forbidden {
		if strings.Contains(prompt, token) {
			t.Fatalf("prompt must not contain contradictory discovery-before-preview rule %q", token)
		}
	}
}

func TestBitsentryAgentPrompt_FrontloadsFirstResponseRulesBeforeMatrices(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	first := strings.Index(prompt, "## Non-negotiable first response rules")
	intentRouting := strings.Index(prompt, "Intent routing:")
	routeDecisionFirst := strings.Index(prompt, "Route decision first (when no flow active):")
	intentMatrix := strings.Index(prompt, "## Intent-to-Route Matrix (Phase 6 MVP)")
	phase5Matrix := strings.Index(prompt, "## Phase 5 Behavior Matrix")
	if first < 0 || intentRouting < 0 || routeDecisionFirst < 0 || intentMatrix < 0 || phase5Matrix < 0 {
		t.Fatalf("expected key prompt sections present")
	}
	if !(first < intentRouting && first < routeDecisionFirst && first < intentMatrix && first < phase5Matrix) {
		t.Fatalf("first response rules must appear before routing and matrices")
	}
}

func TestBitsentryAgentPrompt_ContainsFirstResponseRoutePreviewTemplate(t *testing.T) {
	prompt := buildBitsentryAgentPrompt()
	required := []string{
		"Route Decision Preview",
		"- matched_intent:",
		"- decision:",
		"- matched_signals:",
		"- primary_skills:",
		"- secondary_skills:",
		"- deferred_skills:",
		"- primary_roles:",
		"- secondary_roles:",
		"- capability_reason:",
		"- capability_gates:",
		"- reason:",
		"- requires_bounded_discovery:",
		"- requires_confirmation:",
		"- gates:",
		"- no_edits_in_preview",
		"- no_persistence_in_preview",
		"- no_flow_execution_in_preview",
		"Only AFTER showing Route + Capability Decision Preview may you inspect files",
		"small copy/product-messaging/UX changes may still require compact SDD",
	}
	for _, token := range required {
		if !strings.Contains(prompt, token) {
			t.Fatalf("expected first response template token %q", token)
		}
	}
}

func TestBitsentryEntrypointContainsBoundariesAndGuidance(t *testing.T) {
	entry := buildBitsentryEntrypoint()
	required := []string{
		"Activate agent: `@bitsentry`",
		"/bit-*",
		"installer/projector/validator/pack manager, NOT runtime flow executor",
		"Do not change repository code unless the user explicitly requests implementation",
		"Do not claim memory persistence unless Engram",
		"AVAILABLE / CONFIGURED / MISSING CREDENTIALS / MANUAL STEP NEEDED / UNSUPPORTED",
	}
	for _, token := range required {
		if !strings.Contains(entry, token) {
			t.Fatalf("entrypoint missing %q", token)
		}
	}
}

func TestNativeCommandsContainPhase5Contracts(t *testing.T) {
	sdd := buildNativeBitCommandContent(bitCommandSpec{Name: "bit-sdd-init", Action: "bit-sdd-init", Description: "x"})
	if !strings.Contains(sdd, "Start an SDD-oriented conversation in compact plan-first mode") {
		t.Fatalf("bit-sdd-init must be plan-first")
	}
	if !strings.Contains(sdd, "Only implement code if user explicitly requests implementation") {
		t.Fatalf("bit-sdd-init must stay non-mutating by default")
	}
	for _, token := range []string{
		"route decision first: SDD vs SDR vs Support vs Direct reasoning",
		"Execution mode options:",
		"interactive (default)",
		"autonomous-plan",
		"autonomous-apply (requires explicit approval)",
		"direct reasoning (no SDD artifacts/phases)",
		"Persistence mode options:",
		"engram (default if available/configured)",
		"openspec",
		"both",
		"none",
		"do not auto-advance beyond init until user confirms mode + persistence",
		"Fallback: if Engram unavailable, state it and offer OpenSpec or Engram-ready blocks",
		"Interactive output style (default): compact envelope only",
		"never expose raw Thinking blocks",
		"do not paste delegated phase logs in full",
	} {
		if !strings.Contains(sdd, token) {
			t.Fatalf("bit-sdd-init missing required token %q", token)
		}
	}

	install := buildNativeBitCommandContent(bitCommandSpec{Name: "bit-install-check", Action: "bit-install-check", Description: "x"})
	for _, token := range []string{"PASS", "PASS WITH NOTES", "FAIL"} {
		if !strings.Contains(install, token) {
			t.Fatalf("bit-install-check missing verdict token %q", token)
		}
	}

	status := buildNativeBitCommandContent(bitCommandSpec{Name: "bit-pack-status", Action: "bit-pack-status", Description: "x"})
	if !strings.Contains(status, "without inventing runtime state") {
		t.Fatalf("bit-pack-status must avoid invented state")
	}
}
