package orchestrator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{5,127}$`)

type SessionStore interface {
	SaveSession(session Session) (string, error)
	SaveSessionMetadata(session Session) (string, error)
	LoadSession(sessionID string) (Session, string, error)
}

type LocalSessionStore struct {
	root string
}

func (s LocalSessionStore) Root() string {
	return s.root
}

func NewLocalSessionStore(root string) LocalSessionStore {
	return LocalSessionStore{root: root}
}

func (s LocalSessionStore) SaveSession(session Session) (string, error) {
	if err := validateSessionID(session.ID); err != nil {
		return "", err
	}
	sessionDir := s.sessionDir(session.ID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return "", fmt.Errorf("create session directory: %w", err)
	}

	if err := writeJSON(filepath.Join(sessionDir, "session.json"), session); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "brief.md"), []byte(RenderBriefMarkdown(session.Brief)), 0o644); err != nil {
		return "", fmt.Errorf("write brief.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "handoff.md"), []byte(RenderHandoffMarkdown(session.Handoff)), 0o644); err != nil {
		return "", fmt.Errorf("write handoff.md: %w", err)
	}
	return sessionDir, nil
}

func (s LocalSessionStore) LoadSession(sessionID string) (Session, string, error) {
	if err := validateSessionID(sessionID); err != nil {
		return Session{}, "", err
	}
	path := filepath.Join(s.sessionDir(sessionID), "session.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Session{}, "", fmt.Errorf("session %q not found", sessionID)
		}
		return Session{}, "", fmt.Errorf("read session: %w", err)
	}
	var session Session
	if err := json.Unmarshal(raw, &session); err != nil {
		return Session{}, "", fmt.Errorf("parse session: %w", err)
	}
	return session, s.sessionDir(sessionID), nil
}

func (s LocalSessionStore) SaveSessionMetadata(session Session) (string, error) {
	if err := validateSessionID(session.ID); err != nil {
		return "", err
	}
	sessionDir := s.sessionDir(session.ID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return "", fmt.Errorf("create session directory: %w", err)
	}
	if err := writeJSON(filepath.Join(sessionDir, "session.json"), session); err != nil {
		return "", err
	}
	return sessionDir, nil
}

func (s LocalSessionStore) sessionDir(sessionID string) string {
	return filepath.Join(s.root, ".bitsentry-ai", "sessions", sessionID)
}

func validateSessionID(sessionID string) error {
	trimmed := strings.TrimSpace(sessionID)
	if trimmed == "" {
		return fmt.Errorf("session id is required")
	}
	if filepath.IsAbs(trimmed) || strings.Contains(trimmed, "..") || strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") {
		return fmt.Errorf("invalid session id")
	}
	if !sessionIDPattern.MatchString(trimmed) {
		return fmt.Errorf("invalid session id")
	}
	return nil
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filepath.Base(path), err)
	}
	return nil
}
