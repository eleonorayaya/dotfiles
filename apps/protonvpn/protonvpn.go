package protonvpn

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallCask("protonvpn"); err != nil {
		return fmt.Errorf("failed to install protonvpn: %w", err)
	}

	return nil
}
