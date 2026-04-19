package python

import (
	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "python"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Aliases: []app.Alias{
			{Name: "python", Command: "python3"},
		},
	}, nil
}
