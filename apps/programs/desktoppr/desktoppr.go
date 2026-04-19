package desktoppr

import (
	"embed"
	"fmt"
	"os/exec"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const wallpaperPath = "~/.config/desktoppr/cozy-autumn-rain.png"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "desktoppr"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("desktoppr", true); err != nil {
		return fmt.Errorf("failed to install desktoppr: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("desktoppr", contents, nil, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate desktoppr files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/desktoppr/",
	}, nil
}

func (a *App) Sync(outDir string, cfg *config.Config) error {
	result, err := a.Generate(outDir, cfg)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync desktoppr files: %w", err)
	}

	expandedPath, err := util.NormalizeFilePath(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to expand wallpaper path: %w", err)
	}

	cmd := exec.Command("desktoppr", expandedPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w\nOutput: %s", err, output)
	}

	return nil
}
