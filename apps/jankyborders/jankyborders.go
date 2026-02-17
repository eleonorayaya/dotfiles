package jankyborders

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
	return "jankyborders"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("felixkratz/formulae"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("felixkratz/formulae/borders", false); err != nil {
		return fmt.Errorf("failed to install borders: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	data := map[string]any{
		"ActiveColor":   util.HexToARGB(config.Styles.Theme.Colors.AccentLavender, 100),
		"InactiveColor": util.HexToARGB(config.Styles.Theme.Colors.AccentLavender, 100),
	}

	fileMap, err := shizukuapp.GenerateAppFiles("jankyborders", data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/borders/",
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
