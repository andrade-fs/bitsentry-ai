package capabilities

import (
	"fmt"
	"sort"
	"strings"

	"bitsentry-ai/internal/config"
)

func ValidateSelections(catalog Catalog, target string, mcps []string, skills []string, flows []string) error {
	target = strings.TrimSpace(target)
	if target != "" {
		knownTargets := makeSet(catalog.Targets)
		if !knownTargets[target] {
			return fmt.Errorf("unknown target %q. Available targets: %s", target, strings.Join(sortedKeys(knownTargets), ", "))
		}
	}

	if err := validateIDs("MCP", mcps, makeSet(catalog.MCPs)); err != nil {
		return err
	}
	if err := validateIDs("Skill", skills, makeSet(catalog.Skills)); err != nil {
		return err
	}
	if err := validateIDs("Flow", flows, makeSet(catalog.Flows)); err != nil {
		return err
	}
	return nil
}

func ValidateSavedConfig(catalog Catalog, cfg config.Config) []string {
	issues := make([]string, 0)

	if strings.TrimSpace(cfg.Components.Preset) != "" {
		if _, ok := PresetByID(strings.TrimSpace(cfg.Components.Preset), catalog.Presets); !ok {
			issues = append(issues, fmt.Sprintf("components.preset %q is not a modeled preset", cfg.Components.Preset))
		}
	}

	if len(cfg.Components.Targets.Selected) == 0 {
		issues = append(issues, "components.targets.selected is empty")
	} else {
		for _, target := range cfg.Components.Targets.Selected {
			if err := ValidateSelections(catalog, target, nil, nil, nil); err != nil {
				issues = append(issues, err.Error())
			}
		}
	}

	if err := ValidateSelections(catalog, "", cfg.Components.MCPs.Selected, cfg.Components.Skills.Selected, cfg.Components.Flows.Selected); err != nil {
		issues = append(issues, err.Error())
	}

	return uniqueSortedStrings(issues)
}

func validateIDs(kind string, selected []string, known map[string]bool) error {
	for _, id := range selected {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if !known[trimmed] {
			return fmt.Errorf("unknown %s %q. Available %ss: %s", kind, trimmed, strings.ToLower(kind), strings.Join(sortedKeys(known), ", "))
		}
	}
	return nil
}

func sortedKeys(in map[string]bool) []string {
	out := make([]string, 0, len(in))
	for k := range in {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func uniqueSortedStrings(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" || set[trimmed] {
			continue
		}
		set[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}
