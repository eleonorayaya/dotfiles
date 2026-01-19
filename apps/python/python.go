package python

import (
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "python"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return true
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Aliases: []shizukuapp.Alias{
			{Name: "python", Command: "python3"},
		},
	}, nil
}
