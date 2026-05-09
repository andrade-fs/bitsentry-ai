package orchestrator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const SessionStatusPlanned = "planned"
const SessionSchemaVersionV1 = "route-session/v1"

type Session struct {
	ID           string        `json:"id"`
	Intent       string        `json:"intent"`
	Flow         string        `json:"flow"`
	InitialSkill string        `json:"initial_skill"`
	Status       string        `json:"status"`
	SchemaVersion string       `json:"schema_version,omitempty"`
	Archived     bool          `json:"archived,omitempty"`
	ArchivedAt   *time.Time    `json:"archived_at,omitempty"`
	RestoredAt   *time.Time    `json:"restored_at,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Plan         ExecutionPlan `json:"plan"`
	Progress     *SessionProgress `json:"progress,omitempty"`
	Brief        Brief         `json:"brief"`
	Handoff      Handoff       `json:"handoff"`
}

type SessionProgress struct {
	CurrentStageID    string   `json:"current_stage_id,omitempty"`
	CompletedStageIDs []string `json:"completed_stage_ids,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Brief struct {
	IntentOriginal   string   `json:"intent_original"`
	IntentNormalized string   `json:"intent_normalized"`
	Flow             string   `json:"flow"`
	Constraints      []string `json:"constraints"`
	NonGoals         []string `json:"non_goals"`
}

type Handoff struct {
	SessionID   string   `json:"session_id"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Reason      string   `json:"reason"`
	NextSteps   []string `json:"next_steps"`
	SafetyNotes []string `json:"safety_notes"`
}

func NewSession(intentOriginal string, route RouteResult) Session {
	now := time.Now().UTC()
	normalized := normalizeIntentSession(intentOriginal)
	flow := strings.TrimSpace(route.Flow)
	initialSkill := strings.TrimSpace(route.InitialSkill)
	if initialSkill == "" && len(route.Plan.Stages) > 0 {
		initialSkill = strings.TrimSpace(route.Plan.Stages[0].Skill)
	}

	brief := Brief{
		IntentOriginal:   strings.TrimSpace(intentOriginal),
		IntentNormalized: normalized,
		Flow:             flow,
		Constraints:      []string{},
		NonGoals: []string{
			"No runtime autónomo",
			"No ejecución de agentes externos",
			"No side effects fuera del storage local",
		},
	}

	handoff := Handoff{
		From:   "route",
		To:     initialSkill,
		Reason: fmt.Sprintf("Intent routed to %s flow", flow),
		NextSteps: []string{
			fmt.Sprintf("Run skill %s for implementation", fallback(initialSkill, "<unset>")),
			"Execute tasks in plan order",
			"Keep all writes scoped to local workspace storage",
		},
		SafetyNotes: []string{
			"Do not mutate OpenCode managed areas",
			"Do not call external agents",
			"Persist only in .bitsentry-ai/sessions",
		},
	}

	id := newSessionID(normalized, flow, now)
	handoff.SessionID = id

	return Session{
		ID:           id,
		Intent:       normalized,
		Flow:         flow,
		InitialSkill: initialSkill,
		Status:       SessionStatusPlanned,
		SchemaVersion: SessionSchemaVersionV1,
		CreatedAt:    now,
		UpdatedAt:    now,
		Plan:         route.Plan,
		Progress:     defaultSessionProgress(route.Plan.Stages, now),
		Brief:        brief,
		Handoff:      handoff,
	}
}

func defaultSessionProgress(stages []ExecutionStage, now time.Time) *SessionProgress {
	progress := &SessionProgress{UpdatedAt: now.UTC(), CompletedStageIDs: []string{}}
	if len(stages) == 0 {
		return progress
	}
	progress.CurrentStageID = strings.TrimSpace(stages[0].ID)
	return progress
}

func newSessionID(intent, flow string, now time.Time) string {
	seed := fmt.Sprintf("%s|%s|%d", intent, flow, now.UnixNano())
	sum := sha256.Sum256([]byte(seed))
	short := hex.EncodeToString(sum[:])[:8]
	return fmt.Sprintf("%s-%s", now.UTC().Format("20060102T150405Z"), short)
}

func normalizeIntentSession(intent string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(intent))), " ")
}

func fallback(v string, alt string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return alt
}
