package logs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bitsentry-ai/internal/config"
)

func New() (*log.Logger, func() error, error) {
	logsDir := config.LogsDir()
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create logs directory: %w", err)
	}

	logFile := filepath.Join(logsDir, "app.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open app log file: %w", err)
	}

	l := log.New(f, "bitsentry-ai ", log.Ldate|log.Ltime|log.Lshortfile)
	return l, f.Close, nil
}
