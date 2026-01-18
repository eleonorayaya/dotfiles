package terraform

import "github.com/eleonorayaya/shizuku/internal/shizukuapp"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Aliases: []shizukuapp.Alias{
			{Name: "tf", Command: "terraform"},
			{Name: "tfmt", Command: "terraform fmt -recursive"},
		},
	}, nil
}
