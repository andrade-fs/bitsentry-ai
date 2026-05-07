package components

import (
	"testing"

	"bitsentry-ai/internal/config"
)

func TestSkillsRegistryIncludesCoreFamilies(t *testing.T) {
	cfg := config.Config{}
	entries, _ := SkillsRegistry(cfg)

	has := map[string]bool{}
	for _, e := range entries {
		has[e.ID] = true
	}

	for _, id := range []string{"bitsentry-sdd", "bitsentry-sdr", "bitsentry-support"} {
		if !has[id] {
			t.Fatalf("expected skill id %q in registry", id)
		}
	}
}
