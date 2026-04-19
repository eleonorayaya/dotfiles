package zig

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "zig"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetLanguageEnabled(config.LanguageZig)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("zig", false); err != nil {
		return fmt.Errorf("failed to install zig: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{}, nil
}
