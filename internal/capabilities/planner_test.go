package capabilities

import (
	"testing"

	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
)

func TestBuildPlanFromSelection_DeduplicatesAndSorts(t *testing.T) {
	catalog := Catalog{
		MCPs:   []Entry{{ID: "engram"}, {ID: "context7"}, {ID: "postgres"}},
		Skills: []Entry{{ID: "bitsentry-sdd"}},
		Flows:  []Entry{{ID: "sdd"}, {ID: "notes"}},
	}

	plan, err := BuildPlanFromSelection(catalog, "opencode", "custom", []string{"context7", "engram", "engram"}, []string{"bitsentry-sdd"}, []string{"notes", "sdd", "notes"})
	if err != nil {
		t.Fatalf("BuildPlanFromSelection returned error: %v", err)
	}

	if got, want := len(plan.MCPs), 2; got != want {
		t.Fatalf("unexpected MCP count: got %d want %d", got, want)
	}
	if plan.MCPs[0] != "context7" || plan.MCPs[1] != "engram" {
		t.Fatalf("MCP list not sorted/deduped: %#v", plan.MCPs)
	}
	if plan.Flows[0] != "notes" || plan.Flows[1] != "sdd" {
		t.Fatalf("Flow list not sorted/deduped: %#v", plan.Flows)
	}
}

func TestBuildPlanFromSelection_RejectsUnsupportedTarget(t *testing.T) {
	catalog := Catalog{}
	_, err := BuildPlanFromSelection(catalog, "cursor", "custom", nil, nil, nil)
	if err == nil {
		t.Fatalf("expected error for unsupported target")
	}
}

func TestProjectOpenCode_SplitsManagedAndSkipped(t *testing.T) {
	plan := Plan{MCPs: []string{"engram", "postgres", "context7"}, Skills: []string{"bitsentry-sdd"}, Flows: []string{"sdd"}}
	projection := ProjectOpenCode(plan)

	if len(projection.ManagedMCPs) != 2 {
		t.Fatalf("expected 2 managed MCPs, got %d", len(projection.ManagedMCPs))
	}
	if len(projection.SkippedMCPs) != 1 || projection.SkippedMCPs[0] != "postgres" {
		t.Fatalf("unexpected skipped MCPs: %#v", projection.SkippedMCPs)
	}
}

func TestBuildCatalog_FlowSelectionFromConfig(t *testing.T) {
	cfg := config.Config{}
	cfg.Components.Flows.Enabled = true
	cfg.Components.Flows.Configured = true
	cfg.Components.Flows.Selected = []string{"sdd", "notes"}

	catalog := BuildCatalog(cfg, []components.MCP{}, []components.Skill{})

	selected := map[string]bool{}
	for _, f := range catalog.Flows {
		selected[f.ID] = f.Selected
	}

	if !selected["sdd"] || !selected["notes"] {
		t.Fatalf("expected sdd and notes selected, got %#v", selected)
	}
}

func TestValidateSelections_RejectsUnknownIDs(t *testing.T) {
	catalog := Catalog{
		MCPs:    []Entry{{ID: "engram"}, {ID: "context7"}},
		Skills:  []Entry{{ID: "bitsentry-sdd"}},
		Flows:   []Entry{{ID: "sdd"}},
		Targets: []Entry{{ID: "opencode"}},
	}

	if err := ValidateSelections(catalog, "opencode", []string{"engram"}, []string{"bitsentry-sdd"}, []string{"sdd"}); err != nil {
		t.Fatalf("expected valid selection, got error: %v", err)
	}

	if err := ValidateSelections(catalog, "cursor", nil, nil, nil); err == nil {
		t.Fatalf("expected error for unknown target")
	}
	if err := ValidateSelections(catalog, "opencode", []string{"postgres"}, nil, nil); err == nil {
		t.Fatalf("expected error for unknown mcp")
	}
	if err := ValidateSelections(catalog, "opencode", nil, []string{"unknown-skill"}, nil); err == nil {
		t.Fatalf("expected error for unknown skill")
	}
	if err := ValidateSelections(catalog, "opencode", nil, nil, []string{"unknown-flow"}); err == nil {
		t.Fatalf("expected error for unknown flow")
	}
}

func TestValidateSavedConfig(t *testing.T) {
	catalog := Catalog{
		MCPs:    []Entry{{ID: "engram"}, {ID: "context7"}},
		Skills:  []Entry{{ID: "bitsentry-sdd"}},
		Flows:   []Entry{{ID: "sdd"}, {ID: "notes"}},
		Targets: []Entry{{ID: "opencode"}},
		Presets: []Preset{{ID: "custom"}, {ID: "bitsentry-dev"}},
	}

	valid := config.Config{}
	valid.Components.Preset = "bitsentry-dev"
	valid.Components.Targets.Selected = []string{"opencode"}
	valid.Components.MCPs.Selected = []string{"engram"}
	valid.Components.Skills.Selected = []string{"bitsentry-sdd"}
	valid.Components.Flows.Selected = []string{"sdd"}

	if issues := ValidateSavedConfig(catalog, valid); len(issues) != 0 {
		t.Fatalf("expected no issues for valid config, got %#v", issues)
	}

	invalid := valid
	invalid.Components.Preset = "unknown-preset"
	invalid.Components.Targets.Selected = []string{"cursor"}
	invalid.Components.MCPs.Selected = []string{"postgres"}

	issues := ValidateSavedConfig(catalog, invalid)
	if len(issues) < 3 {
		t.Fatalf("expected at least 3 validation issues, got %#v", issues)
	}
}
