package kitty

import (
	"embed"
	"fmt"
	"log/slog"
	"os/exec"

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
	return "kitty"
}

func (a *App) Install(ctx *app.Context) error {
	if util.BinaryExists("kitty") {
		slog.Info("kitty already installed, skipping")
		return nil
	}

	slog.Debug("installing kitty via curl script")

	cmd := exec.Command("sh", "-c", "curl -L https://sw.kovidgoyal.net/kitty/installer.sh | sh /dev/stdin")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install kitty: %w\nOutput: %s", err, string(output))
	}

	slog.Debug("kitty installed successfully")

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	data := map[string]any{
		"Styles":          ctx.Styles,
		"BackgroundAlpha": float64(ctx.Styles.WindowOpacity) / 100.0,
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
