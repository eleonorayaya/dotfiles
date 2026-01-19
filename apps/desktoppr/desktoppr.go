package desktoppr

import (
	"fmt"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

const wallpaperPath = "~/.config/desktoppr/cozy-autumn-rain.png"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallCask("desktoppr"); err != nil {
		return fmt.Errorf("failed to install desktoppr: %w", err)
	}

	return nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	fileMap, err := shizukuapp.GenerateAppFiles("desktoppr", nil, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate desktoppr files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/desktoppr/"); err != nil {
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
