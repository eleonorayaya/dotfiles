package fastfetch

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
	return "fastfetch"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("fastfetch", false); err != nil {
		return fmt.Errorf("failed to install fastfetch: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	data := map[string]any{
		"Colors": config.Styles.Theme.Colors,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("fastfetch", data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/fastfetch/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		PostInitScripts: []string{"fastfetch"},
	}, nil
}
