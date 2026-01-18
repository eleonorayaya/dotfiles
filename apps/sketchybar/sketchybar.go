package sketchybar

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"Test": "Aayaya",
	}

	fileMap, err := shizukuapp.GenerateAppFiles("sketchybar", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/sketchybar/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
