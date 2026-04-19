package aerospace

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
	return "aerospace"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.AddTap("nikitabobko/tap"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("aerospace", true); err != nil {
		return fmt.Errorf("failed to install aerospace: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
	laptop := cfg.GetAppConfigBool(a.Name(), "laptop", false)
	gaps := config.DefaultGapsForLaptop(laptop)

	data := map[string]any{
		"Gaps": gaps,
	}

	fileMap, err := app.GenerateAppFiles("aerospace", contents, data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/aerospace/",
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
