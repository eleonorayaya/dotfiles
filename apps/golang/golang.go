package golang

import "github.com/eleonorayaya/shizuku/internal/shizukuapp"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		PathDirs: []shizukuapp.PathDir{
			{Path: "$HOME/go/bin", Priority: 20},
		},
	}, nil
}
