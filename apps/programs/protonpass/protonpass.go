package protonpass

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
	return "protonpass"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", false)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("proton-pass", true); err != nil {
		return fmt.Errorf("failed to install proton-pass: %w", err)
	}

	return nil
}
