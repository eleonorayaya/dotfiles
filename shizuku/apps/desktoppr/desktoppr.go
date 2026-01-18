package desktoppr

import (
	"fmt"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

const wallpaperPath = "~/.config/desktoppr/cozy-autumn-rain.png"

func Sync(outDir string, config *shizukuconfig.Config) error {
	fileMap, err := internal.GenerateAppFiles("desktoppr", nil, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate desktoppr files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/desktoppr/"); err != nil {
		return fmt.Errorf("failed to sync desktoppr files: %w", err)
	}

	expandedPath, err := util.NormalizeFilePath(wallpaperPath)
	if err != nil {
		return fmt.Errorf("failed to expand wallpaper path: %w", err)
	}

	cmd := exec.Command("desktoppr", expandedPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w\nOutput: %s", err, output)
	}

	return nil
}
