package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RouteReport struct {
	Session                RouteReportSession   `json:"session"`
	Brief                  RouteReportBrief     `json:"brief"`
	Handoff                RouteReportHandoff   `json:"handoff"`
	Plan                   RouteReportPlan      `json:"plan"`
	Progress               RouteReportProgress  `json:"progress"`
	Lifecycle              RouteReportLifecycle `json:"lifecycle"`
	Safety                 RouteReportSafety    `json:"safety"`
	WouldExecute           bool                 `json:"would_execute"`
	Autonomous             bool                 `json:"autonomous"`
	ExternalAgentExecution bool                 `json:"external_agent_execution"`
	SkillExecution         bool                 `json:"skill_execution"`
	OpenCodeMutation       bool                 `json:"opencode_mutation"`
}

type RouteReportSession struct {
	ID            string `json:"id"`
	Intent        string `json:"intent"`
	Flow          string `json:"flow"`
	Status        string `json:"status"`
	InitialSkill  string `json:"initial_skill"`
	SchemaVersion string `json:"schema_version,omitempty"`
	Path          string `json:"path"`
}

type RouteReportBrief struct {
	Path     string `json:"path"`
	Markdown string `json:"markdown"`
}

type RouteReportHandoff struct {
	Path     string `json:"path"`
	Markdown string `json:"markdown"`
}

type RouteReportPlan struct {
	Flow   string           `json:"flow"`
	Stages []ExecutionStage `json:"stages"`
}

type RouteReportProgress struct {
	CurrentStageID    string   `json:"current_stage_id,omitempty"`
	CompletedStageIDs []string `json:"completed_stage_ids"`
	UpdatedAt         string   `json:"updated_at"`
}

type RouteReportLifecycle struct {
	SchemaVersion string `json:"schema_version,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Archived      bool   `json:"archived"`
	ArchivedAt    string `json:"archived_at,omitempty"`
	RestoredAt    string `json:"restored_at,omitempty"`
}

type RouteReportSafety struct {
	WouldExecute           bool `json:"would_execute"`
	Autonomous             bool `json:"autonomous"`
	ExternalAgentExecution bool `json:"external_agent_execution"`
	SkillExecution         bool `json:"skill_execution"`
	OpenCodeMutation       bool `json:"opencode_mutation"`
}

func BuildRouteReport(session Session, sessionPath string) (RouteReport, error) {
	briefPath := filepath.Join(sessionPath, "brief.md")
	briefRaw, err := os.ReadFile(briefPath)
	if err != nil {
		return RouteReport{}, fmt.Errorf("read brief.md: %w", err)
	}
	handoffPath := filepath.Join(sessionPath, "handoff.md")
	handoffRaw, err := os.ReadFile(handoffPath)
	if err != nil {
		return RouteReport{}, fmt.Errorf("read handoff.md: %w", err)
	}

	progress := SessionProgress{CompletedStageIDs: []string{}}
	if session.Progress != nil {
		progress = *session.Progress
		if progress.CompletedStageIDs == nil {
			progress.CompletedStageIDs = []string{}
		}
	}

	if progress.UpdatedAt.IsZero() {
		if !session.UpdatedAt.IsZero() {
			progress.UpdatedAt = session.UpdatedAt
		} else {
			progress.UpdatedAt = session.CreatedAt
		}
	}

	safety := RouteReportSafety{
		WouldExecute:           false,
		Autonomous:             false,
		ExternalAgentExecution: false,
		SkillExecution:         false,
		OpenCodeMutation:       false,
	}

	report := RouteReport{
		Session: RouteReportSession{
			ID:            session.ID,
			Intent:        session.Intent,
			Flow:          session.Flow,
			Status:        session.Status,
			InitialSkill:  session.InitialSkill,
			SchemaVersion: strings.TrimSpace(session.SchemaVersion),
			Path:          sessionPath,
		},
		Brief:   RouteReportBrief{Path: briefPath, Markdown: string(briefRaw)},
		Handoff: RouteReportHandoff{Path: handoffPath, Markdown: string(handoffRaw)},
		Plan:    RouteReportPlan{Flow: session.Plan.FlowID, Stages: append([]ExecutionStage{}, session.Plan.Stages...)},
		Progress: RouteReportProgress{
			CurrentStageID:    strings.TrimSpace(progress.CurrentStageID),
			CompletedStageIDs: append([]string{}, progress.CompletedStageIDs...),
			UpdatedAt:         formatReportTime(progress.UpdatedAt),
		},
		Lifecycle: RouteReportLifecycle{
			SchemaVersion: strings.TrimSpace(session.SchemaVersion),
			CreatedAt:     formatReportTime(session.CreatedAt),
			UpdatedAt:     formatReportTime(session.UpdatedAt),
			Archived:      session.Archived,
			ArchivedAt:    formatReportOptionalTime(session.ArchivedAt),
			RestoredAt:    formatReportOptionalTime(session.RestoredAt),
		},
		Safety:                 safety,
		WouldExecute:           false,
		Autonomous:             false,
		ExternalAgentExecution: false,
		SkillExecution:         false,
		OpenCodeMutation:       false,
	}
	return report, nil
}

func RenderRouteReportMarkdown(report RouteReport) string {
	var b strings.Builder
	b.WriteString("# Route Report\n\n")
	b.WriteString("## Session\n")
	b.WriteString("- ID: " + report.Session.ID + "\n")
	b.WriteString("- Intent: " + report.Session.Intent + "\n")
	b.WriteString("- Flow: " + report.Session.Flow + "\n")
	b.WriteString("- Status: " + report.Session.Status + "\n")
	b.WriteString("- Initial skill: " + report.Session.InitialSkill + "\n")
	b.WriteString("- Schema version: " + firstNonEmptyValue(report.Session.SchemaVersion, "<legacy>") + "\n")
	b.WriteString("- Path: " + report.Session.Path + "\n\n")

	b.WriteString("## Brief\n")
	b.WriteString("- Path: " + report.Brief.Path + "\n")
	b.WriteString(report.Brief.Markdown)
	if !strings.HasSuffix(report.Brief.Markdown, "\n") {
		b.WriteString("\n")
	}
	b.WriteString("\n## Handoff\n")
	b.WriteString("- Path: " + report.Handoff.Path + "\n")
	b.WriteString(report.Handoff.Markdown)
	if !strings.HasSuffix(report.Handoff.Markdown, "\n") {
		b.WriteString("\n")
	}

	b.WriteString("\n## Plan\n")
	b.WriteString("- Flow: " + report.Plan.Flow + "\n")
	b.WriteString("- Stages:\n")
	for _, stage := range report.Plan.Stages {
		b.WriteString(fmt.Sprintf("  - %d/%s (%s)\n", stage.Order, stage.ID, stage.Skill))
	}

	b.WriteString("\n## Progress\n")
	b.WriteString("- Current stage: " + firstNonEmptyValue(report.Progress.CurrentStageID, "<none>") + "\n")
	b.WriteString("- Completed stages:\n")
	if len(report.Progress.CompletedStageIDs) == 0 {
		b.WriteString("  - (none)\n")
	} else {
		for _, id := range report.Progress.CompletedStageIDs {
			b.WriteString("  - " + id + "\n")
		}
	}
	b.WriteString("- Updated at: " + report.Progress.UpdatedAt + "\n")

	b.WriteString("\n## Lifecycle\n")
	b.WriteString("- Created at: " + report.Lifecycle.CreatedAt + "\n")
	b.WriteString("- Updated at: " + report.Lifecycle.UpdatedAt + "\n")
	b.WriteString("- Schema version: " + firstNonEmptyValue(report.Lifecycle.SchemaVersion, "<legacy>") + "\n")
	b.WriteString(fmt.Sprintf("- Archived: %t\n", report.Lifecycle.Archived))
	if strings.TrimSpace(report.Lifecycle.ArchivedAt) != "" {
		b.WriteString("- Archived at: " + report.Lifecycle.ArchivedAt + "\n")
	}
	if strings.TrimSpace(report.Lifecycle.RestoredAt) != "" {
		b.WriteString("- Restored at: " + report.Lifecycle.RestoredAt + "\n")
	}

	b.WriteString("\n## Safety\n")
	b.WriteString(fmt.Sprintf("- would_execute: %t\n", report.Safety.WouldExecute))
	b.WriteString(fmt.Sprintf("- autonomous: %t\n", report.Safety.Autonomous))
	b.WriteString(fmt.Sprintf("- external_agent_execution: %t\n", report.Safety.ExternalAgentExecution))
	b.WriteString(fmt.Sprintf("- skill_execution: %t\n", report.Safety.SkillExecution))
	b.WriteString(fmt.Sprintf("- opencode_mutation: %t\n", report.Safety.OpenCodeMutation))

	return b.String()
}

func firstNonEmptyValue(v string, alt string) string {
	if strings.TrimSpace(v) == "" {
		return alt
	}
	return strings.TrimSpace(v)
}

func formatReportTime(ts time.Time) string {
	if ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func formatReportOptionalTime(ts *time.Time) string {
	if ts == nil || ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}
