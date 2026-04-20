package protonvpn

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "protonvpn"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("protonvpn", true); err != nil {
		return fmt.Errorf("failed to install protonvpn: %w", err)
	}

	return nil
}
