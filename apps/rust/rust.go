package rust

import (
	"path"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	rustupPrefix, err := util.GetBrewAppPrefix("rustup")
	if err != nil {
		return nil, err
	}

	rustupBin := path.Join(rustupPrefix, "bin")

	return &shizukuapp.EnvSetup{
		PathDirs: []shizukuapp.PathDir{
			{Path: rustupBin, Priority: 10},
			{Path: "$HOME/.cargo/bin", Priority: 10},
		},
		Variables: []shizukuapp.EnvVar{
			{Key: "RUSTUP_HOME", Value: "$HOME/.rustup"},
		},
	}, nil
}
