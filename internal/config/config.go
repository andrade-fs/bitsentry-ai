package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ActiveProfile string           `yaml:"activeProfile"`
	Components    ComponentsConfig `yaml:"components,omitempty"`
}

type ComponentsConfig struct {
	Engram   EngramComponentConfig   `yaml:"engram,omitempty"`
	Context7 Context7ComponentConfig `yaml:"context7,omitempty"`
	MCPs     MCPsComponentConfig     `yaml:"mcps,omitempty"`
	Skills   SkillsComponentConfig   `yaml:"skills,omitempty"`
	Flows    FlowsComponentConfig    `yaml:"flows,omitempty"`
	Targets  TargetsComponentConfig  `yaml:"targets,omitempty"`
	Preset   string                  `yaml:"preset,omitempty"`
}

type FlowsComponentConfig struct {
	Enabled    bool     `yaml:"enabled,omitempty"`
	Configured bool     `yaml:"configured,omitempty"`
	Selected   []string `yaml:"selected,omitempty"`
}

type TargetsComponentConfig struct {
	Selected []string `yaml:"selected,omitempty"`
}

type SkillsComponentConfig struct {
	Enabled    bool     `yaml:"enabled,omitempty"`
	Configured bool     `yaml:"configured,omitempty"`
	Selected   []string `yaml:"selected,omitempty"`
}

type MCPsComponentConfig struct {
	Enabled    bool     `yaml:"enabled,omitempty"`
	Configured bool     `yaml:"configured,omitempty"`
	Selected   []string `yaml:"selected,omitempty"`
}

type EngramComponentConfig struct {
	Enabled    bool   `yaml:"enabled,omitempty"`
	Configured bool   `yaml:"configured,omitempty"`
	BinaryPath string `yaml:"binary_path,omitempty"`
	DataDir    string `yaml:"data_dir,omitempty"`
	Project    string `yaml:"project,omitempty"`
}

type Context7ComponentConfig struct {
	Enabled    bool   `yaml:"enabled,omitempty"`
	Configured bool   `yaml:"configured,omitempty"`
	Command    string `yaml:"command,omitempty"`
	Package    string `yaml:"package,omitempty"`
	Notes      string `yaml:"notes,omitempty"`
}

type Manager struct {
	path string
}

func NewManager() *Manager {
	return &Manager{path: ConfigPath()}
}

func (m *Manager) Load() (Config, error) {
	if err := m.ensureDir(); err != nil {
		return Config{}, err
	}

	if _, err := os.Stat(m.path); errors.Is(err, os.ErrNotExist) {
		cfg := Config{ActiveProfile: "default"}
		if err := m.Save(cfg); err != nil {
			return Config{}, err
		}
		return cfg, nil
	}

	b, err := os.ReadFile(m.path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}
	if cfg.ActiveProfile == "" {
		cfg.ActiveProfile = "default"
	}

	return cfg, nil
}

func (m *Manager) Save(cfg Config) error {
	if err := m.ensureDir(); err != nil {
		return err
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config yaml: %w", err)
	}

	if err := os.WriteFile(m.path, b, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (m *Manager) SetActiveProfile(profile string) error {
	if profile == "" {
		return errors.New("profile cannot be empty")
	}

	cfg, err := m.Load()
	if err != nil {
		return err
	}

	cfg.ActiveProfile = profile
	return m.Save(cfg)
}

func (m *Manager) Path() string {
	return m.path
}

func (m *Manager) ensureDir() error {
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return fmt.Errorf("ensure config directory: %w", err)
	}
	return nil
}
