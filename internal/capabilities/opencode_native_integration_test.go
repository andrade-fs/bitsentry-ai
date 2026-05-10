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
	if perm["edit"] != "ask" || perm["bash"] != "ask" {
		t.Fatalf("unexpected permission block: %#v", perm)
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

func TestExecuteOpenCodeNativeIntegration_PreservesExistingConfigAndNoDupInstruction(t *testing.T) {
	root := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
		existing := map[string]any{
		"$schema": "x",
		"instructions": []string{"keep.md", "bitsentry/opencode-entrypoint.md"},
		"agent": map[string]any{
			"plan":      map[string]any{"name": "plan"},
			"bitsentry": map[string]any{"name": "bitsentry", "prompt": "bitsentry/agents/bitsentry.md"},
		},
		"mcp": map[string]any{"x": map[string]any{"enabled": true}},
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
	if !ok || bPerm["edit"] != "ask" || bPerm["bash"] != "ask" {
		t.Fatalf("expected bitsentry permission block repaired: %#v", b["permission"])
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
