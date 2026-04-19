package sketchybar

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
	return "sketchybar"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.AddTap("felixkratz/formulae"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("sketchybar", false); err != nil {
		return fmt.Errorf("failed to install sketchybar: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
	data := map[string]any{
		"BarColor":                util.HexToARGB(cfg.Styles.Theme.Colors.Surface, cfg.Styles.WindowOpacity),
		"BarBorderColor":          util.HexToARGB(cfg.Styles.Theme.Colors.SurfaceBorder, cfg.Styles.WindowOpacity),
		"IconColor":               util.HexToARGB(cfg.Styles.Theme.Colors.TextOnSurface, 100),
		"IconHighlightColor":      util.HexToARGB(cfg.Styles.Theme.Colors.Primary, 100),
		"LabelColor":              util.HexToARGB(cfg.Styles.Theme.Colors.TextOnSurface, 100),
		"LabelHighlightColor":     util.HexToARGB(cfg.Styles.Theme.Colors.Primary, 100),
		"PopupBorderColor":        util.HexToARGB(cfg.Styles.Theme.Colors.SurfaceBorder, 100),
		"PopupBackgroundColor":    util.HexToARGB(cfg.Styles.Theme.Colors.Surface, 100),
		"ActiveWorkspaceColor":    util.HexToARGB(cfg.Styles.Theme.Colors.Primary, 100),
		"SpacesWrapperBackground": util.HexToARGB(cfg.Styles.Theme.Colors.Surface, 100),
		"SpacesItemBackground":    util.HexToARGB(cfg.Styles.Theme.Colors.Primary, 100),
	}

	fileMap, err := app.GenerateAppFiles("sketchybar", contents, data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/sketchybar/",
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
