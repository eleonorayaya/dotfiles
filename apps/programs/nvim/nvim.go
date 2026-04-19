package nvim

import (
	"embed"
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "nvim"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("neovim", false); err != nil {
		return fmt.Errorf("failed to install neovim: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
	data := map[string]any{
		"ThemeName": cfg.Styles.Theme.Name,
		"ThemeType": cfg.Styles.Theme.Type,
		"Colors":    cfg.Styles.Theme.Colors,
	}

	fileMap, err := app.GenerateAppFiles("nvim", contents, data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/nvim/",
	}, nil
}

func (a *App) Sync(outDir string, cfg *config.Config) error {
	result, err := a.Generate(outDir, cfg)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Variables: []app.EnvVar{
			{Key: "EDITOR", Value: "nvim"},
		},
		Aliases: []app.Alias{
			{Name: "vim", Command: "nvim"},
		},
	}, nil
}
