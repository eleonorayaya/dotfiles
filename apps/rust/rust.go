package rust

import (
	"fmt"
	"path"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "rust"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return true
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("rustup"); err != nil {
		return fmt.Errorf("failed to install rustup: %w", err)
	}

	return nil
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
