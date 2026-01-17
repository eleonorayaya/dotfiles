package aerospace

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
)

func Sync(outDir string) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("aerospace", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/aerospace/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
