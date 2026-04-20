package aerospace

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
	return "aerospace"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("nikitabobko/tap"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("aerospace", true); err != nil {
		return fmt.Errorf("failed to install aerospace: %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	data := map[string]any{
		"Gaps": ctx.Styles.Gaps,
	}

	fileMap, err := app.GenerateAppFiles("aerospace", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/aerospace/",
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
