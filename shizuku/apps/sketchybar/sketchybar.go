package sketchybar

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
)

func Sync(outDir string) error {
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
