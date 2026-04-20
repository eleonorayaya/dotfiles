package bat

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
	return "bat"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("bat", false); err != nil {
		return fmt.Errorf("failed to install bat: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Aliases: []app.Alias{
			{Name: "cat", Command: "bat"},
		},
	}, nil
}
