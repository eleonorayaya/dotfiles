package zig

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
	return "zig"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageZig)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("tristanisham/zvm"); err != nil {
		return fmt.Errorf("failed to add zvm tap: %w", err)
	}

	if err := util.InstallBrewPackage("zvm", false); err != nil {
		return fmt.Errorf("failed to install zvm: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Variables: []shizukuapp.EnvVar{
			{Key: "ZVM_INSTALL", Value: "$HOME/.zvm/self"},
		},
		PathDirs: []shizukuapp.PathDir{
			{Path: "$HOME/.zvm/bin", Priority: 20},
			{Path: "$HOME/.zvm/self", Priority: 20},
		},
	}, nil
}
