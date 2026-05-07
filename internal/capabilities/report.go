package capabilities

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bitsentry-ai/internal/config"
	"gopkg.in/yaml.v3"
)

func ApplyReportDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".bitsentry-ai", "exports", "capabilities", "apply"), nil
}

func LatestApplyReportPath() (string, error) {
	dir, err := ApplyReportDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "latest.yaml"), nil
}

func ReadLatestApplyReport() (string, error) {
	path, err := LatestApplyReportPath()
	if err != nil {
		return "", err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read latest report: %w", err)
	}
	return string(raw), nil
}

func WriteApplyReport(cfg config.Config, plan Plan, projection OpenCodeProjection, mode string, status string, note string) (string, error) {
	reportDir, err := ApplyReportDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(reportDir, 0o755); err != nil {
		return "", fmt.Errorf("create report directory: %w", err)
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	reportPath := filepath.Join(reportDir, fmt.Sprintf("%s-%s.yaml", ts, mode))

	payload := map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"mode":         mode,
		"status":       status,
		"note":         note,
		"config": map[string]any{
			"preset":  cfg.Components.Preset,
			"targets": cfg.Components.Targets.Selected,
			"mcps":    cfg.Components.MCPs.Selected,
			"skills":  cfg.Components.Skills.Selected,
			"flows":   cfg.Components.Flows.Selected,
		},
		"plan":                plan,
		"opencode_projection": projection,
	}

	raw, err := yaml.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal report: %w", err)
	}
	if err := os.WriteFile(reportPath, raw, 0o644); err != nil {
		return "", fmt.Errorf("write report: %w", err)
	}

	latestPath := filepath.Join(reportDir, "latest.yaml")
	if err := os.WriteFile(latestPath, raw, 0o644); err != nil {
		return "", fmt.Errorf("write latest report: %w", err)
	}

	return reportPath, nil
}
