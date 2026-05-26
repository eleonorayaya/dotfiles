package ghdash

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
	return "gh-dash"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("dlvhdr/gh-dash"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("dlvhdr/gh-dash/gh-dash", false); err != nil {
		return fmt.Errorf("failed to install gh-dash: %w", err)
	}

	return nil
}
