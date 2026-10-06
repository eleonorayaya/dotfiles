package tuios

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
)

//go:embed all:contents
var contents embed.FS

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "tuios"
}

var spec = pkg.Spec{
	Brew: "tuios",
	Bin:  "tuios",
	Release: &pkg.Release{
		Repo:  "Gaurav-Gosain/tuios",
		Asset: `tuios_[0-9.]+_Linux_{arch}\.tar\.gz`,
		Arch:  map[string]string{"amd64": "x86_64", "arm64": "arm64"},
	},
}

func (a *App) Install(ctx *app.Context) error {
	if err := pkg.Install(spec); err != nil {
		return fmt.Errorf("failed to install tuios: %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("tuios", contents, map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user config dir: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: filepath.Join(configDir, "tuios"),
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
