package utena

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

const (
	repoURL  = "https://github.com/eleonorayaya/utena"
	cloneDir = ".local/src/utena"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "utena"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	repoPath := filepath.Join(homeDir, cloneDir)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		cmd := exec.Command("git", "clone", repoURL, repoPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to clone utena: %s: %w", string(output), err)
		}
	} else {
		cmd := exec.Command("git", "-C", repoPath, "pull", "origin", "main")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to pull latest utena: %s: %w", string(output), err)
		}
	}

	taskCmd := exec.Command("task", "install")
	taskCmd.Dir = repoPath
	if output, err := taskCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install utena: %s: %w", string(output), err)
	}

	if err := updateClaudePlugin(); err != nil {
		return fmt.Errorf("failed to update claude plugin: %w", err)
	}

	return nil
}

func updateClaudePlugin() error {
	marketplaceCmd := exec.Command("claude", "plugin", "marketplace", "update", "utena")
	if output, err := marketplaceCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("marketplace update failed: %s: %w", string(output), err)
	}

	pluginCmd := exec.Command("claude", "plugin", "update", "utena-claude@utena")
	if output, err := pluginCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("plugin update failed: %s: %w", string(output), err)
	}

	return nil
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	colors := config.Styles.Theme.Colors
	data := map[string]any{
		"Primary":               colors.Primary,
		"PrimaryVariant":        colors.PrimaryVariant,
		"Secondary":             colors.Secondary,
		"Tertiary":              colors.Tertiary,
		"SurfaceVariant":        colors.SurfaceVariant,
		"SurfaceHighlight":      colors.SurfaceHighlight,
		"Selection":             colors.Selection,
		"TextOnSurface":         colors.TextOnSurface,
		"TextOnSurfaceEmphasis": colors.TextOnSurfaceEmphasis,
		"TextOnSurfaceMuted":    colors.TextOnSurfaceMuted,
		"TextOnPrimary":         colors.TextOnPrimary,
		"AccentBlue":            colors.AccentBlue,
		"AccentLavender":        colors.AccentLavender,
		"AccentMint":            colors.AccentMint,
		"AccentGold":            colors.AccentGold,
		"Error":                 colors.Error,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("utena", data, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate utena files: %w", err)
	}

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/utena/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync utena files: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	scriptPath := filepath.Join(homeDir, ".config", "utena", "worktree-setup")
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make worktree-setup executable: %w", err)
	}

	return nil
}
