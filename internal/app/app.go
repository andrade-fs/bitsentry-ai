package app

import (
	"context"
	"fmt"
	"log"

	"bitsentry-ai/internal/agents"
	"bitsentry-ai/internal/config"
	"bitsentry-ai/internal/logs"
	"bitsentry-ai/internal/profiles"
)

type App struct {
	ConfigManager *config.Manager
	ProfileStore  profiles.Store
	AgentRegistry *agents.Registry
	Logger        *log.Logger
	closeLogger   func() error
}

func New() (*App, error) {
	cm := config.NewManager()
	if _, err := cm.Load(); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	l, closeFn, err := logs.New()
	if err != nil {
		return nil, err
	}

	a := &App{
		ConfigManager: cm,
		ProfileStore:  profiles.NewInMemoryStore(),
		AgentRegistry: agents.NewRegistry(agents.OpenCodeDetector{}),
		Logger:        l,
		closeLogger:   closeFn,
	}
	a.Logger.Println("application bootstrapped")

	return a, nil
}

func (a *App) Run() error {
	defer func() {
		if a.closeLogger != nil {
			_ = a.closeLogger()
		}
	}()

	// Fase 1 foundation/core: wiring y contratos mínimos.
	_, _ = a.AgentRegistry.List(context.Background())
	return nil
}

func (a *App) SetActiveProfile(profileID string) error {
	if _, ok := a.ProfileStore.Get(profileID); !ok {
		return fmt.Errorf("profile %q does not exist", profileID)
	}

	if err := a.ProfileStore.Save(profileID); err != nil {
		return err
	}

	if err := a.ConfigManager.SetActiveProfile(profileID); err != nil {
		return err
	}

	a.Logger.Printf("active profile changed to %s", profileID)
	return nil
}
