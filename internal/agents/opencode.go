package agents

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type OpenCodeDetector struct{}

func (d OpenCodeDetector) ID() string   { return "opencode" }
func (d OpenCodeDetector) Name() string { return "OpenCode" }

func (d OpenCodeDetector) Detect(ctx context.Context) (AgentDetectionResult, error) {
	res := AgentDetectionResult{ID: d.ID(), Name: d.Name()}

	path, err := exec.LookPath("opencode")
	if err != nil {
		res.Found = false
		res.Hint = "OpenCode no encontrado. Ejecuta install.sh o instala opencode manualmente."
		return res, nil
	}

	res.Found = true
	res.Path = path

	cmd := exec.CommandContext(ctx, path, "--version")
	b, err := cmd.Output()
	if err == nil {
		res.Version = strings.TrimSpace(string(b))
	}

	return res, nil
}

func OpenCodeConfigPathCandidates(projectDir string) []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return []string{}
	}

	candidates := []string{
		filepath.Join(homeDir, ".config", "opencode"),
		filepath.Join(homeDir, ".opencode"),
	}

	if strings.TrimSpace(projectDir) != "" {
		projectLocal := filepath.Join(projectDir, ".opencode")
		if st, statErr := os.Stat(projectLocal); statErr == nil && st.IsDir() {
			candidates = append(candidates, projectLocal)
		}
	}

	set := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, p := range candidates {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if _, exists := set[trimmed]; exists {
			continue
		}
		set[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}

func ExistingOpenCodeConfigPaths(projectDir string) []string {
	paths := OpenCodeConfigPathCandidates(projectDir)
	existing := make([]string, 0, len(paths))
	for _, p := range paths {
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			existing = append(existing, p)
		}
	}
	return existing
}
