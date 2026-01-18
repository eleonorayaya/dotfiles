package fastfetch

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/shizukuenv"
)

func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("fastfetch", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/fastfetch/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		PostInitScripts: []string{"fastfetch"},
	}, nil
}
