package sketchybar

import (
	"embed"
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
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

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("felixkratz/formulae"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("sketchybar", false); err != nil {
		return fmt.Errorf("failed to install sketchybar: %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	data := map[string]any{
		"BarColor":                util.HexToARGB(ctx.Styles.Theme.Colors.Surface, ctx.Styles.WindowOpacity),
		"BarBorderColor":          util.HexToARGB(ctx.Styles.Theme.Colors.SurfaceBorder, ctx.Styles.WindowOpacity),
		"IconColor":               util.HexToARGB(ctx.Styles.Theme.Colors.TextOnSurface, 100),
		"IconHighlightColor":      util.HexToARGB(ctx.Styles.Theme.Colors.Primary, 100),
		"LabelColor":              util.HexToARGB(ctx.Styles.Theme.Colors.TextOnSurface, 100),
		"LabelHighlightColor":     util.HexToARGB(ctx.Styles.Theme.Colors.Primary, 100),
		"PopupBorderColor":        util.HexToARGB(ctx.Styles.Theme.Colors.SurfaceBorder, 100),
		"PopupBackgroundColor":    util.HexToARGB(ctx.Styles.Theme.Colors.Surface, 100),
		"ActiveWorkspaceColor":    util.HexToARGB(ctx.Styles.Theme.Colors.Primary, 100),
		"SpacesWrapperBackground": util.HexToARGB(ctx.Styles.Theme.Colors.Surface, 100),
		"SpacesItemBackground":    util.HexToARGB(ctx.Styles.Theme.Colors.Primary, 100),
	}

	fileMap, err := app.GenerateAppFiles("sketchybar", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/sketchybar/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
