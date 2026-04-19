package fastfetch

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
	return "fastfetch"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("fastfetch", false); err != nil {
		return fmt.Errorf("failed to install fastfetch: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
	data := map[string]any{
		"Colors": cfg.Styles.Theme.Colors,
	}

	fileMap, err := app.GenerateAppFiles("fastfetch", contents, data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/fastfetch/",
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
		PostInitScripts: []string{"fastfetch"},
	}, nil
}
