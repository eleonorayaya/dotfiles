package kitty

import (
	"embed"
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

type Options struct {
	OpenURLWith string
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "kitty"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("font-maple-mono", true); err != nil {
		return fmt.Errorf("failed to install font-maple-mono: %w", err)
	}

	if err := util.InstallBrewPackage("kitty", true); err != nil {
		return fmt.Errorf("failed to install kitty (if it was installed outside brew, run `brew install --cask --force kitty`): %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	openURLWith := a.opts.OpenURLWith
	if openURLWith == "" {
		openURLWith = "default"
	}
	data := map[string]any{
		"OpenURLWith":     openURLWith,
		"Styles":          ctx.Styles,
		"BackgroundAlpha": float64(ctx.Styles.WindowOpacity) / 100.0,
		"Background":      ctx.Styles.TerminalBackground(),
	}

	fileMap, err := app.GenerateAppFiles("kitty", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/kitty/",
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
