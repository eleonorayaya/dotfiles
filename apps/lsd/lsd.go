package lsd

import "github.com/eleonorayaya/shizuku/internal/shizukuapp"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Aliases: []shizukuapp.Alias{
			{Name: "ls", Command: "lsd"},
		},
	}, nil
}
