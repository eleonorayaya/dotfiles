package tmux

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

const tpmPath = "~/.config/tmux/plugins/tpm"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "tmux"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("tmux", false); err != nil {
		return fmt.Errorf("failed to install tmux: %w", err)
	}

	if err := ensureTPM(); err != nil {
		return fmt.Errorf("failed to ensure TPM: %w", err)
	}

	return nil
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	colors := config.Styles.Theme.Colors
	data := map[string]any{
		"Surface":              colors.Surface,
		"SurfaceVariant":       colors.SurfaceVariant,
		"SurfaceHighlight":     colors.SurfaceHighlight,
		"SurfaceBorder":        colors.SurfaceBorder,
		"TextOnSurface":        colors.TextOnSurface,
		"TextOnSurfaceVariant": colors.TextOnSurfaceVariant,
		"TextOnSurfaceMuted":   colors.TextOnSurfaceMuted,
		"Primary":              colors.Primary,
		"TextOnPrimary":        colors.TextOnPrimary,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("tmux", data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/tmux/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func ensureTPM() error {
	expandedPath, err := util.NormalizeFilePath(tpmPath)
	if err != nil {
		return fmt.Errorf("failed to expand TPM path: %w", err)
	}

	if _, err := os.Stat(expandedPath); err == nil {
		slog.Debug("TPM already installed, skipping")
		return nil
	}

	slog.Debug("cloning TPM")

	cmd := exec.Command("git", "clone", "https://github.com/tmux-plugins/tpm", expandedPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone TPM: %w\nOutput: %s", err, string(output))
	}

	slog.Debug("TPM installed successfully")
	return nil
}
