package sketchybar

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/theme"
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

func (a *App) Sync(outDir string, config *shizukuconfig.Config, themeData *theme.Theme) error {
	data := map[string]any{
		"BarColor":                hexToARGB(themeData.Colors.Surface, "D9"),
		"BarBorderColor":          hexToARGB(themeData.Colors.SurfaceBorder, "FF"),
		"IconColor":               hexToARGB(themeData.Colors.TextOnSurface, "FF"),
		"IconHighlightColor":      hexToARGB(themeData.Colors.Primary, "FF"),
		"LabelColor":              hexToARGB(themeData.Colors.TextOnSurface, "FF"),
		"LabelHighlightColor":     hexToARGB(themeData.Colors.Primary, "FF"),
		"PopupBorderColor":        hexToARGB(themeData.Colors.SurfaceBorder, "FF"),
		"PopupBackgroundColor":    hexToARGB(themeData.Colors.Surface, "FF"),
		"ActiveWorkspaceColor":    hexToARGB(themeData.Colors.Primary, "FF"),
		"SpacesWrapperBackground": hexToARGB(themeData.Colors.Surface, "FF"),
		"SpacesItemBackground":    hexToARGB(themeData.Colors.Primary, "FF"),
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
