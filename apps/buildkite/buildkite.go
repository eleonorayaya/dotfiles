package buildkite

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "buildkite"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", false)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("buildkite/buildkite"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("buildkite/buildkite/bk@3", false); err != nil {
		return fmt.Errorf("failed to install bk: %w", err)
	}

	return nil
}
