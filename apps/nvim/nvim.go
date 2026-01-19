package nvim

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/shizukustyle"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "nvim"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("neovim"); err != nil {
		return fmt.Errorf("failed to install neovim: %w", err)
	}

	return nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config, styles *shizukustyle.Styles) error {
	data := map[string]any{
		"ThemeName": styles.Theme.Name,
		"ThemeType": styles.Theme.Type,
		"Colors":    styles.Theme.Colors,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("nvim", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/nvim/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Variables: []shizukuapp.EnvVar{
			{Key: "EDITOR", Value: "nvim"},
		},
		Aliases: []shizukuapp.Alias{
			{Name: "vim", Command: "nvim"},
		},
	}, nil
}
