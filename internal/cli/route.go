package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bitsentry-ai/internal/capabilities"
	"bitsentry-ai/internal/orchestrator"
	"github.com/spf13/cobra"
)

type routePreviewJSONOutput struct {
	Intent         string                    `json:"intent"`
	Flow           string                    `json:"flow"`
	InitialSkill   string                    `json:"initial_skill"`
	Plan           routePlanJSONOutput       `json:"plan"`
	Warnings       []string                  `json:"warnings"`
	SessionPreview routeSessionPreviewOutput `json:"session_preview"`
}

type routeDecideJSONOutput struct {
	Input                    string   `json:"input"`
	MatchedIntent            string   `json:"matched_intent"`
	MatchedSignals           []string `json:"matched_signals"`
	Decision                 string   `json:"decision"`
	RecommendedFlow          string   `json:"recommended_flow,omitempty"`
	RecommendedRoles         []string `json:"recommended_roles"`
	RecommendedSkills        []string `json:"recommended_skills"`
	PrimarySkills            []string `json:"primary_skills"`
	SecondarySkills          []string `json:"secondary_skills"`
	DeferredSkills           []string `json:"deferred_skills"`
	PrimaryRoles             []string `json:"primary_roles"`
	SecondaryRoles           []string `json:"secondary_roles"`
	CapabilityReason         string   `json:"capability_reason"`
	CapabilityGates          []string `json:"capability_gates"`
	Confidence               string   `json:"confidence"`
	Reason                   string   `json:"reason"`
	RequiresConfirmation     bool     `json:"requires_confirmation"`
	RequiresBoundedDiscovery bool     `json:"requires_bounded_discovery"`
	Gates                    []string `json:"gates"`
	Notes                    []string `json:"notes"`
	WouldPersist             bool     `json:"would_persist"`
	WouldExecute             bool     `json:"would_execute"`
}

type routeInspectJSONOutput struct {
	Flows []routeInspectFlowJSONOutput `json:"flows"`
}

type routeStartJSONOutput struct {
	SessionID    string `json:"session_id"`
	Flow         string `json:"flow"`
	InitialSkill string `json:"initial_skill"`
	Status       string `json:"status"`
	Path         string `json:"path"`
	NextCommand  string `json:"next_command"`
}

type routeStatusJSONOutput struct {
	SessionID    string `json:"session_id"`
	Intent       string `json:"intent"`
	Flow         string `json:"flow"`
	InitialSkill string `json:"initial_skill"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	Path         string `json:"path"`
}

type routeListJSONOutput struct {
	Sessions     []routeListItemJSONOutput `json:"sessions"`
	Total        int                       `json:"total"`
	WouldExecute bool                      `json:"would_execute"`
}

type routeListItemJSONOutput struct {
	SessionID      string `json:"session_id"`
	Flow           string `json:"flow"`
	Status         string `json:"status"`
	Archived       bool   `json:"archived"`
	Intent         string `json:"intent"`
	CurrentStage   string `json:"current_stage"`
	CompletedTotal string `json:"completed_total"`
	UpdatedAt      string `json:"updated_at"`
	Path           string `json:"path"`
}

type routeArchiveRestoreJSONOutput struct {
	SessionID    string `json:"session_id"`
	Archived     bool   `json:"archived"`
	ArchivedAt   string `json:"archived_at,omitempty"`
	RestoredAt   string `json:"restored_at,omitempty"`
	UpdatedAt    string `json:"updated_at"`
	Path         string `json:"path"`
	WouldExecute bool   `json:"would_execute"`
}

type routeValidationCheckJSONOutput struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type routeValidateJSONOutput struct {
	SessionID    string                           `json:"session_id"`
	Path         string                           `json:"path"`
	Valid        bool                             `json:"valid"`
	Errors       []string                         `json:"errors"`
	Warnings     []string                         `json:"warnings"`
	Checks       []routeValidationCheckJSONOutput `json:"checks"`
	WouldExecute bool                             `json:"would_execute"`
}

type routeAuditJSONOutput struct {
	Sessions        []routeValidateJSONOutput `json:"sessions"`
	Total           int                       `json:"total"`
	Passed          int                       `json:"passed"`
	Failed          int                       `json:"failed"`
	Warnings        int                       `json:"warnings"`
	IncludeArchived bool                      `json:"include_archived"`
	WouldExecute    bool                      `json:"would_execute"`
}

type routeRepairJSONOutput struct {
	SessionID    string   `json:"session_id"`
	DryRun       bool     `json:"dry_run"`
	Applied      bool     `json:"applied"`
	Repairable   bool     `json:"repairable"`
	Changes      []string `json:"changes"`
	Unrepairable []string `json:"unrepairable"`
	Warnings     []string `json:"warnings"`
	WouldExecute bool     `json:"would_execute"`
}

type routeMigrateJSONOutput struct {
	SessionID    string   `json:"session_id"`
	DryRun       bool     `json:"dry_run"`
	Applied      bool     `json:"applied"`
	FromVersion  string   `json:"from_version,omitempty"`
	ToVersion    string   `json:"to_version"`
	Changes      []string `json:"changes"`
	Warnings     []string `json:"warnings"`
	WouldExecute bool     `json:"would_execute"`
}

type routeCleanupPolicyJSONOutput struct {
	Mode            string `json:"mode"`
	OlderThan       string `json:"older_than,omitempty"`
	Status          string `json:"status,omitempty"`
	Flow            string `json:"flow,omitempty"`
	CompletedOnly   bool   `json:"completed_only"`
	IncludeArchived bool   `json:"include_archived"`
}

type routeCleanupCandidateJSONOutput struct {
	SessionID string   `json:"session_id"`
	Flow      string   `json:"flow"`
	Status    string   `json:"status"`
	Archived  bool     `json:"archived"`
	UpdatedAt string   `json:"updated_at"`
	Path      string   `json:"path"`
	Reasons   []string `json:"reasons"`
}

type routeCleanupCountsJSONOutput struct {
	Scanned    int `json:"scanned"`
	Candidates int `json:"candidates"`
	Archived   int `json:"archived"`
	Skipped    int `json:"skipped"`
}

type routeCleanupJSONOutput struct {
	Policy       routeCleanupPolicyJSONOutput      `json:"policy"`
	Counts       routeCleanupCountsJSONOutput      `json:"counts"`
	Warnings     []string                          `json:"warnings"`
	Candidates   []routeCleanupCandidateJSONOutput `json:"candidates"`
	WouldExecute bool                              `json:"would_execute"`
}

type routeHandoffJSONOutput struct {
	SessionID   string   `json:"session_id"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Reason      string   `json:"reason"`
	NextSteps   []string `json:"next_steps"`
	SafetyNotes []string `json:"safety_notes"`
}

type routeResumeJSONOutput struct {
	SessionID      string                    `json:"session_id"`
	Intent         string                    `json:"intent"`
	Flow           string                    `json:"flow"`
	Status         string                    `json:"status"`
	InitialSkill   string                    `json:"initial_skill"`
	TotalStages    int                       `json:"total_stages"`
	CurrentStage   *routePlanStageJSONOutput `json:"current_stage,omitempty"`
	NextStage      *routePlanStageJSONOutput `json:"next_stage,omitempty"`
	Path           string                    `json:"path"`
	SafetyReminder string                    `json:"safety_reminder"`
	WouldExecute   bool                      `json:"would_execute"`
}

type routeNextJSONOutput struct {
	SessionID         string                    `json:"session_id"`
	Flow              string                    `json:"flow"`
	Status            string                    `json:"status"`
	NextSkill         string                    `json:"next_skill,omitempty"`
	Stage             *routePlanStageJSONOutput `json:"stage,omitempty"`
	Reason            string                    `json:"reason"`
	SuggestedCommands []string                  `json:"suggested_commands"`
	WouldExecute      bool                      `json:"would_execute"`
}

type routeProgressJSONOutput struct {
	SessionID    string                         `json:"session_id"`
	Flow         string                         `json:"flow"`
	Status       string                         `json:"status"`
	CurrentStage *routePlanStageJSONOutput      `json:"current_stage,omitempty"`
	Stages       []routeProgressStageJSONOutput `json:"stages"`
	UpdatedAt    string                         `json:"updated_at"`
	WouldExecute bool                           `json:"would_execute"`
}

type routeProgressStageJSONOutput struct {
	Order int    `json:"order"`
	ID    string `json:"id"`
	Skill string `json:"skill,omitempty"`
	State string `json:"state"`
}

type routeMarkJSONOutput struct {
	SessionID    string `json:"session_id"`
	StageID      string `json:"stage_id"`
	State        string `json:"state"`
	WouldExecute bool   `json:"would_execute"`
}

type routeInspectFlowJSONOutput struct {
	Flow         string                     `json:"flow"`
	StagesCount  int                        `json:"stages_count"`
	InitialSkill string                     `json:"initial_skill,omitempty"`
	Stages       []routePlanStageJSONOutput `json:"stages"`
	Skills       []string                   `json:"skills"`
}

type routePlanJSONOutput struct {
	Flow   string                        `json:"flow"`
	Stages []orchestrator.ExecutionStage `json:"stages"`
}

type routePlanStageJSONOutput struct {
	Order       int    `json:"order"`
	ID          string `json:"id"`
	Skill       string `json:"skill,omitempty"`
	Description string `json:"description,omitempty"`
}

type routeSessionPreviewOutput struct {
	WouldCreateSession bool `json:"would_create_session"`
	WouldPersist       bool `json:"would_persist"`
	WouldExecute       bool `json:"would_execute"`
}

func newRouteCmd(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route",
		Short: "Inspect and preview declarative routing plans",
	}

	cmd.AddCommand(
		newRouteInspectCmd(rt),
		newRoutePreviewCmd(rt),
		newRouteDecideCmd(rt),
		newRouteStartCmd(rt),
		newRouteReportCmd(rt),
		newRouteListCmd(rt),
		newRouteArchiveCmd(rt),
		newRouteRestoreCmd(rt),
		newRouteStatusCmd(rt),
		newRouteHandoffCmd(rt),
		newRouteResumeCmd(rt),
		newRouteNextCmd(rt),
		newRouteProgressCmd(rt),
		newRouteMarkCurrentCmd(rt),
		newRouteMarkDoneCmd(rt),
		newRouteValidateCmd(rt),
		newRouteMigrateCmd(rt),
		newRouteRepairCmd(rt),
		newRouteAuditCmd(rt),
		newRouteCleanupCmd(rt),
	)

	return cmd
}

func newRouteDecideCmd(rt *Runtime) *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "decide [prompt]",
		Short: "Preview route decision envelope without side effects",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := strings.TrimSpace(args[0])
			if input == "" {
				return fmt.Errorf("prompt must not be empty")
			}

			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			result, err := capabilities.BuildRouteDecisionPreview(root, input)
			if err != nil {
				return err
			}

			payload := routeDecideJSONOutput{
				Input:                    result.Input,
				MatchedIntent:            result.MatchedIntent,
				MatchedSignals:           append([]string{}, result.MatchedSignals...),
				Decision:                 result.Decision,
				RecommendedFlow:          result.RecommendedFlow,
				RecommendedRoles:         append([]string{}, result.RecommendedRoles...),
				RecommendedSkills:        append([]string{}, result.RecommendedSkills...),
				PrimarySkills:            append([]string{}, result.PrimarySkills...),
				SecondarySkills:          append([]string{}, result.SecondarySkills...),
				DeferredSkills:           append([]string{}, result.DeferredSkills...),
				PrimaryRoles:             append([]string{}, result.PrimaryRoles...),
				SecondaryRoles:           append([]string{}, result.SecondaryRoles...),
				CapabilityReason:         result.CapabilityReason,
				CapabilityGates:          append([]string{}, result.CapabilityGates...),
				Confidence:               result.Confidence,
				Reason:                   result.Reason,
				RequiresConfirmation:     result.RequiresConfirmation,
				RequiresBoundedDiscovery: result.RequiresBoundedDiscovery,
				Gates:                    append([]string{}, result.Gates...),
				Notes:                    append([]string{}, result.Notes...),
				WouldPersist:             false,
				WouldExecute:             false,
			}

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route decide output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintln(out, "Route decision preview")
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Input:")
			_, _ = fmt.Fprintf(out, "%q\n\n", payload.Input)
			_, _ = fmt.Fprintln(out, "Matched intent:")
			_, _ = fmt.Fprintf(out, "%s\n\n", payload.MatchedIntent)
			_, _ = fmt.Fprintln(out, "Matched signals:")
			if len(payload.MatchedSignals) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, s := range payload.MatchedSignals {
					_, _ = fmt.Fprintf(out, "- %s\n", s)
				}
			}
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Decision:")
			_, _ = fmt.Fprintf(out, "%s\n\n", payload.Decision)
			_, _ = fmt.Fprintln(out, "Recommended flow:")
			_, _ = fmt.Fprintf(out, "%s\n\n", firstNonEmpty(payload.RecommendedFlow, "none"))

			_, _ = fmt.Fprintln(out, "Primary roles:")
			if len(payload.PrimaryRoles) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, role := range payload.PrimaryRoles {
					_, _ = fmt.Fprintf(out, "- %s\n", role)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Primary skills:")
			if len(payload.PrimarySkills) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, skill := range payload.PrimarySkills {
					_, _ = fmt.Fprintf(out, "- %s\n", skill)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Recommended roles:")
			if len(payload.RecommendedRoles) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, role := range payload.RecommendedRoles {
					_, _ = fmt.Fprintf(out, "- %s\n", role)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Secondary roles:")
			if len(payload.SecondaryRoles) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, role := range payload.SecondaryRoles {
					_, _ = fmt.Fprintf(out, "- %s\n", role)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Recommended skills:")
			if len(payload.RecommendedSkills) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, skill := range payload.RecommendedSkills {
					_, _ = fmt.Fprintf(out, "- %s\n", skill)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Secondary skills:")
			if len(payload.SecondarySkills) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, skill := range payload.SecondarySkills {
					_, _ = fmt.Fprintf(out, "- %s\n", skill)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Deferred skills:")
			if len(payload.DeferredSkills) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, skill := range payload.DeferredSkills {
					_, _ = fmt.Fprintf(out, "- %s\n", skill)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Capability reason:")
			_, _ = fmt.Fprintf(out, "%s\n\n", payload.CapabilityReason)
			_, _ = fmt.Fprintln(out, "Capability gates:")
			if len(payload.CapabilityGates) == 0 {
				_, _ = fmt.Fprintln(out, "- none")
			} else {
				for _, gate := range payload.CapabilityGates {
					_, _ = fmt.Fprintf(out, "- %s\n", gate)
				}
			}
			_, _ = fmt.Fprintln(out)

			_, _ = fmt.Fprintln(out, "Confidence:")
			_, _ = fmt.Fprintf(out, "%s\n\n", payload.Confidence)
			_, _ = fmt.Fprintln(out, "Gates:")
			for _, gate := range payload.Gates {
				_, _ = fmt.Fprintf(out, "- %s\n", gate)
			}
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Reason:")
			_, _ = fmt.Fprintf(out, "%s\n", payload.Reason)
			if len(payload.Notes) > 0 {
				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintln(out, "Notes:")
				for _, n := range payload.Notes {
					_, _ = fmt.Fprintf(out, "- %s\n", n)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteMigrateCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var dryRun bool
	var apply bool
	var confirm bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "migrate --session <session-id>",
		Short: "Migrate persisted route session schema declaratively",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun && apply {
				return fmt.Errorf("--dry-run and --apply are mutually exclusive")
			}
			if !dryRun && !apply {
				return fmt.Errorf("route migrate requires one mode: --dry-run or --apply")
			}
			if apply && !confirm {
				return fmt.Errorf("route migrate --apply requires --confirm")
			}

			result, err := runRouteMigrate(strings.TrimSpace(sessionID), dryRun, apply)
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route migrate output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			mode := "dry-run"
			if apply {
				mode = "apply"
			}
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", result.SessionID)
			_, _ = fmt.Fprintf(out, "Mode: %s\n", mode)
			_, _ = fmt.Fprintf(out, "Applied: %t\n", result.Applied)
			_, _ = fmt.Fprintf(out, "From schema: %s\n", firstNonEmpty(result.FromVersion, "<legacy>"))
			_, _ = fmt.Fprintf(out, "To schema: %s\n", result.ToVersion)
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", result.WouldExecute)
			if len(result.Changes) > 0 {
				_, _ = fmt.Fprintln(out, "Changes:")
				for _, change := range result.Changes {
					_, _ = fmt.Fprintf(out, "- %s\n", change)
				}
			}
			if len(result.Warnings) > 0 {
				_, _ = fmt.Fprintln(out, "Warnings:")
				for _, warning := range result.Warnings {
					_, _ = fmt.Fprintf(out, "- %s\n", warning)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview schema migration changes without writing")
	cmd.Flags().BoolVar(&apply, "apply", false, "Apply schema migration to session.json")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Required with --apply to execute migration")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteRepairCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var dryRun bool
	var apply bool
	var confirm bool
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "repair --session <session-id>",
		Short: "Repair persisted route session metadata safely",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRun && apply {
				return fmt.Errorf("--dry-run and --apply are mutually exclusive")
			}
			if !dryRun && !apply {
				return fmt.Errorf("route repair requires one mode: --dry-run or --apply")
			}
			if apply && !confirm {
				return fmt.Errorf("route repair --apply requires --confirm")
			}

			result, err := runRouteRepair(strings.TrimSpace(sessionID), dryRun, apply)
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route repair output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			mode := "dry-run"
			if apply {
				mode = "apply"
			}
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", result.SessionID)
			_, _ = fmt.Fprintf(out, "Mode: %s\n", mode)
			_, _ = fmt.Fprintf(out, "Applied: %t\n", result.Applied)
			_, _ = fmt.Fprintf(out, "Repairable: %t\n", result.Repairable)
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", result.WouldExecute)
			if len(result.Changes) > 0 {
				_, _ = fmt.Fprintln(out, "Changes:")
				for _, c := range result.Changes {
					_, _ = fmt.Fprintf(out, "- %s\n", c)
				}
			}
			if len(result.Unrepairable) > 0 {
				_, _ = fmt.Fprintln(out, "Unrepairable:")
				for _, item := range result.Unrepairable {
					_, _ = fmt.Fprintf(out, "- %s\n", item)
				}
			}
			if len(result.Warnings) > 0 {
				_, _ = fmt.Fprintln(out, "Warnings:")
				for _, w := range result.Warnings {
					_, _ = fmt.Fprintf(out, "- %s\n", w)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview repair changes without writing")
	cmd.Flags().BoolVar(&apply, "apply", false, "Apply repair changes to session metadata files")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Required with --apply to execute repair")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteCleanupCmd(rt *Runtime) *cobra.Command {
	var dryRun bool
	var apply bool
	var confirm bool
	var jsonOut bool
	var olderThan string
	var status string
	var flow string
	var completedOnly bool
	var includeArchived bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Preview or archive persisted sessions declaratively",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if apply && dryRun {
				return fmt.Errorf("--dry-run and --apply are mutually exclusive")
			}
			if !apply && !dryRun {
				dryRun = true
			}
			if apply && !confirm {
				return fmt.Errorf("route cleanup --apply requires --confirm")
			}

			trimmedFlow := strings.TrimSpace(flow)
			trimmedStatus := strings.TrimSpace(status)
			trimmedOlderThan := strings.TrimSpace(olderThan)
			hasNarrowFilter := trimmedFlow != "" || trimmedStatus != "" || completedOnly || trimmedOlderThan != ""
			if apply && !hasNarrowFilter {
				return fmt.Errorf("route cleanup --apply requires at least one filter: --older-than, --status, --flow, or --completed-only")
			}

			var olderThanDuration time.Duration
			if trimmedOlderThan != "" {
				parsed, err := time.ParseDuration(trimmedOlderThan)
				if err != nil {
					return fmt.Errorf("invalid --older-than duration %q: %w", trimmedOlderThan, err)
				}
				if parsed <= 0 {
					return fmt.Errorf("--older-than must be greater than 0")
				}
				olderThanDuration = parsed
			}

			result, err := runRouteCleanup(routeCleanupPolicyJSONOutput{
				Mode:            firstNonEmpty(map[bool]string{true: "apply", false: "dry-run"}[apply], "dry-run"),
				OlderThan:       trimmedOlderThan,
				Status:          trimmedStatus,
				Flow:            trimmedFlow,
				CompletedOnly:   completedOnly,
				IncludeArchived: includeArchived,
			}, olderThanDuration, apply)
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route cleanup output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Mode: %s\n", result.Policy.Mode)
			_, _ = fmt.Fprintf(out, "Scanned: %d\nCandidates: %d\nArchived: %d\nSkipped: %d\n", result.Counts.Scanned, result.Counts.Candidates, result.Counts.Archived, result.Counts.Skipped)
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", result.WouldExecute)
			if len(result.Warnings) > 0 {
				_, _ = fmt.Fprintln(out, "Warnings:")
				for _, warning := range result.Warnings {
					_, _ = fmt.Fprintf(out, "- %s\n", warning)
				}
			}
			if len(result.Candidates) == 0 {
				_, _ = fmt.Fprintln(out, "No cleanup candidates found.")
				return nil
			}
			_, _ = fmt.Fprintln(out, "Candidates:")
			for _, candidate := range result.Candidates {
				_, _ = fmt.Fprintf(out, "- %s (%s/%s archived=%t)\n", candidate.SessionID, candidate.Flow, candidate.Status, candidate.Archived)
				_, _ = fmt.Fprintf(out, "  Updated: %s\n", candidate.UpdatedAt)
				_, _ = fmt.Fprintf(out, "  Path: %s\n", candidate.Path)
				_, _ = fmt.Fprintf(out, "  Reasons: %s\n", strings.Join(candidate.Reasons, "; "))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview cleanup candidates without mutating sessions")
	cmd.Flags().BoolVar(&apply, "apply", false, "Archive matching sessions in session.json metadata")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Required with --apply to execute archival")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&olderThan, "older-than", "", "Filter by updated_at older than duration (e.g. 24h)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by session status")
	cmd.Flags().StringVar(&flow, "flow", "", "Filter by flow ID")
	cmd.Flags().BoolVar(&completedOnly, "completed-only", false, "Include only fully completed sessions")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived sessions in candidate preview/apply")
	return cmd
}

func runRouteCleanup(policy routeCleanupPolicyJSONOutput, olderThanDuration time.Duration, apply bool) (routeCleanupJSONOutput, error) {
	root, err := resolveRepoRootWithAssets(".")
	if err != nil {
		return routeCleanupJSONOutput{}, fmt.Errorf("resolve assets root: %w", err)
	}
	store := orchestrator.NewLocalSessionStore(root)
	sessionsRoot := filepath.Join(store.Root(), ".bitsentry-ai", "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return routeCleanupJSONOutput{Policy: policy, Counts: routeCleanupCountsJSONOutput{}, Warnings: []string{}, Candidates: []routeCleanupCandidateJSONOutput{}, WouldExecute: false}, nil
		}
		return routeCleanupJSONOutput{}, fmt.Errorf("read sessions directory: %w", err)
	}

	now := time.Now().UTC()
	cutoff := time.Time{}
	if olderThanDuration > 0 {
		cutoff = now.Add(-olderThanDuration)
	}

	warnings := []string{}
	candidates := make([]routeCleanupCandidateJSONOutput, 0, len(entries))
	scanned := 0
	skipped := 0
	archivedCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionID := strings.TrimSpace(entry.Name())
		session, sessionPath, loadErr := store.LoadSession(sessionID)
		if loadErr != nil {
			skipped++
			warnings = append(warnings, fmt.Sprintf("skipping session %q: %v", sessionID, loadErr))
			continue
		}
		scanned++

		reasons := cleanupMatchReasons(session, policy, cutoff)
		if len(reasons) == 0 {
			continue
		}

		if apply {
			nowTS := time.Now().UTC()
			session.Archived = true
			if session.ArchivedAt == nil || session.ArchivedAt.IsZero() {
				session.ArchivedAt = &nowTS
			}
			session.UpdatedAt = nowTS
			if _, saveErr := store.SaveSessionMetadata(session); saveErr != nil {
				skipped++
				warnings = append(warnings, fmt.Sprintf("failed to archive session %q: %v", session.ID, saveErr))
				continue
			}
			archivedCount++
		}

		candidates = append(candidates, routeCleanupCandidateJSONOutput{
			SessionID: session.ID,
			Flow:      strings.TrimSpace(session.Flow),
			Status:    strings.TrimSpace(session.Status),
			Archived:  session.Archived,
			UpdatedAt: session.UpdatedAt.UTC().Format(time.RFC3339),
			Path:      sessionPath,
			Reasons:   reasons,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].SessionID < candidates[j].SessionID
	})

	return routeCleanupJSONOutput{
		Policy:       policy,
		Counts:       routeCleanupCountsJSONOutput{Scanned: scanned, Candidates: len(candidates), Archived: archivedCount, Skipped: skipped},
		Warnings:     warnings,
		Candidates:   candidates,
		WouldExecute: false,
	}, nil
}

func cleanupMatchReasons(session orchestrator.Session, policy routeCleanupPolicyJSONOutput, cutoff time.Time) []string {
	if session.Archived && !policy.IncludeArchived {
		return nil
	}

	reasons := []string{}
	if policy.Flow != "" {
		if strings.TrimSpace(session.Flow) != policy.Flow {
			return nil
		}
		reasons = append(reasons, fmt.Sprintf("flow=%s", policy.Flow))
	}
	if policy.Status != "" {
		if strings.TrimSpace(session.Status) != policy.Status {
			return nil
		}
		reasons = append(reasons, fmt.Sprintf("status=%s", policy.Status))
	}
	if !cutoff.IsZero() {
		updated := session.UpdatedAt
		if updated.IsZero() {
			updated = session.CreatedAt
		}
		if updated.IsZero() || !updated.Before(cutoff) {
			return nil
		}
		reasons = append(reasons, fmt.Sprintf("updated_at_before=%s", cutoff.UTC().Format(time.RFC3339)))
	}
	if policy.CompletedOnly {
		if !isSessionCompleted(session) {
			return nil
		}
		reasons = append(reasons, "completed_only=true")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "matched_default_policy")
	}
	return reasons
}

func isSessionCompleted(session orchestrator.Session) bool {
	progress := normalizedProgress(session)
	if len(session.Plan.Stages) == 0 {
		return false
	}
	completed := map[string]bool{}
	for _, doneID := range progress.CompletedStageIDs {
		completed[strings.TrimSpace(doneID)] = true
	}
	for _, stage := range session.Plan.Stages {
		if !completed[strings.TrimSpace(stage.ID)] {
			return false
		}
	}
	return true
}

func newRouteValidateCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "validate --session <session-id>",
		Short: "Validate persisted route session read-only",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := validateSingleSession(strings.TrimSpace(sessionID))
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route validate output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", result.SessionID)
			_, _ = fmt.Fprintf(out, "Path: %s\n", result.Path)
			_, _ = fmt.Fprintf(out, "Valid: %t\n", result.Valid)
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", result.WouldExecute)
			_, _ = fmt.Fprintln(out, "Checks:")
			for _, check := range result.Checks {
				_, _ = fmt.Fprintf(out, "- [%s] %s\n", strings.ToUpper(check.Status), check.Name)
			}
			if len(result.Errors) > 0 {
				_, _ = fmt.Fprintln(out, "Errors:")
				for _, e := range result.Errors {
					_, _ = fmt.Fprintf(out, "- %s\n", e)
				}
			}
			if len(result.Warnings) > 0 {
				_, _ = fmt.Fprintln(out, "Warnings:")
				for _, w := range result.Warnings {
					_, _ = fmt.Fprintf(out, "- %s\n", w)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteAuditCmd(rt *Runtime) *cobra.Command {
	var jsonOut bool
	var includeArchived bool

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Audit persisted route sessions read-only",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := auditSessions(includeArchived)
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route audit output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Total: %d\nPassed: %d\nFailed: %d\nWarnings: %d\nWould execute: %t\n", payload.Total, payload.Passed, payload.Failed, payload.Warnings, payload.WouldExecute)
			for _, session := range payload.Sessions {
				_, _ = fmt.Fprintf(out, "\n- %s => valid=%t (errors=%d warnings=%d)\n", session.SessionID, session.Valid, len(session.Errors), len(session.Warnings))
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	cmd.Flags().BoolVar(&includeArchived, "include-archived", false, "Include archived sessions")
	return cmd
}

func newRouteListCmd(rt *Runtime) *cobra.Command {
	var jsonOut bool
	var flow string
	var status string
	var archivedOnly bool
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List persisted route sessions declaratively",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			store := orchestrator.NewLocalSessionStore(root)
			sessions, err := listPersistedSessions(store, strings.TrimSpace(flow), strings.TrimSpace(status), archivedOnly, all, cmd.ErrOrStderr())
			if err != nil {
				return err
			}

			payload := buildRouteListPayload(root, sessions)
			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route list output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			if len(payload.Sessions) == 0 {
				_, _ = fmt.Fprintln(out, "No persisted sessions found.")
				return nil
			}
			for _, item := range payload.Sessions {
				_, _ = fmt.Fprintf(out, "Session ID: %s\n", item.SessionID)
				_, _ = fmt.Fprintf(out, "Flow/Status: %s / %s\n", item.Flow, item.Status)
				_, _ = fmt.Fprintf(out, "Archived: %t\n", item.Archived)
				_, _ = fmt.Fprintf(out, "Intent: %s\n", item.Intent)
				_, _ = fmt.Fprintf(out, "Current stage: %s\n", item.CurrentStage)
				_, _ = fmt.Fprintf(out, "Completed: %s\n", item.CompletedTotal)
				_, _ = fmt.Fprintf(out, "Updated at: %s\n", item.UpdatedAt)
				_, _ = fmt.Fprintf(out, "Path: %s\n\n", item.Path)
			}
			_, _ = fmt.Fprintf(out, "Total: %d\n", payload.Total)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&flow, "flow", "", "Filter by flow ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by session status")
	cmd.Flags().BoolVar(&archivedOnly, "archived", false, "Show only archived sessions")
	cmd.Flags().BoolVar(&all, "all", false, "Show both active and archived sessions")
	return cmd
}

func newRouteArchiveCmd(rt *Runtime) *cobra.Command {
	return newRouteArchiveRestoreCmd("archive", true)
}

func newRouteRestoreCmd(rt *Runtime) *cobra.Command {
	return newRouteArchiveRestoreCmd("restore", false)
}

func newRouteArchiveRestoreCmd(use string, archived bool) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s --session <session-id>", use),
		Short: fmt.Sprintf("%s a persisted route session declaratively", strings.Title(use)),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, session, sessionPath, err := loadRouteSession(sessionID)
			if err != nil {
				return err
			}
			now := time.Now().UTC()
			session.Archived = archived
			session.UpdatedAt = now
			if archived {
				session.ArchivedAt = &now
			} else {
				session.RestoredAt = &now
			}
			if _, err := store.SaveSessionMetadata(session); err != nil {
				return err
			}

			payload := routeArchiveRestoreJSONOutput{
				SessionID:    session.ID,
				Archived:     session.Archived,
				ArchivedAt:   formatOptionalTime(session.ArchivedAt),
				RestoredAt:   formatOptionalTime(session.RestoredAt),
				UpdatedAt:    session.UpdatedAt.Format(time.RFC3339),
				Path:         sessionPath,
				WouldExecute: false,
			}

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route %s output: %w", use, err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			verb := "archived"
			if !archived {
				verb = "restored"
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session ID: %s\nState: %s\nPath: %s\nWould execute: false\n", payload.SessionID, verb, payload.Path)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteProgressCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "progress --session <session-id>",
		Short: "Show declarative progress for a persisted session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, session, _, err := loadRouteSession(sessionID)
			if err != nil {
				return err
			}
			_ = store
			payload := buildRouteProgressPayload(session)

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route progress output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", payload.SessionID)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", payload.Flow)
			_, _ = fmt.Fprintf(out, "Status: %s\n", payload.Status)
			_, _ = fmt.Fprintf(out, "Current stage: %s\n", renderSuggestedStage(payload.CurrentStage))
			_, _ = fmt.Fprintln(out, "Stages:")
			for _, stage := range payload.Stages {
				_, _ = fmt.Fprintf(out, "- %d/%s (%s) => %s\n", stage.Order, stage.ID, firstNonEmpty(stage.Skill, "<unset>"), stage.State)
			}
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", payload.WouldExecute)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteMarkCurrentCmd(rt *Runtime) *cobra.Command {
	return newRouteMarkCmd("mark-current", "current")
}

func newRouteMarkDoneCmd(rt *Runtime) *cobra.Command {
	return newRouteMarkCmd("mark-done", "done")
}

func newRouteMarkCmd(use string, targetState string) *cobra.Command {
	var sessionID string
	var stageID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s --session <session-id> --stage <stage-id>", use),
		Short: "Update declarative stage progress for a session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, session, _, err := loadRouteSession(sessionID)
			if err != nil {
				return err
			}
			updated, err := applyStageMark(session, strings.TrimSpace(stageID), targetState)
			if err != nil {
				return err
			}
			if _, err := store.SaveSession(updated); err != nil {
				return err
			}

			state := progressStateForStage(updated, strings.TrimSpace(stageID))
			payload := routeMarkJSONOutput{SessionID: updated.ID, StageID: strings.TrimSpace(stageID), State: state, WouldExecute: false}
			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route mark output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Session ID: %s\nStage: %s\nState: %s\nWould execute: false\n", payload.SessionID, payload.StageID, payload.State)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().StringVar(&stageID, "stage", "", "Stage ID from the declared plan")
	_ = cmd.MarkFlagRequired("stage")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteResumeCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "resume --session <session-id>",
		Short: "Resume a persisted route session declaratively",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			store := orchestrator.NewLocalSessionStore(root)
			session, sessionPath, err := store.LoadSession(sessionID)
			if err != nil {
				return err
			}

			current, next := suggestedStagesFromProgress(session)
			safety := "Declarative only: no execution is performed and no external agents are called."
			payload := routeResumeJSONOutput{
				SessionID:      session.ID,
				Intent:         session.Intent,
				Flow:           session.Flow,
				Status:         session.Status,
				InitialSkill:   session.InitialSkill,
				TotalStages:    len(session.Plan.Stages),
				CurrentStage:   current,
				NextStage:      next,
				Path:           sessionPath,
				SafetyReminder: safety,
				WouldExecute:   false,
			}

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route resume output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", payload.SessionID)
			_, _ = fmt.Fprintf(out, "Intent: %s\n", payload.Intent)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", payload.Flow)
			_, _ = fmt.Fprintf(out, "Status: %s\n", payload.Status)
			_, _ = fmt.Fprintf(out, "Initial skill: %s\n", payload.InitialSkill)
			_, _ = fmt.Fprintf(out, "Total stages: %d\n", payload.TotalStages)
			_, _ = fmt.Fprintf(out, "Current stage: %s\n", renderSuggestedStage(payload.CurrentStage))
			_, _ = fmt.Fprintf(out, "Next stage: %s\n", renderSuggestedStage(payload.NextStage))
			_, _ = fmt.Fprintf(out, "Path: %s\n", payload.Path)
			_, _ = fmt.Fprintf(out, "Safety reminder: %s\n", payload.SafetyReminder)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteNextCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "next --session <session-id>",
		Short: "Recommend the next declarative step for a session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			store := orchestrator.NewLocalSessionStore(root)
			session, _, err := store.LoadSession(sessionID)
			if err != nil {
				return err
			}

			current, _ := suggestedStagesFromProgress(session)
			nextSkill := strings.TrimSpace(session.InitialSkill)
			if current != nil && strings.TrimSpace(current.Skill) != "" {
				nextSkill = strings.TrimSpace(current.Skill)
			}

			reason := "Session progress is declarative-only; recommend the first pending stage without executing anything."
			if current == nil {
				nextSkill = ""
				reason = "All declared stages are marked done; there is no next stage to recommend."
			}
			suggested := []string{
				fmt.Sprintf("bitsentry route handoff --session %s", session.ID),
				fmt.Sprintf("bitsentry route resume --session %s", session.ID),
			}

			payload := routeNextJSONOutput{
				SessionID:         session.ID,
				Flow:              session.Flow,
				Status:            session.Status,
				NextSkill:         nextSkill,
				Stage:             current,
				Reason:            reason,
				SuggestedCommands: suggested,
				WouldExecute:      false,
			}

			if jsonOut {
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route next output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", payload.SessionID)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", payload.Flow)
			_, _ = fmt.Fprintf(out, "Status: %s\n", payload.Status)
			_, _ = fmt.Fprintf(out, "Recommended skill: %s\n", firstNonEmpty(payload.NextSkill, "<unset>"))
			_, _ = fmt.Fprintf(out, "Recommended stage: %s\n", renderSuggestedStage(payload.Stage))
			_, _ = fmt.Fprintf(out, "Reason: %s\n", payload.Reason)
			_, _ = fmt.Fprintln(out, "Suggested commands:")
			for _, c := range payload.SuggestedCommands {
				_, _ = fmt.Fprintf(out, "- %s\n", c)
			}
			_, _ = fmt.Fprintf(out, "Would execute: %t\n", payload.WouldExecute)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteStartCmd(rt *Runtime) *cobra.Command {
	var flowHint string
	var flowHintAlias string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "start [intent]",
		Short: "Route and persist a local planned session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			intent := strings.TrimSpace(args[0])
			if intent == "" {
				return fmt.Errorf("intent must not be empty")
			}
			hint := strings.TrimSpace(flowHint)
			if hint == "" {
				hint = strings.TrimSpace(flowHintAlias)
			}

			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}

			o := orchestrator.New(root)
			result, err := o.Route(orchestrator.RouteRequest{Intent: orchestrator.Intent(intent), FlowHint: hint})
			if err != nil {
				return err
			}

			session := orchestrator.NewSession(intent, result)
			store := orchestrator.NewLocalSessionStore(root)
			sessionPath, err := store.SaveSession(session)
			if err != nil {
				return err
			}

			next := fmt.Sprintf("bitsentry route status --session %s", session.ID)
			if jsonOut {
				payload := routeStartJSONOutput{
					SessionID:    session.ID,
					Flow:         session.Flow,
					InitialSkill: session.InitialSkill,
					Status:       session.Status,
					Path:         sessionPath,
					NextCommand:  next,
				}
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route start output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", session.ID)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", session.Flow)
			_, _ = fmt.Fprintf(out, "Initial skill: %s\n", session.InitialSkill)
			_, _ = fmt.Fprintf(out, "Status: %s\n", session.Status)
			_, _ = fmt.Fprintf(out, "Path: %s\n", sessionPath)
			_, _ = fmt.Fprintf(out, "Next: %s\n", next)
			return nil
		},
	}

	cmd.Flags().StringVar(&flowHint, "flow-hint", "", "Force a specific flow ID (e.g., sdd, sdr, support)")
	cmd.Flags().StringVar(&flowHintAlias, "flow", "", "Alias of --flow-hint")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteStatusCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "status --session <session-id>",
		Short: "Show persisted route session status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			store := orchestrator.NewLocalSessionStore(root)
			session, sessionPath, err := store.LoadSession(sessionID)
			if err != nil {
				return err
			}

			if jsonOut {
				payload := routeStatusJSONOutput{
					SessionID:    session.ID,
					Intent:       session.Intent,
					Flow:         session.Flow,
					InitialSkill: session.InitialSkill,
					Status:       session.Status,
					CreatedAt:    session.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
					UpdatedAt:    session.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
					Path:         sessionPath,
				}
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route status output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Session ID: %s\n", session.ID)
			_, _ = fmt.Fprintf(out, "Intent: %s\n", session.Intent)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", session.Flow)
			_, _ = fmt.Fprintf(out, "Initial skill: %s\n", session.InitialSkill)
			_, _ = fmt.Fprintf(out, "Status: %s\n", session.Status)
			_, _ = fmt.Fprintf(out, "Path: %s\n", sessionPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRouteHandoffCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "handoff --session <session-id>",
		Short: "Show persisted handoff details for a session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}
			outputPath = strings.TrimSpace(outputPath)
			if outputPath != "" {
				if err := validateRouteOutputPath(root, outputPath); err != nil {
					return err
				}
			}
			store := orchestrator.NewLocalSessionStore(root)
			session, _, err := store.LoadSession(sessionID)
			if err != nil {
				return err
			}
			h := session.Handoff

			if jsonOut {
				payload := routeHandoffJSONOutput{
					SessionID:   h.SessionID,
					From:        h.From,
					To:          h.To,
					Reason:      h.Reason,
					NextSteps:   append([]string{}, h.NextSteps...),
					SafetyNotes: append([]string{}, h.SafetyNotes...),
				}
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route handoff output: %w", err)
				}
				if outputPath != "" {
					if err := writeRouteOutputFile(outputPath, append(raw, '\n')); err != nil {
						return err
					}
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			lines := []string{
				fmt.Sprintf("Session ID: %s", h.SessionID),
				fmt.Sprintf("From: %s", h.From),
				fmt.Sprintf("To: %s", h.To),
				fmt.Sprintf("Reason: %s", h.Reason),
				"Next steps:",
			}
			for _, s := range h.NextSteps {
				lines = append(lines, fmt.Sprintf("- %s", s))
			}
			lines = append(lines, "Safety notes:")
			for _, n := range h.SafetyNotes {
				lines = append(lines, fmt.Sprintf("- %s", n))
			}

			content := strings.Join(lines, "\n") + "\n"
			if outputPath != "" {
				if err := writeRouteOutputFile(outputPath, []byte(content)); err != nil {
					return err
				}
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprint(out, content)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write handoff report to file (strict safety policy)")
	return cmd
}

func validateRouteOutputPath(repoRoot, outputPath string) error {
	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repo root: %w", err)
	}

	assetsRoot := filepath.Join(absRepoRoot, "assets")
	if isSameOrWithin(absOutput, assetsRoot) {
		return fmt.Errorf("--output cannot write under assets/: %s", outputPath)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve user home directory: %w", err)
	}
	opencodeManaged := filepath.Join(homeDir, ".config", "opencode", "bitsentry")
	if isSameOrWithin(absOutput, opencodeManaged) {
		return fmt.Errorf("--output cannot write under OpenCode managed area: %s", outputPath)
	}

	opencodeExports := filepath.Join(homeDir, ".bitsentry-ai", "exports", "opencode-skills")
	if isSameOrWithin(absOutput, opencodeExports) {
		return fmt.Errorf("--output cannot write under OpenCode exports: %s", outputPath)
	}

	opencodeBackups := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode")
	if isSameOrWithin(absOutput, opencodeBackups) {
		return fmt.Errorf("--output cannot write under OpenCode backups: %s", outputPath)
	}

	opencodeSkillsBackups := filepath.Join(homeDir, ".bitsentry-ai", "backups", "opencode-skills")
	if isSameOrWithin(absOutput, opencodeSkillsBackups) {
		return fmt.Errorf("--output cannot write under OpenCode backups: %s", outputPath)
	}

	return nil
}

func newRouteReportCmd(rt *Runtime) *cobra.Command {
	var sessionID string
	var jsonOut bool
	var outputPath string

	cmd := &cobra.Command{
		Use:   "report --session <session-id>",
		Short: "Generate persisted route report without side effects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, session, sessionPath, err := loadRouteSession(sessionID)
			if err != nil {
				return err
			}
			root := store.Root()
			if strings.TrimSpace(outputPath) != "" {
				if err := validateRouteOutputPath(root, outputPath); err != nil {
					return err
				}
			}

			report, err := orchestrator.BuildRouteReport(session, sessionPath)
			if err != nil {
				return err
			}

			if jsonOut {
				raw, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route report output: %w", err)
				}
				if strings.TrimSpace(outputPath) != "" {
					if err := writeRouteOutputFile(outputPath, append(raw, '\n')); err != nil {
						return err
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Route report written to: %s\n", outputPath)
					return nil
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			markdown := orchestrator.RenderRouteReportMarkdown(report)
			if strings.TrimSpace(outputPath) != "" {
				if err := writeRouteOutputFile(outputPath, []byte(markdown)); err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Route report written to: %s\n", outputPath)
				return nil
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), markdown)
			return nil
		},
	}

	cmd.Flags().StringVar(&sessionID, "session", "", "Persisted session ID")
	_ = cmd.MarkFlagRequired("session")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&outputPath, "output", "", "Write route report to file (strict safety policy)")
	return cmd
}

func writeRouteOutputFile(outputPath string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := os.WriteFile(outputPath, content, 0o644); err != nil {
		return fmt.Errorf("write output file: %w", err)
	}
	return nil
}

func isSameOrWithin(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	return strings.HasPrefix(path, root+string(filepath.Separator))
}

func newRouteInspectCmd(rt *Runtime) *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect discovered flows and staged skills",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}

			catalog, err := capabilities.DiscoverAssets(root)
			if err != nil {
				return fmt.Errorf("discover assets: %w", err)
			}

			inspect := buildRouteInspectPayload(catalog.Flows)
			if jsonOut {
				raw, err := json.MarshalIndent(inspect, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal inspect output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			for _, flow := range inspect.Flows {
				_, _ = fmt.Fprintf(out, "Flow: %s\n", flow.Flow)
				_, _ = fmt.Fprintf(out, "Stages: %d\n", flow.StagesCount)
				if strings.TrimSpace(flow.InitialSkill) != "" {
					_, _ = fmt.Fprintf(out, "Initial skill: %s\n", flow.InitialSkill)
				}
				_, _ = fmt.Fprintln(out)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func newRoutePreviewCmd(rt *Runtime) *cobra.Command {
	var flowHint string
	var flowHintAlias string
	var jsonOut bool

	cmd := &cobra.Command{
		Use:   "preview [intent]",
		Short: "Preview flow routing plan without side effects",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			intent := strings.TrimSpace(args[0])
			if intent == "" {
				return fmt.Errorf("intent must not be empty")
			}

			hint := strings.TrimSpace(flowHint)
			if hint == "" {
				hint = strings.TrimSpace(flowHintAlias)
			}

			root, err := resolveRepoRootWithAssets(".")
			if err != nil {
				return fmt.Errorf("resolve assets root: %w", err)
			}

			o := orchestrator.New(root)
			result, err := o.Route(orchestrator.RouteRequest{Intent: orchestrator.Intent(intent), FlowHint: hint})
			if err != nil {
				return err
			}
			normalizedIntent := normalizeIntent(intent)
			sessionPreview := routeSessionPreviewOutput{
				WouldCreateSession: true,
				WouldPersist:       false,
				WouldExecute:       false,
			}

			if jsonOut {
				payload := routePreviewJSONOutput{
					Intent:       normalizedIntent,
					Flow:         result.Flow,
					InitialSkill: result.InitialSkill,
					Plan: routePlanJSONOutput{
						Flow:   result.Plan.FlowID,
						Stages: result.Plan.Stages,
					},
					Warnings:       append([]string{}, result.Warnings...),
					SessionPreview: sessionPreview,
				}
				raw, err := json.MarshalIndent(payload, "", "  ")
				if err != nil {
					return fmt.Errorf("marshal route output: %w", err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(raw))
				return nil
			}

			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "Intent: %s\n", normalizedIntent)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", result.Flow)
			if strings.TrimSpace(result.InitialSkill) != "" {
				_, _ = fmt.Fprintf(out, "Initial skill: %s\n", result.InitialSkill)
			}
			_, _ = fmt.Fprintln(out, "Plan stages:")
			for _, stage := range result.Plan.Stages {
				line := fmt.Sprintf("- %d. %s", stage.Order, firstNonEmpty(stage.Skill, stage.ID))
				if strings.TrimSpace(stage.Description) != "" {
					line += fmt.Sprintf(" — %s", stage.Description)
				}
				_, _ = fmt.Fprintln(out, line)
			}
			_, _ = fmt.Fprintln(out, "Session preview:")
			_, _ = fmt.Fprintf(out, "- would_create_session: %t\n", sessionPreview.WouldCreateSession)
			_, _ = fmt.Fprintf(out, "- would_persist: %t\n", sessionPreview.WouldPersist)
			_, _ = fmt.Fprintf(out, "- would_execute: %t\n", sessionPreview.WouldExecute)
			if len(result.Warnings) > 0 {
				_, _ = fmt.Fprintln(out, "Warnings:")
				for _, w := range result.Warnings {
					_, _ = fmt.Fprintf(out, "- %s\n", w)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flowHint, "flow-hint", "", "Force a specific flow ID (e.g., sdd, sdr, support)")
	cmd.Flags().StringVar(&flowHintAlias, "flow", "", "Alias of --flow-hint")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Emit JSON output")
	return cmd
}

func buildRouteInspectPayload(flows []capabilities.DiscoveredFlow) routeInspectJSONOutput {
	items := make([]routeInspectFlowJSONOutput, 0, len(flows))
	for _, flow := range flows {
		stages := make([]routePlanStageJSONOutput, 0, len(flow.Stages))
		skillSet := map[string]bool{}
		for i, rawStage := range flow.Stages {
			stage := routePlanStageJSONOutput{
				Order:       i + 1,
				ID:          stageValueString(rawStage["id"]),
				Skill:       stageValueString(rawStage["skill"]),
				Description: stageValueString(rawStage["description"]),
			}
			if stage.Skill != "" {
				skillSet[stage.Skill] = true
			}
			stages = append(stages, stage)
		}
		skills := sortedSkillList(skillSet)
		item := routeInspectFlowJSONOutput{
			Flow:         strings.TrimSpace(flow.ID),
			StagesCount:  len(stages),
			InitialSkill: initialSkill(stages),
			Stages:       stages,
			Skills:       skills,
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Flow < items[j].Flow })
	return routeInspectJSONOutput{Flows: items}
}

func initialSkill(stages []routePlanStageJSONOutput) string {
	for _, stage := range stages {
		if strings.TrimSpace(stage.Skill) != "" {
			return stage.Skill
		}
	}
	return ""
}

func sortedSkillList(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normalizeIntent(intent string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(intent))), " ")
}

func stageValueString(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func firstSuggestedStage(stages []orchestrator.ExecutionStage) (*routePlanStageJSONOutput, *routePlanStageJSONOutput) {
	if len(stages) == 0 {
		return nil, nil
	}
	stage := toRoutePlanStage(stages[0])
	return &stage, &stage
}

func suggestedStagesFromProgress(session orchestrator.Session) (*routePlanStageJSONOutput, *routePlanStageJSONOutput) {
	progress := normalizedProgress(session)
	if stage, ok := stageByID(session.Plan.Stages, progress.CurrentStageID); ok {
		current := toRoutePlanStage(stage)
		return &current, &current
	}
	for _, stage := range session.Plan.Stages {
		if progressStateForStage(session, strings.TrimSpace(stage.ID)) == "pending" {
			next := toRoutePlanStage(stage)
			return &next, &next
		}
	}
	return nil, nil
}

func loadRouteSession(sessionID string) (orchestrator.LocalSessionStore, orchestrator.Session, string, error) {
	root, err := resolveRepoRootWithAssets(".")
	if err != nil {
		return orchestrator.LocalSessionStore{}, orchestrator.Session{}, "", fmt.Errorf("resolve assets root: %w", err)
	}
	store := orchestrator.NewLocalSessionStore(root)
	session, sessionPath, err := store.LoadSession(sessionID)
	if err != nil {
		return orchestrator.LocalSessionStore{}, orchestrator.Session{}, "", err
	}
	return store, session, sessionPath, nil
}

func normalizedProgress(session orchestrator.Session) orchestrator.SessionProgress {
	now := session.UpdatedAt
	if now.IsZero() {
		now = session.CreatedAt
	}
	if session.Progress == nil {
		progress := orchestrator.SessionProgress{CompletedStageIDs: []string{}, UpdatedAt: now}
		if len(session.Plan.Stages) > 0 {
			progress.CurrentStageID = strings.TrimSpace(session.Plan.Stages[0].ID)
		}
		return progress
	}
	progress := *session.Progress
	if progress.CompletedStageIDs == nil {
		progress.CompletedStageIDs = []string{}
	}
	if progress.UpdatedAt.IsZero() {
		progress.UpdatedAt = now
	}
	return progress
}

func stageByID(stages []orchestrator.ExecutionStage, stageID string) (orchestrator.ExecutionStage, bool) {
	needle := strings.TrimSpace(stageID)
	for _, stage := range stages {
		if strings.TrimSpace(stage.ID) == needle {
			return stage, true
		}
	}
	return orchestrator.ExecutionStage{}, false
}

func applyStageMark(session orchestrator.Session, stageID string, targetState string) (orchestrator.Session, error) {
	trimmed := strings.TrimSpace(stageID)
	if trimmed == "" {
		return session, fmt.Errorf("stage id is required")
	}
	if _, ok := stageByID(session.Plan.Stages, trimmed); !ok {
		return session, fmt.Errorf("unknown stage id %q", trimmed)
	}
	progress := normalizedProgress(session)
	completed := map[string]bool{}
	for _, id := range progress.CompletedStageIDs {
		completed[strings.TrimSpace(id)] = true
	}

	switch targetState {
	case "current":
		delete(completed, trimmed)
		progress.CurrentStageID = trimmed
	case "done":
		completed[trimmed] = true
		if progress.CurrentStageID == trimmed {
			progress.CurrentStageID = ""
		}
		for _, stage := range session.Plan.Stages {
			id := strings.TrimSpace(stage.ID)
			if !completed[id] {
				progress.CurrentStageID = id
				break
			}
		}
	default:
		return session, fmt.Errorf("unsupported target state %q", targetState)
	}

	progress.CompletedStageIDs = make([]string, 0, len(session.Plan.Stages))
	for _, stage := range session.Plan.Stages {
		id := strings.TrimSpace(stage.ID)
		if completed[id] {
			progress.CompletedStageIDs = append(progress.CompletedStageIDs, id)
		}
	}
	progress.UpdatedAt = time.Now().UTC()

	session.Progress = &progress
	session.UpdatedAt = progress.UpdatedAt
	return session, nil
}

func progressStateForStage(session orchestrator.Session, stageID string) string {
	progress := normalizedProgress(session)
	id := strings.TrimSpace(stageID)
	for _, doneID := range progress.CompletedStageIDs {
		if strings.TrimSpace(doneID) == id {
			return "done"
		}
	}
	if strings.TrimSpace(progress.CurrentStageID) == id {
		return "current"
	}
	return "pending"
}

func buildRouteProgressPayload(session orchestrator.Session) routeProgressJSONOutput {
	progress := normalizedProgress(session)
	stages := make([]routeProgressStageJSONOutput, 0, len(session.Plan.Stages))
	for _, stage := range session.Plan.Stages {
		id := strings.TrimSpace(stage.ID)
		order := stage.Order
		if order <= 0 {
			order = len(stages) + 1
		}
		stages = append(stages, routeProgressStageJSONOutput{Order: order, ID: id, Skill: strings.TrimSpace(stage.Skill), State: progressStateForStage(session, id)})
	}
	current, _ := suggestedStagesFromProgress(session)
	return routeProgressJSONOutput{
		SessionID:    session.ID,
		Flow:         session.Flow,
		Status:       session.Status,
		CurrentStage: current,
		Stages:       stages,
		UpdatedAt:    progress.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		WouldExecute: false,
	}
}

func toRoutePlanStage(stage orchestrator.ExecutionStage) routePlanStageJSONOutput {
	order := stage.Order
	if order <= 0 {
		order = 1
	}
	return routePlanStageJSONOutput{
		Order:       order,
		ID:          strings.TrimSpace(stage.ID),
		Skill:       strings.TrimSpace(stage.Skill),
		Description: strings.TrimSpace(stage.Description),
	}
}

func renderSuggestedStage(stage *routePlanStageJSONOutput) string {
	if stage == nil {
		return "<none>"
	}
	return fmt.Sprintf("%d/%s (%s)", stage.Order, firstNonEmpty(stage.ID, "<unset>"), firstNonEmpty(stage.Skill, "<unset>"))
}

func listPersistedSessions(store orchestrator.LocalSessionStore, flow string, status string, archivedOnly bool, all bool, warnOut io.Writer) ([]orchestrator.Session, error) {
	sessionsRoot := filepath.Join(store.Root(), ".bitsentry-ai", "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return []orchestrator.Session{}, nil
		}
		return nil, fmt.Errorf("read sessions directory: %w", err)
	}
	out := make([]orchestrator.Session, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionID := strings.TrimSpace(entry.Name())
		session, _, loadErr := store.LoadSession(sessionID)
		if loadErr != nil {
			if warnOut != nil {
				_, _ = fmt.Fprintf(warnOut, "warning: skipping session %q: %v\n", sessionID, loadErr)
			}
			continue
		}
		if !matchesArchiveFilter(session, archivedOnly, all) {
			continue
		}
		if flow != "" && strings.TrimSpace(session.Flow) != flow {
			continue
		}
		if status != "" && strings.TrimSpace(session.Status) != status {
			continue
		}
		out = append(out, session)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func buildRouteListPayload(root string, sessions []orchestrator.Session) routeListJSONOutput {
	items := make([]routeListItemJSONOutput, 0, len(sessions))
	for _, s := range sessions {
		progress := normalizedProgress(s)
		total := len(s.Plan.Stages)
		completed := len(progress.CompletedStageIDs)
		items = append(items, routeListItemJSONOutput{
			SessionID:      s.ID,
			Flow:           strings.TrimSpace(s.Flow),
			Status:         strings.TrimSpace(s.Status),
			Archived:       s.Archived,
			Intent:         summarizeIntent(s.Intent),
			CurrentStage:   firstNonEmpty(strings.TrimSpace(progress.CurrentStageID), "<none>"),
			CompletedTotal: fmt.Sprintf("%d/%d", completed, total),
			UpdatedAt:      s.UpdatedAt.Format(time.RFC3339),
			Path:           filepath.Join(root, ".bitsentry-ai", "sessions", s.ID),
		})
	}
	return routeListJSONOutput{Sessions: items, Total: len(items), WouldExecute: false}
}

func summarizeIntent(intent string) string {
	trimmed := strings.TrimSpace(intent)
	if len(trimmed) <= 80 {
		return trimmed
	}
	return strings.TrimSpace(trimmed[:77]) + "..."
}

func matchesArchiveFilter(session orchestrator.Session, archivedOnly bool, all bool) bool {
	if all {
		return true
	}
	if archivedOnly {
		return session.Archived
	}
	return !session.Archived
}

func formatOptionalTime(ts *time.Time) string {
	if ts == nil || ts.IsZero() {
		return ""
	}
	return ts.UTC().Format(time.RFC3339)
}

func validateSingleSession(sessionID string) (routeValidateJSONOutput, error) {
	store, session, sessionPath, err := loadRouteSession(sessionID)
	if err != nil {
		return routeValidateJSONOutput{}, err
	}
	_ = store
	return validateLoadedSession(session, sessionID, sessionPath), nil
}

func runRouteRepair(sessionID string, dryRun bool, apply bool) (routeRepairJSONOutput, error) {
	store, session, sessionPath, err := loadRouteSession(sessionID)
	if err != nil {
		return routeRepairJSONOutput{}, err
	}

	result := routeRepairJSONOutput{
		SessionID:    strings.TrimSpace(sessionID),
		DryRun:       dryRun,
		Applied:      false,
		Repairable:   true,
		Changes:      []string{},
		Unrepairable: []string{},
		Warnings:     []string{},
		WouldExecute: false,
	}

	if session.Plan.Stages == nil {
		result.Repairable = false
		result.Unrepairable = append(result.Unrepairable, "plan missing: unable to repair without plan stages")
		return result, nil
	}

	mutated := session
	now := time.Now().UTC()

	if strings.TrimSpace(mutated.Plan.FlowID) == "" {
		if strings.TrimSpace(mutated.Flow) != "" {
			result.Changes = append(result.Changes, "set plan.flow_id from session.flow")
			if apply {
				mutated.Plan.FlowID = strings.TrimSpace(mutated.Flow)
			}
		} else {
			result.Unrepairable = append(result.Unrepairable, "plan flow missing and session.flow empty")
			result.Repairable = false
		}
	}

	progress := normalizedProgress(mutated)
	if mutated.Progress == nil {
		result.Changes = append(result.Changes, "initialize missing progress")
		if apply {
			mutated.Progress = &progress
		}
	}

	planStageIDs := map[string]bool{}
	for _, stage := range mutated.Plan.Stages {
		id := strings.TrimSpace(stage.ID)
		if id != "" {
			planStageIDs[id] = true
		}
	}

	if len(progress.CompletedStageIDs) > 0 {
		deduped := make([]string, 0, len(progress.CompletedStageIDs))
		seen := map[string]bool{}
		hadDuplicates := false
		hadUnknown := false
		for _, raw := range progress.CompletedStageIDs {
			id := strings.TrimSpace(raw)
			if id == "" || !planStageIDs[id] {
				hadUnknown = true
				continue
			}
			if seen[id] {
				hadDuplicates = true
				continue
			}
			seen[id] = true
			deduped = append(deduped, id)
		}
		if hadDuplicates {
			result.Changes = append(result.Changes, "deduplicate progress.completed_stage_ids")
		}
		if hadUnknown {
			result.Changes = append(result.Changes, "remove progress.completed_stage_ids not present in plan")
		}
		if apply && (hadDuplicates || hadUnknown) {
			progress.CompletedStageIDs = deduped
		}
	}

	if current := strings.TrimSpace(progress.CurrentStageID); current != "" && !planStageIDs[current] {
		result.Changes = append(result.Changes, "clear progress.current_stage_id not present in plan")
		if apply {
			progress.CurrentStageID = ""
		}
	}

	if apply {
		timestampsFixed := false
		if mutated.CreatedAt.IsZero() {
			mutated.CreatedAt = now
			timestampsFixed = true
		}
		if progress.UpdatedAt.IsZero() {
			progress.UpdatedAt = now
			timestampsFixed = true
		}
		if mutated.UpdatedAt.IsZero() {
			if !progress.UpdatedAt.IsZero() {
				mutated.UpdatedAt = progress.UpdatedAt
			} else {
				mutated.UpdatedAt = now
			}
			timestampsFixed = true
		}
		if timestampsFixed {
			result.Changes = append(result.Changes, "repair empty timestamps")
		}
		if mutated.Archived && (mutated.ArchivedAt == nil || mutated.ArchivedAt.IsZero()) {
			result.Changes = append(result.Changes, "set archived_at for archived session")
			ts := now
			mutated.ArchivedAt = &ts
		}
	}

	briefPath := filepath.Join(sessionPath, "brief.md")
	handoffPath := filepath.Join(sessionPath, "handoff.md")
	_, briefErr := os.Stat(briefPath)
	_, handoffErr := os.Stat(handoffPath)
	if os.IsNotExist(briefErr) {
		result.Changes = append(result.Changes, "create missing brief.md from session.json")
	}
	if os.IsNotExist(handoffErr) {
		result.Changes = append(result.Changes, "create missing handoff.md from session.json")
	}

	if apply {
		progress.UpdatedAt = now
		mutated.Progress = &progress
		mutated.UpdatedAt = now
		if _, err := store.SaveSessionMetadata(mutated); err != nil {
			return routeRepairJSONOutput{}, err
		}
		if os.IsNotExist(briefErr) {
			if err := os.WriteFile(briefPath, []byte(orchestrator.RenderBriefMarkdown(mutated.Brief)), 0o644); err != nil {
				return routeRepairJSONOutput{}, fmt.Errorf("write brief.md: %w", err)
			}
		}
		if os.IsNotExist(handoffErr) {
			if err := os.WriteFile(handoffPath, []byte(orchestrator.RenderHandoffMarkdown(mutated.Handoff)), 0o644); err != nil {
				return routeRepairJSONOutput{}, fmt.Errorf("write handoff.md: %w", err)
			}
		}
		result.Applied = true
	}

	if len(result.Unrepairable) > 0 {
		result.Repairable = false
	}
	return result, nil
}

func runRouteMigrate(sessionID string, dryRun bool, apply bool) (routeMigrateJSONOutput, error) {
	store, session, _, err := loadRouteSession(sessionID)
	if err != nil {
		return routeMigrateJSONOutput{}, err
	}
	fromVersion := strings.TrimSpace(session.SchemaVersion)
	toVersion := orchestrator.SessionSchemaVersionV1
	result := routeMigrateJSONOutput{
		SessionID:    strings.TrimSpace(sessionID),
		DryRun:       dryRun,
		Applied:      false,
		FromVersion:  fromVersion,
		ToVersion:    toVersion,
		Changes:      []string{},
		Warnings:     []string{},
		WouldExecute: false,
	}

	if fromVersion == "" {
		result.Changes = append(result.Changes, fmt.Sprintf("set schema_version to %s", toVersion))
		result.Warnings = append(result.Warnings, "legacy session without schema_version")
	} else if fromVersion != toVersion {
		result.Changes = append(result.Changes, fmt.Sprintf("set schema_version from %s to %s", fromVersion, toVersion))
	} else {
		result.Changes = append(result.Changes, "schema_version already current")
	}

	if apply {
		if fromVersion != toVersion {
			now := time.Now().UTC()
			session.SchemaVersion = toVersion
			session.UpdatedAt = now
			if session.Progress != nil {
				session.Progress.UpdatedAt = now
			}
			if _, err := store.SaveSessionMetadata(session); err != nil {
				return routeMigrateJSONOutput{}, err
			}
		}
		result.Applied = true
	}

	return result, nil
}

func validateLoadedSession(session orchestrator.Session, expectedSessionID string, sessionPath string) routeValidateJSONOutput {
	errors := []string{}
	warnings := []string{}
	checks := []routeValidationCheckJSONOutput{}

	addCheck := func(name string, ok bool, warning bool, msg string) {
		status := "pass"
		if !ok {
			status = "fail"
			if warning {
				warnings = append(warnings, msg)
			} else {
				errors = append(errors, msg)
			}
		}
		checks = append(checks, routeValidationCheckJSONOutput{Name: name, Status: status})
	}

	sessionFile := filepath.Join(sessionPath, "session.json")
	_, statErr := os.Stat(sessionFile)
	addCheck("session.json exists", statErr == nil, false, "missing session.json")

	addCheck("path id matches session.id", strings.TrimSpace(expectedSessionID) == strings.TrimSpace(session.ID), false, fmt.Sprintf("session id mismatch: path=%q session.id=%q", expectedSessionID, session.ID))
	addCheck("intent/flow/status non-empty", strings.TrimSpace(session.Intent) != "" && strings.TrimSpace(session.Flow) != "" && strings.TrimSpace(session.Status) != "", false, "intent, flow and status must be non-empty")
	addCheck("initial_skill required when stages exist", len(session.Plan.Stages) == 0 || strings.TrimSpace(session.InitialSkill) != "", false, "initial_skill must be non-empty when plan has stages")
	addCheck("plan exists", session.Plan.Stages != nil, false, "plan must exist")
	addCheck("plan flow matches session flow", strings.TrimSpace(session.Plan.FlowID) == strings.TrimSpace(session.Flow), false, fmt.Sprintf("plan flow mismatch: plan.flow=%q session.flow=%q", session.Plan.FlowID, session.Flow))
	currentSchema := strings.TrimSpace(session.SchemaVersion)
	addCheck("schema_version present", currentSchema != "", true, "schema_version missing (legacy session)")
	addCheck("schema_version supported", currentSchema == "" || currentSchema == orchestrator.SessionSchemaVersionV1, false, fmt.Sprintf("unsupported schema_version %q", currentSchema))

	stageIDs := map[string]bool{}
	duplicateStageIDs := map[string]bool{}
	stagesConsistent := true
	for idx, stage := range session.Plan.Stages {
		id := strings.TrimSpace(stage.ID)
		skill := strings.TrimSpace(stage.Skill)
		if stage.Order <= 0 || id == "" || skill == "" {
			stagesConsistent = false
		}
		if stage.Order != idx+1 {
			stagesConsistent = false
		}
		if id != "" {
			if stageIDs[id] {
				duplicateStageIDs[id] = true
			}
			stageIDs[id] = true
		}
	}
	addCheck("stages have consistent order/id/skill", stagesConsistent, false, "stages must have sequential order and non-empty id/skill")
	addCheck("no duplicate stage IDs", len(duplicateStageIDs) == 0, false, fmt.Sprintf("duplicate stage ids: %v", mapKeys(duplicateStageIDs)))

	progress := normalizedProgress(session)
	currentStageID := strings.TrimSpace(progress.CurrentStageID)
	addCheck("progress.current_stage_id exists in plan", currentStageID == "" || stageIDs[currentStageID], false, fmt.Sprintf("current stage %q not present in plan", currentStageID))

	seenCompleted := map[string]bool{}
	completedInPlan := true
	completedUnique := true
	for _, completedIDRaw := range progress.CompletedStageIDs {
		completedID := strings.TrimSpace(completedIDRaw)
		if completedID == "" || !stageIDs[completedID] {
			completedInPlan = false
		}
		if seenCompleted[completedID] {
			completedUnique = false
		}
		seenCompleted[completedID] = true
	}
	addCheck("progress.completed_stage_ids exist in plan", completedInPlan, false, "completed stage ids must exist in plan")
	addCheck("no duplicate completed_stage_ids", completedUnique, false, "duplicate completed_stage_ids detected")

	archivedCoherent := true
	if session.Archived && (session.ArchivedAt == nil || session.ArchivedAt.IsZero()) {
		archivedCoherent = false
	}
	if !session.Archived && session.ArchivedAt != nil && !session.ArchivedAt.IsZero() {
		archivedCoherent = false
	}
	addCheck("archived/archived_at coherence", archivedCoherent, false, "archived and archived_at are incoherent")

	restoredCoherent := session.RestoredAt == nil || !session.RestoredAt.IsZero()
	if session.RestoredAt != nil && session.Archived {
		restoredCoherent = false
	}
	addCheck("restored_at coherence", restoredCoherent, false, "restored_at is incoherent")

	briefPath := filepath.Join(sessionPath, "brief.md")
	handoffPath := filepath.Join(sessionPath, "handoff.md")
	_, briefErr := os.Stat(briefPath)
	_, handoffErr := os.Stat(handoffPath)
	addCheck("brief and handoff files present", briefErr == nil && handoffErr == nil, true, "brief.md or handoff.md missing")

	handoffTo := strings.TrimSpace(session.Handoff.To)
	initialSkill := strings.TrimSpace(session.InitialSkill)
	addCheck("handoff.to coherent with initial_skill", initialSkill == "" || handoffTo == "" || handoffTo == initialSkill, false, fmt.Sprintf("handoff.to=%q initial_skill=%q mismatch", handoffTo, initialSkill))

	return routeValidateJSONOutput{
		SessionID:    strings.TrimSpace(expectedSessionID),
		Path:         sessionPath,
		Valid:        len(errors) == 0,
		Errors:       errors,
		Warnings:     warnings,
		Checks:       checks,
		WouldExecute: false,
	}
}

func auditSessions(includeArchived bool) (routeAuditJSONOutput, error) {
	root, err := resolveRepoRootWithAssets(".")
	if err != nil {
		return routeAuditJSONOutput{}, fmt.Errorf("resolve assets root: %w", err)
	}
	store := orchestrator.NewLocalSessionStore(root)
	sessionsRoot := filepath.Join(store.Root(), ".bitsentry-ai", "sessions")
	entries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return routeAuditJSONOutput{Sessions: []routeValidateJSONOutput{}, IncludeArchived: includeArchived, WouldExecute: false}, nil
		}
		return routeAuditJSONOutput{}, fmt.Errorf("read sessions directory: %w", err)
	}

	results := make([]routeValidateJSONOutput, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		sessionID := strings.TrimSpace(entry.Name())
		sessionPath := filepath.Join(sessionsRoot, sessionID)
		session, _, loadErr := store.LoadSession(sessionID)
		if loadErr != nil {
			results = append(results, routeValidateJSONOutput{SessionID: sessionID, Path: sessionPath, Valid: false, Errors: []string{loadErr.Error()}, Warnings: []string{}, Checks: []routeValidationCheckJSONOutput{{Name: "session.json parse/load", Status: "fail"}}, WouldExecute: false})
			continue
		}
		if session.Archived && !includeArchived {
			continue
		}
		results = append(results, validateLoadedSession(session, sessionID, sessionPath))
	}

	sort.SliceStable(results, func(i, j int) bool { return results[i].SessionID < results[j].SessionID })
	passed := 0
	failed := 0
	warnings := 0
	for _, r := range results {
		if r.Valid {
			passed++
		} else {
			failed++
		}
		if len(r.Warnings) > 0 {
			warnings++
		}
	}

	return routeAuditJSONOutput{Sessions: results, Total: len(results), Passed: passed, Failed: failed, Warnings: warnings, IncludeArchived: includeArchived, WouldExecute: false}, nil
}

func mapKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func resolveRepoRootWithAssets(start string) (string, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}

	current := absStart
	for {
		assetsDir := filepath.Join(current, "assets")
		if st, err := os.Stat(assetsDir); err == nil && st.IsDir() {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	return "", fmt.Errorf("assets directory not found from root: %s", absStart)
}
