package config

import (
	"os"
	"path/filepath"
)

const appDirName = ".bitsentry-ai"

func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "bitsentry-ai")
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return appDirName
	}

	return filepath.Join(home, appDirName)
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func LogsDir() string {
	return filepath.Join(ConfigDir(), "logs")
}
