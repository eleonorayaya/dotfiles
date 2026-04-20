package protonpass

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
	return "protonpass"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("proton-pass", true); err != nil {
		return fmt.Errorf("failed to install proton-pass: %w", err)
	}

	return nil
}
