package buildkite

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "buildkite"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", false)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.AddTap("buildkite/buildkite"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("buildkite/buildkite/bk@3", false); err != nil {
		return fmt.Errorf("failed to install bk: %w", err)
	}

	return nil
}
