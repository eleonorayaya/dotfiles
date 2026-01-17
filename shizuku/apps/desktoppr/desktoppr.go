package desktoppr

import (
	"fmt"
	"os/exec"
)

const wallpaperPath = "~/.dotfiles/wallpaper/cozy-autumn-rain.png"

func Sync(outDir string) error {
	cmd := exec.Command("desktoppr", wallpaperPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set wallpaper: %w\nOutput: %s", err, output)
	}

	return nil
}

