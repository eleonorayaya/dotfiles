package sfsymbols

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
	return "sfsymbols"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("sf-symbols", true); err != nil {
		return fmt.Errorf("failed to install sf-symbols: %w", err)
	}

	return nil
}
