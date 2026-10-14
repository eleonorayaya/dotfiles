package sketchybar

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/programs/sketchybar/bar"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const (
	configName      = "bar.json"
	configFile      = "~/.config/sketchybar/" + configName
	defaultUtenaURL = "http://localhost:3333"
)

var DefaultWorkspaceIcons = map[string]string{
	"1": "",
	"2": "",
	"3": "\U000F066F",
	"5": "\U000F0EBF",
	"7": "",
	"8": "\U000F0D8E",
	"T": "",
}

type Options struct {
	MailApp        string
	WorkspaceIcons map[string]string
	DisableUtena   bool
	DisableBT      bool
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	if opts.WorkspaceIcons == nil {
		opts.WorkspaceIcons = DefaultWorkspaceIcons
	}
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

	if !a.opts.DisableBT {
		if err := util.InstallBrewPackage("blueutil", false); err != nil {
			return fmt.Errorf("failed to install blueutil: %w", err)
		}
	}

	return nil
}

func (a *App) barConfig(ctx *app.Context) bar.Config {
	colors := ctx.Styles.Theme.Colors
	cfg := bar.Config{
		Bar: bar.BarStyle{
			Height:       32,
			Margin:       12,
			YOffset:      12,
			CornerRadius: 8,
			BorderWidth:  2,
			BlurRadius:   48,
			Color:        util.HexToARGB(colors.Surface, ctx.Styles.WindowOpacity),
			BorderColor:  util.HexToARGB(colors.SurfaceBorder, ctx.Styles.WindowOpacity),
		},
		Colors: bar.Colors{
			Icon:            util.HexToARGB(colors.TextOnSurface, 100),
			IconHighlight:   util.HexToARGB(colors.Primary, 100),
			Label:           util.HexToARGB(colors.TextOnSurface, 100),
			LabelHighlight:  util.HexToARGB(colors.Primary, 100),
			PopupBackground: util.HexToARGB(colors.Surface, 100),
			PopupBorder:     util.HexToARGB(colors.SurfaceBorder, 100),
		},
		Fonts: bar.Fonts{
			Text: "SF Pro Display",
			Icon: "Liga SFMono Nerd Font",
		},
		WorkspaceIcons: a.opts.WorkspaceIcons,
		MailApp:        a.opts.MailApp,
		Bluetooth:      !a.opts.DisableBT,
	}
	if !a.opts.DisableUtena {
		cfg.UtenaURL = defaultUtenaURL
	}
	return cfg
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("sketchybar", contents, nil, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	data, err := json.MarshalIndent(a.barConfig(ctx), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to encode bar config: %w", err)
	}
	configPath := path.Join(ctx.OutDir, "sketchybar", configName)
	if err := os.WriteFile(configPath, append(data, '\n'), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write bar config: %w", err)
	}
	fileMap[configName] = configPath

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

	if util.BinaryExists("sketchybar") {
		if out, err := exec.Command("sketchybar", "--reload").CombinedOutput(); err != nil {
			slog.Warn("failed to reload sketchybar", "error", err, "output", string(out))
		}
	}

	return nil
}
