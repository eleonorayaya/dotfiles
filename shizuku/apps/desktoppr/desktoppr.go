package desktoppr

import (
	"fmt"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

const wallpaperPath = "~/.dotfiles/wallpaper/cozy-autumn-rain.png"

func Sync(outDir string, config *shizukuconfig.Config) error {
	cmd := exec.Command("desktoppr", wallpaperPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w\nOutput: %s", err, output)
	}

	return nil
}
