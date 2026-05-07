package capabilities

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"bitsentry-ai/internal/app"
	"bitsentry-ai/internal/components"
	"bitsentry-ai/internal/config"
)

type SelectionDraft struct {
	TargetAgent string
	Preset      string
	MCPs        []string
	Skills      []string
	Flows       []string
}

type ValidationResult struct {
	Valid  bool
	Issues []string
}

type PlanResult struct {
	Plan       Plan
	Projection OpenCodeProjection
}

type ApplyResult struct {
	Output           string
	LatestReportPath string
}

type Service struct {
	app      *app.App
	execPath string
}

func NewService(a *app.App, execPath string) *Service {
	return &Service{app: a, execPath: strings.TrimSpace(execPath)}
}

func (s *Service) LoadSelection() (SelectionDraft, error) {
	cfg, err := s.app.ConfigManager.Load()
	if err != nil {
		return SelectionDraft{}, err
	}
	target := "opencode"
	if len(cfg.Components.Targets.Selected) > 0 && strings.TrimSpace(cfg.Components.Targets.Selected[0]) != "" {
		target = cfg.Components.Targets.Selected[0]
	}
	preset := strings.TrimSpace(cfg.Components.Preset)
	if preset == "" {
		preset = "custom"
	}
	return SelectionDraft{
		TargetAgent: target,
		Preset:      preset,
		MCPs:        append([]string{}, cfg.Components.MCPs.Selected...),
		Skills:      append([]string{}, cfg.Components.Skills.Selected...),
		Flows:       append([]string{}, cfg.Components.Flows.Selected...),
	}, nil
}

func (s *Service) SaveSelection(draft SelectionDraft) error {
	cfg, err := s.app.ConfigManager.Load()
	if err != nil {
		return err
	}
	target := strings.TrimSpace(draft.TargetAgent)
	if target == "" {
		target = "opencode"
	}
	preset := strings.TrimSpace(draft.Preset)
	if preset == "" {
		preset = "custom"
	}
	next := cfg
	next.Components.Preset = preset
	next.Components.Targets.Selected = []string{target}
	next.Components.MCPs.Enabled = true
	next.Components.MCPs.Configured = true
	next.Components.MCPs.Selected = uniqueSorted(draft.MCPs)
	next.Components.Skills.Enabled = true
	next.Components.Skills.Configured = true
	next.Components.Skills.Selected = uniqueSorted(draft.Skills)
	next.Components.Flows.Enabled = true
	next.Components.Flows.Configured = true
	next.Components.Flows.Selected = uniqueSorted(draft.Flows)
	return s.app.ConfigManager.Save(next)
}

func (s *Service) ValidateSelection(draft SelectionDraft) (ValidationResult, error) {
	cfg, catalog, err := s.loadContext()
	if err != nil {
		return ValidationResult{}, err
	}
	_ = cfg
	issues := make([]string, 0)
	if err := ValidateSelections(catalog, draft.TargetAgent, draft.MCPs, draft.Skills, draft.Flows); err != nil {
		issues = append(issues, err.Error())
	}
	return ValidationResult{Valid: len(issues) == 0, Issues: issues}, nil
}

func (s *Service) BuildPlan(draft SelectionDraft) (PlanResult, error) {
	_, catalog, err := s.loadContext()
	if err != nil {
		return PlanResult{}, err
	}
	plan, err := BuildPlanFromSelection(catalog, draft.TargetAgent, draft.Preset, draft.MCPs, draft.Skills, draft.Flows)
	if err != nil {
		return PlanResult{}, err
	}
	return PlanResult{Plan: plan, Projection: ProjectOpenCode(plan)}, nil
}

func (s *Service) ApplyDryRun(draft SelectionDraft) (ApplyResult, error) {
	if strings.TrimSpace(s.execPath) == "" {
		return ApplyResult{}, fmt.Errorf("capability service has no executable path configured")
	}
	if err := s.SaveSelection(draft); err != nil {
		return ApplyResult{}, err
	}
	output, err := s.runExec("--dry-run", "capabilities", "apply")
	if err != nil {
		return ApplyResult{}, err
	}
	latest, _ := s.ReadLatestReportPath()
	return ApplyResult{Output: output, LatestReportPath: latest}, nil
}

func (s *Service) Apply(draft SelectionDraft) (ApplyResult, error) {
	if strings.TrimSpace(s.execPath) == "" {
		return ApplyResult{}, fmt.Errorf("capability service has no executable path configured")
	}
	if err := s.SaveSelection(draft); err != nil {
		return ApplyResult{}, err
	}
	output, err := s.runExec("capabilities", "apply")
	if err != nil {
		return ApplyResult{}, err
	}
	latest, _ := s.ReadLatestReportPath()
	return ApplyResult{Output: output, LatestReportPath: latest}, nil
}

func (s *Service) ReadLatestReportPath() (string, error) {
	path, err := LatestApplyReportPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Service) runExec(args ...string) (string, error) {
	cmd := exec.Command(s.execPath, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s %s: %w | %s", s.execPath, strings.Join(args, " "), err, strings.TrimSpace(out.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func (s *Service) loadContext() (config.Config, Catalog, error) {
	cfg, err := s.app.ConfigManager.Load()
	if err != nil {
		return config.Config{}, Catalog{}, err
	}
	engramDetails := components.DetectEngramRuntime(context.Background(), cfg)
	context7Details := components.DetectContext7Runtime(context.Background(), cfg)
	mcpEntries, _ := components.MCPRegistry(context.Background(), cfg, engramDetails, context7Details)
	skillEntries, _ := components.SkillsRegistry(cfg)
	return cfg, BuildCatalog(cfg, mcpEntries, skillEntries), nil
}

func uniqueSorted(values []string) []string {
	set := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, v := range values {
		t := strings.TrimSpace(v)
		if t == "" || set[t] {
			continue
		}
		set[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
