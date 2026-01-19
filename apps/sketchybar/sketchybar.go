package sketchybar

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
	return "sketchybar"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.AddTap("felixkratz/formulae"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("felixkratz/formulae/sketchybar"); err != nil {
		return fmt.Errorf("failed to install sketchybar: %w", err)
	}

	return nil
}

func hexToARGB(hex string, alpha string) string {
	hex = hex[1:]
	return "0x" + alpha + hex
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"BarColor":                hexToARGB(config.Styles.Theme.Colors.Surface, "D9"),
		"BarBorderColor":          hexToARGB(config.Styles.Theme.Colors.SurfaceBorder, "FF"),
		"IconColor":               hexToARGB(config.Styles.Theme.Colors.TextOnSurface, "FF"),
		"IconHighlightColor":      hexToARGB(config.Styles.Theme.Colors.Primary, "FF"),
		"LabelColor":              hexToARGB(config.Styles.Theme.Colors.TextOnSurface, "FF"),
		"LabelHighlightColor":     hexToARGB(config.Styles.Theme.Colors.Primary, "FF"),
		"PopupBorderColor":        hexToARGB(config.Styles.Theme.Colors.SurfaceBorder, "FF"),
		"PopupBackgroundColor":    hexToARGB(config.Styles.Theme.Colors.Surface, "FF"),
		"ActiveWorkspaceColor":    hexToARGB(config.Styles.Theme.Colors.Primary, "FF"),
		"SpacesWrapperBackground": hexToARGB(config.Styles.Theme.Colors.Surface, "FF"),
		"SpacesItemBackground":    hexToARGB(config.Styles.Theme.Colors.Primary, "FF"),
	}

	fileMap, err := shizukuapp.GenerateAppFiles("sketchybar", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/sketchybar/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
