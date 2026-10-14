package sketchybar

import (
	"embed"
	"fmt"
	"log/slog"
	"os/exec"
	"path"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/programs/sketchybar/bar"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const (
	configDir  = "~/.config/sketchybar/"
	configName = "bar.json"
)

type Options struct {
	MailApp string
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
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

	if err := util.InstallBrewPackage("blueutil", false); err != nil {
		return fmt.Errorf("failed to install blueutil: %w", err)
	}

	return nil
}

func (a *App) barConfig(ctx *app.Context) bar.Config {
	colors := ctx.Styles.Theme.Colors
	text := util.HexToARGB(colors.TextOnSurface, 100)
	primary := util.HexToARGB(colors.Primary, 100)
	return bar.Config{
		Colors: bar.Colors{
			Bar:             util.HexToARGB(colors.Surface, ctx.Styles.WindowOpacity),
			BarBorder:       util.HexToARGB(colors.SurfaceBorder, ctx.Styles.WindowOpacity),
			Icon:            text,
			IconHighlight:   primary,
			Label:           text,
			LabelHighlight:  primary,
			PopupBackground: util.HexToARGB(colors.Surface, 100),
			PopupBorder:     util.HexToARGB(colors.SurfaceBorder, 100),
		},
		MailApp: a.opts.MailApp,
	}
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("sketchybar", contents, nil, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	configPath := path.Join(ctx.OutDir, "sketchybar", configName)
	if err := util.WriteJSON(configPath, a.barConfig(ctx)); err != nil {
		return nil, fmt.Errorf("failed to write bar config: %w", err)
	}
	fileMap[configName] = configPath

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: configDir,
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

	if util.BinaryExists("sketchybar") {
		if out, err := exec.Command("sketchybar", "--reload").CombinedOutput(); err != nil {
			slog.Warn("failed to reload sketchybar", "error", err, "output", string(out))
		}
	}

	return nil
}
