package kitty

import (
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "kitty"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if util.BinaryExists("kitty") {
		slog.Info("kitty already installed, skipping")
		return nil
	}

	slog.Debug("installing kitty via curl script")

	cmd := exec.Command("sh", "-c", "curl -L https://sw.kovidgoyal.net/kitty/installer.sh | sh /dev/stdin")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install kitty: %w\nOutput: %s", err, string(output))
	}

	slog.Debug("kitty installed successfully")

	return nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"Styles": config.Styles,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("kitty", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/kitty/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
