package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type OpenCodeConfigFileInspection struct {
	Path         string
	Type         string
	SizeBytes    int64
	Readable     bool
	TopLevelKeys []string
	ParseWarning string
}

func InferOpenCodeFileType(path string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(path)))
	switch ext {
	case ".yaml", ".yml":
		return "yaml"
	case ".json":
		return "json"
	default:
		return "unknown"
	}
}

func FindLikelyOpenCodeConfigFiles(configDir string) ([]string, []string) {
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to list config dir %s: %v", configDir, err)}
	}

	knownNames := map[string]struct{}{
		"config.yaml":   {},
		"config.yml":    {},
		"config.json":   {},
		"opencode.yaml": {},
		"opencode.yml":  {},
		"opencode.json": {},
		"settings.json": {},
	}

	files := make([]string, 0)
	warnings := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		lowerName := strings.ToLower(name)
		_, isKnown := knownNames[lowerName]
		isObviousRelated := strings.Contains(lowerName, "mcp") || strings.Contains(lowerName, "agent") || strings.Contains(lowerName, "skill")

		if isKnown || isObviousRelated {
			files = append(files, filepath.Join(configDir, name))
		}
	}

	sort.Strings(files)
	if len(files) == 0 {
		warnings = append(warnings, fmt.Sprintf("no likely config files found directly under %s", configDir))
	}

	return files, warnings
}

func InspectOpenCodeConfigFile(path string) OpenCodeConfigFileInspection {
	inspection := OpenCodeConfigFileInspection{
		Path: path,
		Type: InferOpenCodeFileType(path),
	}

	st, err := os.Stat(path)
	if err != nil {
		inspection.ParseWarning = fmt.Sprintf("stat failed: %v", err)
		return inspection
	}
	inspection.SizeBytes = st.Size()

	f, openErr := os.Open(path)
	if openErr != nil {
		inspection.Readable = false
		inspection.ParseWarning = fmt.Sprintf("not readable: %v", openErr)
		return inspection
	}
	inspection.Readable = true
	_ = f.Close()

	if inspection.Type == "unknown" {
		return inspection
	}

	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		inspection.ParseWarning = fmt.Sprintf("read failed: %v", readErr)
		return inspection
	}

	keys, parseErr := parseTopLevelKeys(raw, inspection.Type)
	if parseErr != nil {
		inspection.ParseWarning = parseErr.Error()
		return inspection
	}

	inspection.TopLevelKeys = keys
	return inspection
}

func parseTopLevelKeys(raw []byte, fileType string) ([]string, error) {
	var payload any
	var err error

	switch fileType {
	case "json":
		err = json.Unmarshal(raw, &payload)
	case "yaml":
		err = yaml.Unmarshal(raw, &payload)
	default:
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	keys := extractMapKeys(payload)
	sort.Strings(keys)
	return keys, nil
}

func extractMapKeys(v any) []string {
	switch m := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		return keys
	case map[any]any:
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, fmt.Sprint(k))
		}
		return keys
	default:
		return nil
	}
}
