package system

import (
	"path/filepath"
	"strings"

	"os"
)

func DetectShell() string {
	sh := strings.TrimSpace(os.Getenv("SHELL"))
	if sh == "" {
		return "unknown"
	}
	return filepath.Base(sh)
}
