package securityweb

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoNetHTTPImportInPackage(t *testing.T) {
	assertNoImport(t, "\"net/http\"")
}

func TestNoOSExecImportInPackage(t *testing.T) {
	assertNoImport(t, "\"os/exec\"")
}

func assertNoImport(t *testing.T, banned string) {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		bytes, err := os.ReadFile(filepath.Join(".", name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if strings.Contains(string(bytes), banned) {
			t.Fatalf("banned import %s found in %s", banned, name)
		}
	}
}
