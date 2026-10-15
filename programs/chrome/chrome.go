package chrome

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

const themeDir = "~/.config/chrome-theme/"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "chrome"
}

func (a *App) Install(ctx *app.Context) error {
	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	appOutDir := filepath.Join(ctx.OutDir, "chrome")

	images, err := patternImages(ctx.Styles.Theme.Colors)
	if err != nil {
		return nil, fmt.Errorf("failed to draw new tab pattern: %w", err)
	}

	manifest, err := themeManifest(ctx.Styles.Theme, images)
	if err != nil {
		return nil, fmt.Errorf("failed to build chrome theme: %w", err)
	}

	manifestPath := filepath.Join(appOutDir, "manifest.json")
	if err := util.WriteJSON(manifestPath, manifest); err != nil {
		return nil, fmt.Errorf("failed to write chrome theme manifest: %w", err)
	}

	fileMap := map[string]string{"manifest.json": manifestPath}
	for rel, data := range images {
		path := filepath.Join(appOutDir, rel)
		if err := util.EnsureDirExists(filepath.Dir(path)); err != nil {
			return nil, fmt.Errorf("failed to create images dir: %w", err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", rel, err)
		}
		fileMap[rel] = path
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: themeDir,
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync chrome theme: %w", err)
	}

	slog.Info("chrome theme synced; apply changes via chrome://extensions > Load unpacked", "path", themeDir)

	return nil
}
