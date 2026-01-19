package aerospace

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
	return "aerospace"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("nikitabobko/tap"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("aerospace", true); err != nil {
		return fmt.Errorf("failed to install aerospace: %w", err)
	}

	return nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := shizukuapp.GenerateAppFiles("aerospace", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/aerospace/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
