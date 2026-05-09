package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestRouteCommandPrintsHumanPlan(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"preview", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route command: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Flow: sdd") {
		t.Fatalf("expected sdd flow in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Plan stages:") {
		t.Fatalf("expected plan stages in output, got:\n%s", got)
	}
}

func TestRouteStartSetsSessionSchemaVersion(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	var payload struct {
		SchemaVersion string `json:"schema_version"`
	}
	raw := mustReadSessionFile(t, repo, sessionID)
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("parse session json: %v", err)
	}
	if payload.SchemaVersion != "route-session/v1" {
		t.Fatalf("expected schema_version route-session/v1, got %q", payload.SchemaVersion)
	}
}

func TestRouteValidateChecksSchemaVersion(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		delete(payload, "schema_version")
	})

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate --json: %v", err)
	}

	var result struct {
		Warnings []string `json:"warnings"`
		Checks   []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatalf("parse route validate json: %v", err)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected schema_version warning for legacy session")
	}
	foundCheck := false
	for _, check := range result.Checks {
		if check.Name == "schema_version present" {
			foundCheck = true
			if check.Status != "fail" {
				t.Fatalf("expected schema_version present check fail for legacy session, got %q", check.Status)
			}
		}
	}
	if !foundCheck {
		t.Fatalf("expected schema_version present check in validate output")
	}
}

func TestRouteReportIncludesSchemaVersion(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report: %v", err)
	}
	if !strings.Contains(out.String(), "Schema version: route-session/v1") {
		t.Fatalf("expected schema_version in markdown report, got:\n%s", out.String())
	}

	cmdJSON := newRouteCmd(&Runtime{})
	outJSON := &bytes.Buffer{}
	cmdJSON.SetOut(outJSON)
	cmdJSON.SetErr(outJSON)
	cmdJSON.SetArgs([]string{"report", "--session", sessionID, "--json"})
	if err := cmdJSON.Execute(); err != nil {
		t.Fatalf("execute route report --json: %v", err)
	}
	var payload struct {
		Session struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"session"`
		Lifecycle struct {
			SchemaVersion string `json:"schema_version"`
		} `json:"lifecycle"`
	}
	if err := json.Unmarshal(outJSON.Bytes(), &payload); err != nil {
		t.Fatalf("parse report json: %v", err)
	}
	if payload.Session.SchemaVersion != "route-session/v1" || payload.Lifecycle.SchemaVersion != "route-session/v1" {
		t.Fatalf("expected schema_version in report json, got session=%q lifecycle=%q", payload.Session.SchemaVersion, payload.Lifecycle.SchemaVersion)
	}
}

func TestRouteMigrateDryRunIsReadOnly(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	before := mustReadSessionFile(t, repo, sessionID)
	runRouteMigrateCmd(t, []string{"--session", sessionID, "--dry-run"})
	after := mustReadSessionFile(t, repo, sessionID)
	if before != after {
		t.Fatalf("session.json changed after migrate dry-run")
	}
}

func TestRouteMigrateDryRunSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := runRouteMigrateCmd(t, []string{"--session", sessionID, "--dry-run", "--json"})
	var payload struct {
		SessionID    string   `json:"session_id"`
		DryRun       bool     `json:"dry_run"`
		Applied      bool     `json:"applied"`
		FromVersion  string   `json:"from_version"`
		ToVersion    string   `json:"to_version"`
		Changes      []string `json:"changes"`
		WouldExecute bool     `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route migrate json invalid: %v", err)
	}
	if payload.SessionID != sessionID || !payload.DryRun || payload.Applied || payload.ToVersion != "route-session/v1" || payload.WouldExecute {
		t.Fatalf("unexpected route migrate dry-run payload: %+v", payload)
	}
}

func TestRouteMigrateApplyRequiresConfirm(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"migrate", "--session", sessionID, "--apply"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("expected migrate apply requires confirm error, got %v", err)
	}
}

func TestRouteMigrateRequiresMode(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"migrate", "--session", sessionID})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires one mode") {
		t.Fatalf("expected migrate mode error, got %v", err)
	}
}

func TestRouteMigrateApplySetsSchemaVersion(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		delete(payload, "schema_version")
	})
	runRouteMigrateCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	raw := mustReadSessionFile(t, repo, sessionID)
	var payload struct {
		SchemaVersion string `json:"schema_version"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("parse session json: %v", err)
	}
	if payload.SchemaVersion != "route-session/v1" {
		t.Fatalf("expected migrated schema_version route-session/v1, got %q", payload.SchemaVersion)
	}
}

func TestRouteMigrateRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"migrate", "--session", "../evil", "--dry-run"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteMigrateMissingSessionReturnsSafeError(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"migrate", "--session", "20260101T000000Z-missing", "--dry-run"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing session safe error, got %v", err)
	}
}

func TestRouteMigrateDoesNotMutateOpenCode(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	runRouteMigrateCmd(t, []string{"--session", sessionID, "--dry-run"})
	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route migrate not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteMigrateDoesNotTouchAssets(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	assetFile := filepath.Join(repo, "assets", "_route_migrate_sentinel.txt")
	if err := os.WriteFile(assetFile, []byte("sentinel"), 0o644); err != nil {
		t.Fatalf("write sentinel asset: %v", err)
	}
	before, err := os.ReadFile(assetFile)
	if err != nil {
		t.Fatalf("read sentinel asset before: %v", err)
	}
	runRouteMigrateCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	after, err := os.ReadFile(assetFile)
	if err != nil {
		t.Fatalf("read sentinel asset after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("assets file mutated by route migrate")
	}
}

func TestRouteMigrateDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := runRouteMigrateCmd(t, []string{"--session", sessionID, "--dry-run", "--json"})
	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("parse migrate json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected would_execute=false for route migrate")
	}
}

func TestRouteCommandSupportsFlowHint(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "flow-hint", args: []string{"preview", "--flow-hint", "sdr", "design feature change"}},
		{name: "flow alias", args: []string{"preview", "--flow", "sdr", "design feature change"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRouteCmd(&Runtime{})
			out := &bytes.Buffer{}
			cmd.SetOut(out)
			cmd.SetErr(out)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("execute route command: %v", err)
			}

			if !strings.Contains(out.String(), "Flow: sdr") {
				t.Fatalf("expected hinted flow sdr, got:\n%s", out.String())
			}
		})
	}
}

func TestRouteCommandSupportsJSON(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"preview", "--json", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route command: %v", err)
	}

	var payload struct {
		Intent       string `json:"intent"`
		Flow         string `json:"flow"`
		InitialSkill string `json:"initial_skill"`
		Plan         struct {
			Flow   string `json:"flow"`
			Stages []any  `json:"stages"`
		} `json:"plan"`
		Warnings []string `json:"warnings"`
		SessionPreview struct {
			WouldCreateSession bool `json:"would_create_session"`
			WouldPersist       bool `json:"would_persist"`
			WouldExecute       bool `json:"would_execute"`
		} `json:"session_preview"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("output is not valid json: %v\noutput:\n%s", err, out.String())
	}
	if payload.Flow != "sdd" {
		t.Fatalf("expected flow sdd, got %q", payload.Flow)
	}
	if payload.Intent != "design feature change" {
		t.Fatalf("expected normalized intent, got %q", payload.Intent)
	}
	if payload.Plan.Flow != "sdd" {
		t.Fatalf("expected plan flow sdd, got %q", payload.Plan.Flow)
	}
	if len(payload.Plan.Stages) == 0 {
		t.Fatalf("expected at least one stage in plan")
	}
	if !payload.SessionPreview.WouldCreateSession || payload.SessionPreview.WouldPersist || payload.SessionPreview.WouldExecute {
		t.Fatalf("unexpected session preview semantics: %+v", payload.SessionPreview)
	}
}

func TestRouteCommandRejectsUnknownIntent(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "totally unrelated tokens"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected unknown intent error")
	}
	if !strings.Contains(err.Error(), "unable to resolve flow safely from intent") {
		t.Fatalf("expected safe unknown intent error, got %v", err)
	}
}

func TestRouteCommandRejectsInvalidFlowHint(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "--flow-hint", "invalid", "design feature change"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected invalid flow hint error")
	}
	if !strings.Contains(err.Error(), "invalid flow hint") {
		t.Fatalf("expected invalid flow hint error message, got %v", err)
	}
}

func TestRouteCommandDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "help troubleshoot error"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route command: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route command not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteInspectListsDiscoveredFlows(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"inspect"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route inspect: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Flow: sdd") || !strings.Contains(got, "Flow: sdr") || !strings.Contains(got, "Flow: support") {
		t.Fatalf("expected discovered flows in output, got:\n%s", got)
	}
	if !strings.Contains(got, "Stages:") {
		t.Fatalf("expected stage count in output, got:\n%s", got)
	}
}

func TestRouteInspectSupportsJSON(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"inspect", "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route inspect --json: %v", err)
	}

	var payload struct {
		Flows []struct {
			Flow        string   `json:"flow"`
			StagesCount int      `json:"stages_count"`
			Skills      []string `json:"skills"`
		} `json:"flows"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("inspect output is not valid json: %v\noutput:\n%s", err, out.String())
	}
	if len(payload.Flows) < 3 {
		t.Fatalf("expected at least 3 flows, got %d", len(payload.Flows))
	}
}

func TestRoutePreviewPrintsDrySessionPreview(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"preview", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route preview: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Session preview:") || !strings.Contains(got, "would_create_session: true") || !strings.Contains(got, "would_persist: false") || !strings.Contains(got, "would_execute: false") {
		t.Fatalf("expected dry session preview in output, got:\n%s", got)
	}
}

func TestRoutePreviewSupportsJSON(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"preview", "--json", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route preview --json: %v", err)
	}

	var payload struct {
		Intent string `json:"intent"`
		Flow   string `json:"flow"`
		SessionPreview struct {
			WouldCreateSession bool `json:"would_create_session"`
			WouldPersist       bool `json:"would_persist"`
			WouldExecute       bool `json:"would_execute"`
		} `json:"session_preview"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("preview output is not valid json: %v\noutput:\n%s", err, out.String())
	}
	if payload.Flow != "sdd" {
		t.Fatalf("expected flow sdd, got %q", payload.Flow)
	}
}

func TestRoutePreviewRejectsUnknownIntent(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "totally unrelated tokens"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected unknown intent error")
	}
	if !strings.Contains(err.Error(), "unable to resolve flow safely from intent") {
		t.Fatalf("expected safe unknown intent error, got %v", err)
	}
}

func TestRouteInspectDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"inspect"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route inspect: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route inspect not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRoutePreviewDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "help troubleshoot error"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route preview: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route preview not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRoutePreviewDoesNotPersistSession(t *testing.T) {
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"preview", "--json", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route preview --json: %v", err)
	}

	var payload struct {
		SessionPreview struct {
			WouldCreateSession bool `json:"would_create_session"`
			WouldPersist       bool `json:"would_persist"`
			WouldExecute       bool `json:"would_execute"`
		} `json:"session_preview"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("preview output is not valid json: %v", err)
	}
	if !payload.SessionPreview.WouldCreateSession || payload.SessionPreview.WouldPersist || payload.SessionPreview.WouldExecute {
		t.Fatalf("unexpected session persistence semantics: %+v", payload.SessionPreview)
	}
}

func TestRouteStartPersistsSession(t *testing.T) {
	repo := setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"start", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route start: %v", err)
	}

	sessionsDir := filepath.Join(repo, ".bitsentry-ai", "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("read sessions dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one persisted session, got %d", len(entries))
	}
	sessionDir := filepath.Join(sessionsDir, entries[0].Name())
	for _, rel := range []string{"session.json", "brief.md", "handoff.md"} {
		if _, err := os.Stat(filepath.Join(sessionDir, rel)); err != nil {
			t.Fatalf("expected persisted file %s: %v", rel, err)
		}
	}
}

func TestRouteStartPrintsSessionID(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"start", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route start: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Session ID:") {
		t.Fatalf("expected session id in output, got:\n%s", got)
	}
}

func TestRouteStartSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"start", "--json", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route start --json: %v", err)
	}

	var payload struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
		Flow      string `json:"flow"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route start json invalid: %v", err)
	}
	if payload.SessionID == "" || payload.Status != "planned" || payload.Flow != "sdd" {
		t.Fatalf("unexpected route start payload: %+v", payload)
	}
}

func TestRouteStatusReadsPersistedSession(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"status", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route status: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Status: planned") {
		t.Fatalf("expected planned status, got:\n%s", got)
	}
}

func TestRouteStatusSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"status", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route status --json: %v", err)
	}

	var payload struct {
		SessionID string `json:"session_id"`
		Status    string `json:"status"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route status json invalid: %v", err)
	}
	if payload.SessionID != sessionID || payload.Status != "planned" {
		t.Fatalf("unexpected route status payload: %+v", payload)
	}
}

func TestRouteResumeReadsPersistedSession(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"resume", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route resume: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Session ID: "+sessionID) || !strings.Contains(got, "Current stage:") || !strings.Contains(got, "Safety reminder:") {
		t.Fatalf("expected persisted session resume details, got:\n%s", got)
	}
}

func TestRouteResumeSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"resume", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route resume --json: %v", err)
	}

	var payload struct {
		SessionID    string `json:"session_id"`
		Flow         string `json:"flow"`
		WouldExecute bool   `json:"would_execute"`
		CurrentStage *struct {
			Order int    `json:"order"`
			ID    string `json:"id"`
			Skill string `json:"skill"`
		} `json:"current_stage"`
		NextStage *struct {
			Order int    `json:"order"`
			ID    string `json:"id"`
			Skill string `json:"skill"`
		} `json:"next_stage"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route resume json invalid: %v", err)
	}
	if payload.SessionID != sessionID || payload.Flow != "sdd" || payload.WouldExecute {
		t.Fatalf("unexpected route resume payload: %+v", payload)
	}
	if payload.CurrentStage == nil || payload.NextStage == nil {
		t.Fatalf("expected suggested stages in route resume payload")
	}
}

func TestRouteNextShowsRecommendedSkill(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"next", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route next: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Recommended skill: sdd-apply") || !strings.Contains(got, "Suggested commands:") {
		t.Fatalf("expected recommended next skill output, got:\n%s", got)
	}
}

func TestRouteNextSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"next", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route next --json: %v", err)
	}

	var payload struct {
		SessionID    string `json:"session_id"`
		NextSkill    string `json:"next_skill"`
		WouldExecute bool   `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route next json invalid: %v", err)
	}
	if payload.SessionID != sessionID || payload.NextSkill != "sdd-apply" || payload.WouldExecute {
		t.Fatalf("unexpected route next payload: %+v", payload)
	}
}

func TestRouteHandoffReadsPersistedHandoff(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"handoff", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route handoff: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "From: route") || !strings.Contains(got, "Next steps:") {
		t.Fatalf("expected handoff output, got:\n%s", got)
	}
}

func TestRouteHandoffSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"handoff", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route handoff --json: %v", err)
	}

	var payload struct {
		SessionID string `json:"session_id"`
		From      string `json:"from"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route handoff json invalid: %v", err)
	}
	if payload.SessionID != sessionID || payload.From != "route" {
		t.Fatalf("unexpected route handoff payload: %+v", payload)
	}
}

func TestRouteHandoffWritesOutputFileWithoutMutatingSession(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	repo, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	sessionPath := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("read session before: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "handoff.txt")
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"handoff", "--session", sessionID, "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route handoff --output: %v", err)
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	if !strings.Contains(string(raw), "From: route") {
		t.Fatalf("unexpected output content:\n%s", string(raw))
	}

	after, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("read session after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("session.json changed after handoff output write")
	}
}

func TestRouteHandoffRejectsRestrictedOutputPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	repo, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "assets", path: filepath.Join(repo, "assets", "handoff.txt"), want: "assets/"},
		{name: "managed", path: filepath.Join(home, ".config", "opencode", "bitsentry", "handoff.txt"), want: "OpenCode managed area"},
		{name: "exports", path: filepath.Join(home, ".bitsentry-ai", "exports", "opencode-skills", "handoff.txt"), want: "OpenCode exports"},
		{name: "backups", path: filepath.Join(home, ".bitsentry-ai", "backups", "opencode", "handoff.txt"), want: "OpenCode backups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRouteCmd(&Runtime{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{"handoff", "--session", sessionID, "--output", tt.path})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected restricted output path error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got %v", tt.want, err)
			}
		})
	}
}

func TestRouteReportMarkdownSectionsPresent(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report: %v", err)
	}

	got := out.String()
	for _, section := range []string{"## Session", "## Brief", "## Handoff", "## Plan", "## Progress", "## Lifecycle", "## Safety"} {
		if !strings.Contains(got, section) {
			t.Fatalf("expected markdown section %q, got:\n%s", section, got)
		}
	}
}

func TestRouteReportJSONKeysPresentAndProgressIncluded(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report --json: %v", err)
	}

	var payload struct {
		Session  map[string]any `json:"session"`
		Brief    map[string]any `json:"brief"`
		Handoff  map[string]any `json:"handoff"`
		Plan     map[string]any `json:"plan"`
		Progress struct {
			CurrentStageID string   `json:"current_stage_id"`
			Completed      []string `json:"completed_stage_ids"`
		} `json:"progress"`
		Lifecycle map[string]any `json:"lifecycle"`
		Safety    struct {
			WouldExecute           bool `json:"would_execute"`
			Autonomous             bool `json:"autonomous"`
			ExternalAgentExecution bool `json:"external_agent_execution"`
			SkillExecution         bool `json:"skill_execution"`
			OpenCodeMutation       bool `json:"opencode_mutation"`
		} `json:"safety"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route report json invalid: %v", err)
	}
	if payload.Session == nil || payload.Brief == nil || payload.Handoff == nil || payload.Plan == nil || payload.Lifecycle == nil {
		t.Fatalf("expected required top-level keys, got %+v", payload)
	}
	if payload.Progress.CurrentStageID == "" {
		t.Fatalf("expected progress.current_stage_id to be present")
	}
	if payload.Safety.WouldExecute || payload.Safety.Autonomous || payload.Safety.ExternalAgentExecution || payload.Safety.SkillExecution || payload.Safety.OpenCodeMutation {
		t.Fatalf("expected all safety flags false, got %+v", payload.Safety)
	}
}

func TestRouteReportLifecycleArchivedAfterArchive(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	archive := newRouteCmd(&Runtime{})
	archive.SetOut(&bytes.Buffer{})
	archive.SetErr(&bytes.Buffer{})
	archive.SetArgs([]string{"archive", "--session", sessionID})
	if err := archive.Execute(); err != nil {
		t.Fatalf("execute route archive: %v", err)
	}

	report := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	report.SetOut(out)
	report.SetErr(out)
	report.SetArgs([]string{"report", "--session", sessionID, "--json"})
	if err := report.Execute(); err != nil {
		t.Fatalf("execute route report --json: %v", err)
	}

	var payload struct {
		Lifecycle struct {
			Archived bool `json:"archived"`
		} `json:"lifecycle"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route report json invalid: %v", err)
	}
	if !payload.Lifecycle.Archived {
		t.Fatalf("expected lifecycle.archived=true after archive")
	}
}

func TestRouteReportWritesMarkdownOutput(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	outputPath := filepath.Join(t.TempDir(), "report.md")
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID, "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report --output: %v", err)
	}
	if !strings.Contains(out.String(), "Route report written to:") {
		t.Fatalf("expected write confirmation, got:\n%s", out.String())
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read markdown report output: %v", err)
	}
	if !strings.Contains(string(raw), "## Safety") {
		t.Fatalf("expected markdown report content, got:\n%s", string(raw))
	}
}

func TestRouteReportWritesJSONOutput(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	outputPath := filepath.Join(t.TempDir(), "report.json")
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID, "--json", "--output", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report --json --output: %v", err)
	}
	if !strings.Contains(out.String(), "Route report written to:") {
		t.Fatalf("expected write confirmation, got:\n%s", out.String())
	}

	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read json report output: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("report output file is not valid json: %v", err)
	}
	if _, ok := payload["progress"]; !ok {
		t.Fatalf("expected progress key in report output json")
	}
}

func TestRouteReportMissingSessionSafeError(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"report", "--session", "20260101T000000Z-abcdef12"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected safe missing session error, got %v", err)
	}
}

func TestRouteReportRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"report", "--session", "../evil"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteReportDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"report", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route report not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteReportDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"report", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report --json: %v", err)
	}

	var payload struct {
		WouldExecute           bool `json:"would_execute"`
		Autonomous             bool `json:"autonomous"`
		ExternalAgentExecution bool `json:"external_agent_execution"`
		SkillExecution         bool `json:"skill_execution"`
		OpenCodeMutation       bool `json:"opencode_mutation"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route report json: %v", err)
	}
	if payload.WouldExecute || payload.Autonomous || payload.ExternalAgentExecution || payload.SkillExecution || payload.OpenCodeMutation {
		t.Fatalf("expected execution and mutation flags false, got %+v", payload)
	}
}

func TestRouteReportOutputDoesNotModifySessionBeforeAfter(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	repo, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	sessionPath := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("read session before: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "report.md")
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"report", "--session", sessionID, "--output", outputPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route report --output: %v", err)
	}

	after, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("read session after: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("session.json changed after route report output write")
	}
}

func TestRouteReportRejectsRestrictedOutputPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	repo, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "assets", path: filepath.Join(repo, "assets", "report.md"), want: "assets/"},
		{name: "managed", path: filepath.Join(home, ".config", "opencode", "bitsentry", "report.md"), want: "OpenCode managed area"},
		{name: "exports", path: filepath.Join(home, ".bitsentry-ai", "exports", "opencode-skills", "report.md"), want: "OpenCode exports"},
		{name: "backups", path: filepath.Join(home, ".bitsentry-ai", "backups", "opencode", "report.md"), want: "OpenCode backups"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRouteCmd(&Runtime{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs([]string{"report", "--session", sessionID, "--output", tt.path})

			err := cmd.Execute()
			if err == nil {
				t.Fatalf("expected restricted output path error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error to contain %q, got %v", tt.want, err)
			}
		})
	}
}

func TestRoutePreviewStillDoesNotPersistSession(t *testing.T) {
	repo := setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"preview", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route preview: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, ".bitsentry-ai", "sessions")); !os.IsNotExist(err) {
		t.Fatalf("expected no sessions dir for preview")
	}
}

func TestRouteNormalStillDoesNotPersistSession(t *testing.T) {
	repo := setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route command: %v", err)
	}

	if _, err := os.Stat(filepath.Join(repo, ".bitsentry-ai", "sessions")); !os.IsNotExist(err) {
		t.Fatalf("expected no sessions dir for route command")
	}
}

func TestRouteStartDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"start", "design feature change"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route start: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route start not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)

	tests := []struct {
		name string
		args []string
	}{
		{name: "status traversal", args: []string{"status", "--session", "../evil"}},
		{name: "handoff absolute", args: []string{"handoff", "--session", "/tmp/evil"}},
		{name: "status separator", args: []string{"status", "--session", "a/b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRouteCmd(&Runtime{})
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), "invalid session id") {
				t.Fatalf("expected invalid session id error, got %v", err)
			}
		})
	}
}

func TestRouteResumeRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"resume", "--session", "../evil"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteNextRejectsMissingSession(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"next", "--session", "20260101T000000Z-abcdef12"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected next missing session error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected safe not found error, got %v", err)
	}
}

func TestRouteResumeDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"resume", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route resume: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route resume not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteNextDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"next", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route next: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route next not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteResumeDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"resume", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route resume --json: %v", err)
	}

	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route resume json invalid: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected route resume would_execute=false")
	}
}

func TestRouteNextDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"next", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route next --json: %v", err)
	}

	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("route next json invalid: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected route next would_execute=false")
	}
}

func TestRouteStatusMissingSessionReturnsSafeError(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"status", "--session", "20260101T000000Z-abcdef12"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected status missing session error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected safe not found error, got %v", err)
	}
}

func TestRouteProgressShowsInitialState(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"progress", "--session", sessionID})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route progress: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Current stage: 1/apply") || !strings.Contains(got, "=> pending") {
		t.Fatalf("unexpected route progress output:\n%s", got)
	}
}

func TestRouteProgressSupportsJSON(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"progress", "--session", sessionID, "--json"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route progress --json: %v", err)
	}
	var payload struct {
		SessionID    string `json:"session_id"`
		WouldExecute bool   `json:"would_execute"`
		Stages       []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"stages"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route progress json: %v", err)
	}
	if payload.SessionID != sessionID || payload.WouldExecute || len(payload.Stages) != 2 {
		t.Fatalf("unexpected payload: %+v", payload)
	}
}

func TestRouteMarkCurrentUpdatesProgress(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mark-current", "--session", sessionID, "--stage", "verify"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route mark-current: %v", err)
	}

	progress := runRouteProgressJSON(t, sessionID)
	if progress.CurrentStage.ID != "verify" {
		t.Fatalf("expected current stage verify, got %+v", progress.CurrentStage)
	}
}

func TestRouteMarkDoneUpdatesProgress(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"mark-done", "--session", sessionID, "--stage", "apply", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route mark-done --json: %v", err)
	}
	var payload struct {
		SessionID string `json:"session_id"`
		StageID   string `json:"stage_id"`
		State     string `json:"state"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid mark-done json: %v", err)
	}
	if payload.SessionID != sessionID || payload.StageID != "apply" || payload.State != "done" {
		t.Fatalf("unexpected mark-done payload: %+v", payload)
	}
}

func TestRouteMarkDoneAdvancesCurrentStage(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mark-done", "--session", sessionID, "--stage", "apply"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route mark-done: %v", err)
	}

	progress := runRouteProgressJSON(t, sessionID)
	if progress.CurrentStage == nil || progress.CurrentStage.ID != "verify" {
		t.Fatalf("expected advance to verify stage, got %+v", progress.CurrentStage)
	}
}

func TestRouteMarkRejectsUnknownStage(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mark-current", "--session", sessionID, "--stage", "missing-stage"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "unknown stage id") {
		t.Fatalf("expected unknown stage error, got %v", err)
	}
}

func TestRouteProgressRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"progress", "--session", "../evil"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteMarkDoneRejectsMissingSession(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mark-done", "--session", "20260101T000000Z-abcdef12", "--stage", "apply"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected missing session error, got %v", err)
	}
}

func TestRouteNextUsesProgress(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)

	markDone(t, sessionID, "apply")

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"next", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route next --json: %v", err)
	}
	var payload struct {
		Stage *struct {
			ID string `json:"id"`
		} `json:"stage"`
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route next json: %v", err)
	}
	if payload.Stage == nil || payload.Stage.ID != "verify" || payload.WouldExecute {
		t.Fatalf("unexpected route next payload: %+v", payload)
	}
}

func TestRouteResumeUsesProgress(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	markDone(t, sessionID, "apply")

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"resume", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route resume --json: %v", err)
	}
	var payload struct {
		CurrentStage *struct {
			ID string `json:"id"`
		} `json:"current_stage"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route resume json: %v", err)
	}
	if payload.CurrentStage == nil || payload.CurrentStage.ID != "verify" {
		t.Fatalf("expected current stage verify, got %+v", payload.CurrentStage)
	}
}

func TestRouteProgressDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"progress", "--session", sessionID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route progress: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route progress not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteMarkDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	markDone(t, sessionID, "apply")

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route mark not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteMarkDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"mark-done", "--session", sessionID, "--stage", "apply", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route mark-done --json: %v", err)
	}
	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route mark json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected route mark would_execute=false")
	}
}

func TestLegacySessionWithoutProgressStillWorks(t *testing.T) {
	repo := setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session file: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse session file: %v", err)
	}
	delete(doc, "progress")
	raw, err = json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("marshal legacy session: %v", err)
	}
	if err := os.WriteFile(sessionFile, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write legacy session: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"progress", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute progress for legacy session: %v", err)
	}
	var payload struct {
		CurrentStage *struct {
			ID string `json:"id"`
		} `json:"current_stage"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid legacy progress json: %v", err)
	}
	if payload.CurrentStage == nil || payload.CurrentStage.ID != "apply" {
		t.Fatalf("expected fallback current stage apply, got %+v", payload.CurrentStage)
	}
}

func TestRouteListShowsPersistedSessions(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Session ID:") || !strings.Contains(got, "Flow/Status:") {
		t.Fatalf("expected session listing output, got:\n%s", got)
	}
}

func TestRouteListSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --json: %v", err)
	}
	var payload struct {
		Sessions []struct {
			SessionID string `json:"session_id"`
		} `json:"sessions"`
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if len(payload.Sessions) != 1 || payload.Sessions[0].SessionID == "" || payload.WouldExecute {
		t.Fatalf("unexpected route list payload: %+v", payload)
	}
}

func TestRouteListFiltersByFlow(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply"})
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list", "--flow", "support", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --flow support: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 0 {
		t.Fatalf("expected no sessions for support flow, got %d", payload.Total)
	}
}

func TestRouteListFiltersByStatus(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list", "--status", "done", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --status done: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 0 {
		t.Fatalf("expected no sessions for status done, got %d", payload.Total)
	}
}

func TestRouteArchiveMarksSessionArchived(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	archiveSessionForTest(t, sessionID, false)
	session := readPersistedSessionForTest(t, repo, sessionID)
	if !session.Archived || session.ArchivedAt == nil || session.UpdatedAt == "" {
		t.Fatalf("expected archived session metadata, got %+v", session)
	}
}

func TestRouteArchiveSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := archiveSessionForTest(t, sessionID, true)

	var payload struct {
		SessionID    string `json:"session_id"`
		Archived     bool   `json:"archived"`
		WouldExecute bool   `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid archive json: %v", err)
	}
	if payload.SessionID != sessionID || !payload.Archived || payload.WouldExecute {
		t.Fatalf("unexpected archive payload: %+v", payload)
	}
}

func TestRouteListHidesArchivedByDefault(t *testing.T) {
	setupRouteTestRepo(t)
	a := startRouteSessionForTest(t)
	_ = startRouteSessionForTest(t)
	archiveSessionForTest(t, a, false)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --json: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 1 {
		t.Fatalf("expected only non-archived session, got %d", payload.Total)
	}
}

func TestRouteListArchivedShowsArchived(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list", "--archived", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --archived --json: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 1 {
		t.Fatalf("expected archived session list, got total=%d", payload.Total)
	}
}

func TestRouteListAllShowsActiveAndArchived(t *testing.T) {
	setupRouteTestRepo(t)
	a := startRouteSessionForTest(t)
	_ = startRouteSessionForTest(t)
	archiveSessionForTest(t, a, false)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"list", "--all", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list --all --json: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 2 {
		t.Fatalf("expected both sessions, got total=%d", payload.Total)
	}
}

func TestRouteRestoreUnarchivesSession(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)

	restoreSessionForTest(t, sessionID, false)
	session := readPersistedSessionForTest(t, repo, sessionID)
	if session.Archived || session.RestoredAt == nil || session.UpdatedAt == "" {
		t.Fatalf("expected restored session metadata, got %+v", session)
	}
}

func TestRouteRestoreSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)
	out := restoreSessionForTest(t, sessionID, true)

	var payload struct {
		SessionID    string `json:"session_id"`
		Archived     bool   `json:"archived"`
		WouldExecute bool   `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid restore json: %v", err)
	}
	if payload.SessionID != sessionID || payload.Archived || payload.WouldExecute {
		t.Fatalf("unexpected restore payload: %+v", payload)
	}
}

func TestRouteArchiveRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"archive", "--session", "../evil"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteListDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route list: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route list not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteArchiveDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route archive not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteRestoreDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)
	restoreSessionForTest(t, sessionID, false)

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route restore not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteArchiveDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := archiveSessionForTest(t, sessionID, true)
	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route archive json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected route archive would_execute=false")
	}
}

func TestRouteListHandlesCorruptSessionSafely(t *testing.T) {
	repo := setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	corruptDir := filepath.Join(repo, ".bitsentry-ai", "sessions", "20260101T000000Z-abcxyz12")
	if err := os.MkdirAll(corruptDir, 0o755); err != nil {
		t.Fatalf("mkdir corrupt dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "session.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write corrupt session: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("route list should ignore corrupt session, got err: %v", err)
	}
	var payload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid route list json: %v", err)
	}
	if payload.Total != 1 {
		t.Fatalf("expected one valid session listed, got %d", payload.Total)
	}
}

func TestRouteValidateValidSessionPasses(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}

	var payload struct {
		Valid        bool `json:"valid"`
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if !payload.Valid || payload.WouldExecute {
		t.Fatalf("unexpected validate payload: %+v", payload)
	}
}

func TestRouteValidateSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate --json: %v", err)
	}
	var payload struct {
		SessionID string `json:"session_id"`
		Checks    []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if payload.SessionID != sessionID || len(payload.Checks) == 0 {
		t.Fatalf("unexpected validate payload: %+v", payload)
	}
}

func TestRouteValidateDetectsMissingSession(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"validate", "--session", "20260101T000000Z-deadbeef"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRouteValidateRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"validate", "--session", "../evil"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteValidateDetectsProgressStageMissingFromPlan(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session file: %v", err)
	}
	updated := strings.Replace(string(raw), "\"current_stage_id\": \"apply\"", "\"current_stage_id\": \"ghost\"", 1)
	if err := os.WriteFile(sessionFile, []byte(updated), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}
	var payload struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if payload.Valid || !containsAny(payload.Errors, "current stage") {
		t.Fatalf("expected current stage validation error, got %+v", payload)
	}
}

func TestRouteValidateDetectsDuplicateStageIDs(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "apply"})
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}
	var payload struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if payload.Valid || !containsAny(payload.Errors, "duplicate stage ids") {
		t.Fatalf("expected duplicate stage id error, got %+v", payload)
	}
}

func TestRouteValidateDetectsPlanFlowMismatch(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session file: %v", err)
	}
	updated := strings.Replace(string(raw), "\"FlowID\": \"sdd\"", "\"FlowID\": \"support\"", 1)
	if err := os.WriteFile(sessionFile, []byte(updated), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}
	var payload struct {
		Valid  bool     `json:"valid"`
		Errors []string `json:"errors"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if payload.Valid || !containsAny(payload.Errors, "plan flow mismatch") {
		t.Fatalf("expected plan flow mismatch error, got %+v", payload)
	}
}

func TestRouteValidateDoesNotModifySession(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session before validate: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"validate", "--session", sessionID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}
	after, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session after validate: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("session.json changed after validate")
	}
}

func TestRouteValidateDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"validate", "--session", sessionID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}

	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route validate not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteValidateDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route validate: %v", err)
	}
	var payload struct {
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected route validate would_execute=false")
	}
}

func TestRouteAuditSummarizesSessions(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	_ = startRouteSessionForTest(t)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"audit", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route audit: %v", err)
	}
	var payload struct {
		Total  int `json:"total"`
		Passed int `json:"passed"`
		Failed int `json:"failed"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid audit json: %v", err)
	}
	if payload.Total != 2 || payload.Passed != 2 || payload.Failed != 0 {
		t.Fatalf("unexpected audit summary: %+v", payload)
	}
}

func TestRouteAuditSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"audit", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route audit --json: %v", err)
	}
	var payload struct {
		Sessions []struct {
			SessionID string `json:"session_id"`
		} `json:"sessions"`
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid audit json: %v", err)
	}
	if len(payload.Sessions) != 1 || payload.Sessions[0].SessionID == "" || payload.WouldExecute {
		t.Fatalf("unexpected audit payload: %+v", payload)
	}
}

func TestRouteAuditHandlesCorruptSessionSafely(t *testing.T) {
	repo := setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)

	corruptDir := filepath.Join(repo, ".bitsentry-ai", "sessions", "20260101T000000Z-abcxyz12")
	if err := os.MkdirAll(corruptDir, 0o755); err != nil {
		t.Fatalf("mkdir corrupt dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "session.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write corrupt session: %v", err)
	}

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"audit", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("route audit should not panic on corrupt session, got err: %v", err)
	}
	var payload struct {
		Total  int `json:"total"`
		Failed int `json:"failed"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid audit json: %v", err)
	}
	if payload.Total != 2 || payload.Failed != 1 {
		t.Fatalf("unexpected audit summary: %+v", payload)
	}
}

func TestRouteAuditIncludeArchived(t *testing.T) {
	setupRouteTestRepo(t)
	active := startRouteSessionForTest(t)
	archived := startRouteSessionForTest(t)
	archiveSessionForTest(t, archived, false)

	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"audit", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route audit default: %v", err)
	}
	var defaultPayload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out.Bytes(), &defaultPayload); err != nil {
		t.Fatalf("invalid default audit json: %v", err)
	}
	if defaultPayload.Total != 1 {
		t.Fatalf("expected archived excluded by default, got total=%d", defaultPayload.Total)
	}

	cmd2 := newRouteCmd(&Runtime{})
	out2 := &bytes.Buffer{}
	cmd2.SetOut(out2)
	cmd2.SetErr(out2)
	cmd2.SetArgs([]string{"audit", "--json", "--include-archived"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("execute route audit include archived: %v", err)
	}
	var includePayload struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out2.Bytes(), &includePayload); err != nil {
		t.Fatalf("invalid include audit json: %v", err)
	}
	if includePayload.Total != 2 || active == "" {
		t.Fatalf("expected include archived total=2, got %+v", includePayload)
	}
}

func TestRouteAuditDoesNotModifySessions(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session before audit: %v", err)
	}
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"audit", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route audit: %v", err)
	}
	after, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session after audit: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("session.json changed after audit")
	}
}

func TestRouteCleanupDryRunIsReadOnly(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session before cleanup: %v", err)
	}
	runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd"})
	after, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session after cleanup: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("cleanup dry-run must be read-only")
	}
}

func TestRouteCleanupDryRunSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd", "--json"})
	var payload struct {
		Counts struct {
			Candidates int `json:"candidates"`
		} `json:"counts"`
		WouldExecute bool `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid cleanup json: %v", err)
	}
	if payload.Counts.Candidates != 1 || payload.WouldExecute {
		t.Fatalf("unexpected cleanup dry-run json: %+v", payload)
	}
}

func TestRouteCleanupRequiresFilterForApply(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"cleanup", "--apply", "--confirm"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires at least one filter") {
		t.Fatalf("expected missing filter error, got %v", err)
	}
}

func TestRouteCleanupApplyRequiresConfirm(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"cleanup", "--apply", "--flow", "sdd"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("expected confirm error, got %v", err)
	}
}

func TestRouteCleanupApplyArchivesCandidates(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	runRouteCleanupCmd(t, []string{"--apply", "--confirm", "--flow", "sdd"})
	session := readPersistedSessionForTest(t, repo, sessionID)
	if !session.Archived || session.ArchivedAt == nil {
		t.Fatalf("expected archived metadata after cleanup apply")
	}
}

func TestRouteCleanupDoesNotPhysicallyDeleteSessions(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	runRouteCleanupCmd(t, []string{"--apply", "--confirm", "--flow", "sdd"})
	sessionDir := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID)
	if _, err := os.Stat(sessionDir); err != nil {
		t.Fatalf("session directory should still exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(sessionDir, "brief.md")); err != nil {
		t.Fatalf("brief.md should not be deleted: %v", err)
	}
}

func TestRouteCleanupFiltersByFlow(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "support", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	_ = json.Unmarshal(out.Bytes(), &payload)
	if payload.Counts.Candidates != 0 {
		t.Fatalf("expected zero candidates for unmatched flow")
	}
}

func TestRouteCleanupFiltersByStatus(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--status", "done", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	_ = json.Unmarshal(out.Bytes(), &payload)
	if payload.Counts.Candidates != 0 {
		t.Fatalf("expected zero candidates for unmatched status")
	}
}

func TestRouteCleanupOlderThanFilter(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	setSessionUpdatedAtForTest(t, repo, sessionID, time.Now().UTC().Add(-2*time.Hour))
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--older-than", "1h", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid cleanup json: %v", err)
	}
	if payload.Counts.Candidates != 1 {
		t.Fatalf("expected older-than filter to match one candidate, got %d", payload.Counts.Candidates)
	}
}

func TestRouteCleanupCompletedOnly(t *testing.T) {
	setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	markDone(t, sessionID, "apply")
	markDone(t, sessionID, "verify")
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--completed-only", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid cleanup json: %v", err)
	}
	if payload.Counts.Candidates != 1 {
		t.Fatalf("expected completed session candidate, got %d", payload.Counts.Candidates)
	}
}

func TestRouteCleanupExcludesArchivedByDefault(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	_ = json.Unmarshal(out.Bytes(), &payload)
	if payload.Counts.Candidates != 0 {
		t.Fatalf("expected archived session excluded by default")
	}
}

func TestRouteCleanupIncludeArchived(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	archiveSessionForTest(t, sessionID, false)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd", "--include-archived", "--json"})
	var payload struct{ Counts struct{ Candidates int `json:"candidates"` } `json:"counts"` }
	_ = json.Unmarshal(out.Bytes(), &payload)
	if payload.Counts.Candidates != 1 {
		t.Fatalf("expected archived session included with --include-archived")
	}
}

func TestRouteCleanupHandlesCorruptSessionSafely(t *testing.T) {
	repo := setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	corruptDir := filepath.Join(repo, ".bitsentry-ai", "sessions", "20260101T000000Z-abcxyz12")
	if err := os.MkdirAll(corruptDir, 0o755); err != nil {
		t.Fatalf("mkdir corrupt dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(corruptDir, "session.json"), []byte("{"), 0o644); err != nil {
		t.Fatalf("write corrupt session: %v", err)
	}
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd", "--json"})
	var payload struct {
		Counts struct{ Candidates int `json:"candidates"`; Skipped int `json:"skipped"` } `json:"counts"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid cleanup json: %v", err)
	}
	if payload.Counts.Candidates != 1 || payload.Counts.Skipped != 1 {
		t.Fatalf("expected one valid candidate and one skipped corrupt session, got %+v", payload.Counts)
	}
}

func TestRouteCleanupDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd"})
	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route cleanup not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteCleanupDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	out := runRouteCleanupCmd(t, []string{"--dry-run", "--flow", "sdd", "--json"})
	var payload struct{ WouldExecute bool `json:"would_execute"` }
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid cleanup json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected cleanup would_execute=false")
	}
}

func TestRouteCleanupDoesNotTouchAssets(t *testing.T) {
	repo := setupRouteTestRepo(t)
	_ = startRouteSessionForTest(t)
	flowPath := filepath.Join(repo, "assets", "flows", "sdd.yaml")
	before, err := os.ReadFile(flowPath)
	if err != nil {
		t.Fatalf("read assets flow before cleanup: %v", err)
	}
	runRouteCleanupCmd(t, []string{"--apply", "--confirm", "--flow", "sdd"})
	after, err := os.ReadFile(flowPath)
	if err != nil {
		t.Fatalf("read assets flow after cleanup: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("cleanup must not modify assets")
	}
}

func TestRouteRepairDryRunIsReadOnly(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	sessionFile := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	before, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session before repair: %v", err)
	}
	runRouteRepairCmd(t, []string{"--session", sessionID, "--dry-run"})
	after, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read session after repair: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("repair dry-run must be read-only")
	}
}

func TestRouteRepairDryRunSupportsJSON(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := runRouteRepairCmd(t, []string{"--session", sessionID, "--dry-run", "--json"})
	var payload struct {
		SessionID    string `json:"session_id"`
		DryRun       bool   `json:"dry_run"`
		WouldExecute bool   `json:"would_execute"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid repair json: %v", err)
	}
	if payload.SessionID != sessionID || !payload.DryRun || payload.WouldExecute {
		t.Fatalf("unexpected repair dry-run payload: %+v", payload)
	}
}

func TestRouteRepairApplyRequiresConfirm(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"repair", "--session", sessionID, "--apply"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires --confirm") {
		t.Fatalf("expected confirm error, got %v", err)
	}
}

func TestRouteRepairRequiresMode(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"repair", "--session", sessionID})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "requires one mode") {
		t.Fatalf("expected mode requirement error, got %v", err)
	}
}

func TestRouteRepairCreatesMissingBrief(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	briefPath := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "brief.md")
	if err := os.Remove(briefPath); err != nil {
		t.Fatalf("remove brief: %v", err)
	}
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	if _, err := os.Stat(briefPath); err != nil {
		t.Fatalf("expected brief.md recreated: %v", err)
	}
}

func TestRouteRepairCreatesMissingHandoff(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	handoffPath := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "handoff.md")
	if err := os.Remove(handoffPath); err != nil {
		t.Fatalf("remove handoff: %v", err)
	}
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	if _, err := os.Stat(handoffPath); err != nil {
		t.Fatalf("expected handoff.md recreated: %v", err)
	}
}

func TestRouteRepairInitializesMissingProgress(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		delete(payload, "progress")
	})
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	raw := mustReadSessionFile(t, repo, sessionID)
	if !strings.Contains(raw, "\"progress\"") || !strings.Contains(raw, "\"current_stage_id\": \"apply\"") {
		t.Fatalf("expected initialized progress, got:\n%s", raw)
	}
}

func TestRouteRepairRemovesDuplicateCompletedStages(t *testing.T) {
	repo := setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		progress := payload["progress"].(map[string]any)
		progress["completed_stage_ids"] = []any{"apply", "apply", "verify"}
	})
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	raw := mustReadSessionFile(t, repo, sessionID)
	if strings.Count(raw, "\"apply\"") < 1 {
		t.Fatalf("expected apply stage present after dedupe")
	}
	if strings.Contains(raw, "\"completed_stage_ids\": [\n      \"apply\",\n      \"apply\"") {
		t.Fatalf("expected duplicate completed stage removed")
	}
}

func TestRouteRepairDropsUnknownCompletedStages(t *testing.T) {
	repo := setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		progress := payload["progress"].(map[string]any)
		progress["completed_stage_ids"] = []any{"ghost", "apply"}
	})
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	raw := mustReadSessionFile(t, repo, sessionID)
	if strings.Contains(raw, "ghost") {
		t.Fatalf("expected unknown completed stage removed")
	}
}

func TestRouteRepairClearsUnknownCurrentStage(t *testing.T) {
	repo := setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		progress := payload["progress"].(map[string]any)
		progress["current_stage_id"] = "ghost"
	})
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	raw := mustReadSessionFile(t, repo, sessionID)
	if strings.Contains(raw, "\"current_stage_id\": \"ghost\"") {
		t.Fatalf("expected unknown current stage cleared")
	}
}

func TestRouteRepairRejectsInvalidSessionIDPathTraversal(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"repair", "--session", "../evil", "--dry-run"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid session id") {
		t.Fatalf("expected invalid session id error, got %v", err)
	}
}

func TestRouteRepairMissingSessionReturnsSafeError(t *testing.T) {
	setupRouteTestRepo(t)
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"repair", "--session", "20260101T000000Z-deadbeef", "--dry-run"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRouteRepairDoesNotMutateOpenCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	runRouteRepairCmd(t, []string{"--session", sessionID, "--dry-run"})
	opencodeConfig := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(opencodeConfig); !os.IsNotExist(err) {
		t.Fatalf("expected route repair not to mutate opencode path %s", opencodeConfig)
	}
}

func TestRouteRepairDoesNotTouchAssets(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	flowPath := filepath.Join(repo, "assets", "flows", "sdd.yaml")
	before, err := os.ReadFile(flowPath)
	if err != nil {
		t.Fatalf("read assets flow before repair: %v", err)
	}
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	after, err := os.ReadFile(flowPath)
	if err != nil {
		t.Fatalf("read assets flow after repair: %v", err)
	}
	if string(before) != string(after) {
		t.Fatalf("repair must not modify assets")
	}
}

func TestRouteRepairDoesNotExecuteAgents(t *testing.T) {
	setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	out := runRouteRepairCmd(t, []string{"--session", sessionID, "--dry-run", "--json"})
	var payload struct{ WouldExecute bool `json:"would_execute"` }
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid repair json: %v", err)
	}
	if payload.WouldExecute {
		t.Fatalf("expected repair would_execute=false")
	}
}

func TestRouteRepairUnrepairableMissingPlan(t *testing.T) {
	repo := setupRouteTestRepo(t)
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		payload["plan"] = map[string]any{"FlowID": "sdd", "Stages": nil}
	})
	out := runRouteRepairCmd(t, []string{"--session", sessionID, "--dry-run", "--json"})
	var payload struct {
		Repairable   bool     `json:"repairable"`
		Unrepairable []string `json:"unrepairable"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid repair json: %v", err)
	}
	if payload.Repairable || len(payload.Unrepairable) == 0 {
		t.Fatalf("expected unrepairable missing plan, got %+v", payload)
	}
}

func TestRouteRepairThenValidatePasses(t *testing.T) {
	repo := setupRouteTestRepoWithFlow(t, []string{"apply", "verify"})
	sessionID := startRouteSessionForTest(t)
	setSessionJSONForTest(t, repo, sessionID, func(payload map[string]any) {
		progress := payload["progress"].(map[string]any)
		progress["current_stage_id"] = "ghost"
		progress["completed_stage_ids"] = []any{"apply", "apply", "ghost"}
	})
	runRouteRepairCmd(t, []string{"--session", sessionID, "--apply", "--confirm"})
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"validate", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("validate after repair: %v", err)
	}
	var payload struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid validate json: %v", err)
	}
	if !payload.Valid {
		t.Fatalf("expected repaired session to validate")
	}
}

func containsAny(values []string, needle string) bool {
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func startRouteSessionForTest(t *testing.T) string {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"start", "--json", "design feature change"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route start: %v", err)
	}
	var payload struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("parse route start json: %v", err)
	}
	if payload.SessionID == "" {
		t.Fatalf("missing session id in route start output")
	}
	if ok, _ := regexp.MatchString(`^[0-9]{8}T[0-9]{6}Z-[a-f0-9]{8}$`, payload.SessionID); !ok {
		t.Fatalf("unexpected session id format: %s", payload.SessionID)
	}
	return payload.SessionID
}

func setupRouteTestRepo(t *testing.T) string {
	t.Helper()
	return setupRouteTestRepoWithFlow(t, []string{"apply"})
}

func setupRouteTestRepoWithFlow(t *testing.T, stageIDs []string) string {
	t.Helper()
	repo := t.TempDir()
	assetsDir := filepath.Join(repo, "assets", "flows")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("mkdir assets flows: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "assets", "skills", "_shared"), 0o755); err != nil {
		t.Fatalf("mkdir assets skills: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "assets", "orchestrators"), 0o755); err != nil {
		t.Fatalf("mkdir assets orchestrators: %v", err)
	}
	var flowBuilder strings.Builder
	flowBuilder.WriteString("id: sdd\n")
	flowBuilder.WriteString("name: Spec Driven Development\n")
	flowBuilder.WriteString("kind: user\n")
	flowBuilder.WriteString("selectable: true\n")
	flowBuilder.WriteString("top_level_flow: true\n")
	flowBuilder.WriteString("family: core\n")
	flowBuilder.WriteString("skill_pack: sdd\n")
	flowBuilder.WriteString("orchestrator_skill: sdd-orchestrator\n")
	flowBuilder.WriteString("status: active\n")
	flowBuilder.WriteString("stages:\n")
	for _, stageID := range stageIDs {
		flowBuilder.WriteString("  - id: " + stageID + "\n")
		flowBuilder.WriteString("    skill: sdd-" + stageID + "\n")
		flowBuilder.WriteString("    description: " + stageID + " stage\n")
	}
	flow := flowBuilder.String()
	if err := os.WriteFile(filepath.Join(assetsDir, "sdd.yaml"), []byte(flow), 0o644); err != nil {
		t.Fatalf("write sdd flow: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(repo); err != nil {
		t.Fatalf("chdir temp repo: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	return repo
}

func markDone(t *testing.T, sessionID string, stageID string) {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"mark-done", "--session", sessionID, "--stage", stageID})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mark stage done: %v", err)
	}
}

func runRouteProgressJSON(t *testing.T, sessionID string) struct {
	CurrentStage *struct {
		ID string `json:"id"`
	} `json:"current_stage"`
	Stages []struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"stages"`
} {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	cmd.SetArgs([]string{"progress", "--session", sessionID, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route progress --json: %v", err)
	}
	var payload struct {
		CurrentStage *struct {
			ID string `json:"id"`
		} `json:"current_stage"`
		Stages []struct {
			ID    string `json:"id"`
			State string `json:"state"`
		} `json:"stages"`
	}
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("parse progress json: %v", err)
	}
	return payload
}

func archiveSessionForTest(t *testing.T, sessionID string, jsonOut bool) *bytes.Buffer {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	args := []string{"archive", "--session", sessionID}
	if jsonOut {
		args = append(args, "--json")
	}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route archive: %v", err)
	}
	return out
}

func runRouteCleanupCmd(t *testing.T, args []string) *bytes.Buffer {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	allArgs := append([]string{"cleanup"}, args...)
	cmd.SetArgs(allArgs)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route cleanup: %v", err)
	}
	return out
}

func runRouteRepairCmd(t *testing.T, args []string) *bytes.Buffer {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	allArgs := append([]string{"repair"}, args...)
	cmd.SetArgs(allArgs)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route repair: %v", err)
	}
	return out
}

func runRouteMigrateCmd(t *testing.T, args []string) *bytes.Buffer {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	allArgs := append([]string{"migrate"}, args...)
	cmd.SetArgs(allArgs)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route migrate: %v", err)
	}
	return out
}

func setSessionJSONForTest(t *testing.T, repo string, sessionID string, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read session json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse session json: %v", err)
	}
	mutate(payload)
	updated, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal session json: %v", err)
	}
	if err := os.WriteFile(path, append(updated, '\n'), 0o644); err != nil {
		t.Fatalf("write session json: %v", err)
	}
}

func mustReadSessionFile(t *testing.T, repo string, sessionID string) string {
	t.Helper()
	path := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read session json: %v", err)
	}
	return string(raw)
}

func setSessionUpdatedAtForTest(t *testing.T, repo string, sessionID string, ts time.Time) {
	t.Helper()
	path := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read session json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse session json: %v", err)
	}
	payload["updated_at"] = ts.UTC().Format(time.RFC3339)
	updated, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal updated session json: %v", err)
	}
	if err := os.WriteFile(path, append(updated, '\n'), 0o644); err != nil {
		t.Fatalf("write session json: %v", err)
	}
}

func restoreSessionForTest(t *testing.T, sessionID string, jsonOut bool) *bytes.Buffer {
	t.Helper()
	cmd := newRouteCmd(&Runtime{})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(out)
	args := []string{"restore", "--session", sessionID}
	if jsonOut {
		args = append(args, "--json")
	}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute route restore: %v", err)
	}
	return out
}

func readPersistedSessionForTest(t *testing.T, repo string, sessionID string) struct {
	Archived   bool    `json:"archived"`
	ArchivedAt *string `json:"archived_at"`
	RestoredAt *string `json:"restored_at"`
	UpdatedAt  string  `json:"updated_at"`
} {
	t.Helper()
	path := filepath.Join(repo, ".bitsentry-ai", "sessions", sessionID, "session.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read session json: %v", err)
	}
	var payload struct {
		Archived   bool    `json:"archived"`
		ArchivedAt *string `json:"archived_at"`
		RestoredAt *string `json:"restored_at"`
		UpdatedAt  string  `json:"updated_at"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("parse session json: %v", err)
	}
	if payload.UpdatedAt == "" {
		t.Fatalf("expected updated_at in session json")
	}
	return struct {
		Archived   bool    `json:"archived"`
		ArchivedAt *string `json:"archived_at"`
		RestoredAt *string `json:"restored_at"`
		UpdatedAt  string  `json:"updated_at"`
	}{
		Archived:   payload.Archived,
		ArchivedAt: payload.ArchivedAt,
		RestoredAt: payload.RestoredAt,
		UpdatedAt:  payload.UpdatedAt,
	}
}
