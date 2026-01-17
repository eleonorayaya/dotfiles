package kitty

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("kitty", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/kitty/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
