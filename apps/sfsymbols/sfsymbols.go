package sfsymbols

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
	return "sfsymbols"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", false)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallCask("sf-symbols"); err != nil {
		return fmt.Errorf("failed to install sf-symbols: %w", err)
	}

	return nil
}
