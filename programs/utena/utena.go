package utena

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
)

//go:embed all:contents
var contents embed.FS

const (
	repoURL  = "https://github.com/eleonorayaya/utena"
	cloneDir = ".local/src/utena"
)

type Options struct {
	Browser string
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "utena"
}

func (a *App) Install(ctx *app.Context) error {
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

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	colors := ctx.Styles.Theme.Colors
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
		"TextOnSurfaceVariant":  colors.TextOnSurfaceVariant,
		"TextOnSurfaceMuted":    colors.TextOnSurfaceMuted,
		"TextOnPrimary":         colors.TextOnPrimary,
		"Blue":                  colors.Palette.Blue,
		"BrightMagenta":         colors.Palette.BrightMagenta,
		"Green":                 colors.Palette.Green,
		"BrightYellow":          colors.Palette.BrightYellow,
		"Path":                  colors.Path,
		"Error":                 colors.Error,
		"Browser":               a.opts.Browser,
	}

	fileMap, err := app.GenerateAppFiles("utena", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate utena files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/utena/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync utena files: %w", err)
	}

	if err := configureGitHook(); err != nil {
		return fmt.Errorf("failed to configure utena git hook: %w", err)
	}

	return nil
}

var gitHookConfig = [][]string{
	{"config", "--global", "--replace-all", "hook.utena.command", `utena hook post-checkout "$@" >/dev/null 2>&1 || :`},
	{"config", "--global", "--replace-all", "hook.utena.event", "post-checkout"},
}

func configureGitHook() error {
	for _, args := range gitHookConfig {
		cmd := exec.Command("git", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set git hook config %v: %s: %w", args, string(output), err)
		}
	}
	return nil
}
