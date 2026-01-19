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

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"BarColor":                util.HexToARGB(config.Styles.Theme.Colors.Surface, config.Styles.WindowOpacity),
		"BarBorderColor":          util.HexToARGB(config.Styles.Theme.Colors.SurfaceBorder, config.Styles.WindowOpacity),
		"IconColor":               util.HexToARGB(config.Styles.Theme.Colors.TextOnSurface, 100),
		"IconHighlightColor":      util.HexToARGB(config.Styles.Theme.Colors.Primary, 100),
		"LabelColor":              util.HexToARGB(config.Styles.Theme.Colors.TextOnSurface, 100),
		"LabelHighlightColor":     util.HexToARGB(config.Styles.Theme.Colors.Primary, 100),
		"PopupBorderColor":        util.HexToARGB(config.Styles.Theme.Colors.SurfaceBorder, 100),
		"PopupBackgroundColor":    util.HexToARGB(config.Styles.Theme.Colors.Surface, 100),
		"ActiveWorkspaceColor":    util.HexToARGB(config.Styles.Theme.Colors.Primary, 100),
		"SpacesWrapperBackground": util.HexToARGB(config.Styles.Theme.Colors.Surface, 100),
		"SpacesItemBackground":    util.HexToARGB(config.Styles.Theme.Colors.Primary, 100),
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
