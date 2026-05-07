package capabilities

import (
	"fmt"
	"strings"
)

func ApplySummary(plan Plan, projection OpenCodeProjection) string {
	lines := []string{
		"Capabilities apply summary",
		fmt.Sprintf("- target agent: %s", plan.TargetAgent),
		fmt.Sprintf("- preset: %s", plan.Preset),
		fmt.Sprintf("- managed MCPs to apply: %s", fallback(strings.Join(projection.ManagedMCPs, ", "), "none")),
		fmt.Sprintf("- skipped MCPs: %s", fallback(strings.Join(projection.SkippedMCPs, ", "), "none")),
		fmt.Sprintf("- declarative skills: %s", fallback(strings.Join(projection.DeclarativeSkills, ", "), "none")),
		fmt.Sprintf("- declarative flows: %s", fallback(strings.Join(projection.DeclarativeFlows, ", "), "none")),
	}
	return strings.Join(lines, "\n")
}

func fallback(v string, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
