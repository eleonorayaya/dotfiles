package golang

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "golang"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return true
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("go-task"); err != nil {
		return fmt.Errorf("failed to install go-task: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		PathDirs: []shizukuapp.PathDir{
			{Path: "$HOME/go/bin", Priority: 20},
		},
	}, nil
}
