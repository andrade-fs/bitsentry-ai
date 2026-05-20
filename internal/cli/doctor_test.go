package cli

import "testing"

func TestReadinessVerdict(t *testing.T) {
	t.Run("fail when OpenCode missing", func(t *testing.T) {
		got := readinessVerdict(mvpReadiness{OpenCodeDetected: false, OpenCodeConfigRoot: ""})
		if got != "FAIL" {
			t.Fatalf("expected FAIL, got %s", got)
		}
	})

	t.Run("pass with notes when pack not installed", func(t *testing.T) {
		got := readinessVerdict(mvpReadiness{OpenCodeDetected: true, OpenCodeConfigRoot: "/tmp/.opencode", BitsentryPackStatus: "not installed"})
		if got != "PASS WITH NOTES" {
			t.Fatalf("expected PASS WITH NOTES, got %s", got)
		}
	})

	t.Run("pass with notes when edit contract invalid", func(t *testing.T) {
		got := readinessVerdict(mvpReadiness{OpenCodeDetected: true, OpenCodeConfigRoot: "/tmp/.opencode", BitsentryPackStatus: "installed", EditPermissionStatus: "invalid value (allow)"})
		if got != "PASS WITH NOTES" {
			t.Fatalf("expected PASS WITH NOTES, got %s", got)
		}
	})

	t.Run("pass when core readiness met", func(t *testing.T) {
		got := readinessVerdict(mvpReadiness{OpenCodeDetected: true, OpenCodeConfigRoot: "/tmp/.opencode", BitsentryPackStatus: "installed", EditPermissionStatus: "agent.bitsentry.permission.edit=deny"})
		if got != "PASS" {
			t.Fatalf("expected PASS, got %s", got)
		}
	})
}
