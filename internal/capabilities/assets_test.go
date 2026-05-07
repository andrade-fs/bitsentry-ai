package capabilities

import (
	"os"
	"testing"
)

func TestFlowManifestsExist(t *testing.T) {
	paths := []string{
		"../../assets/flows/sdd.yaml",
		"../../assets/flows/sdr.yaml",
		"../../assets/flows/support.yaml",
	}
	for _, p := range paths {
		if !fileExists(p) {
			t.Fatalf("expected flow manifest to exist: %s", p)
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
