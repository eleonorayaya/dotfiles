package sketchybar

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"Test": "Aayaya",
	}

	fileMap, err := internal.GenerateAppFiles("sketchybar", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/sketchybar/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
