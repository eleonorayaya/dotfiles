package protonvpn

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
	return "protonvpn"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", false)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("protonvpn", true); err != nil {
		return fmt.Errorf("failed to install protonvpn: %w", err)
	}

	return nil
}
