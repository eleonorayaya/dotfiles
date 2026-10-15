package firefox

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
)

//go:embed all:contents
var contents embed.FS

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "firefox"
}

func (a *App) Install(ctx *app.Context) error {
	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve home dir: %w", err)
	}

	profileDir, err := defaultProfileDir(filepath.Join(home, "Library", "Application Support", "Firefox"))
	if err != nil {
		return nil, fmt.Errorf("failed to find firefox profile: %w", err)
	}
	if profileDir == "" {
		slog.Warn("no firefox profile found; open firefox once, then sync again")
		return &app.GenerateResult{FileMap: map[string]string{}}, nil
	}

	data := map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}

	fileMap, err := app.GenerateAppFiles("firefox", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate firefox files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: profileDir,
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync firefox files: %w", err)
	}

	if len(result.FileMap) > 0 {
		slog.Info("firefox theme synced; restart firefox to apply it", "profile", result.DestDir)
	}

	return nil
}
