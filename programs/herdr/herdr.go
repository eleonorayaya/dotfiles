package herdr

import (
	"embed"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

var plugins = []string{
	"tdi/herdr-worktree-setup",
	"mrcndz/herdr-routines",
	"lmilojevicc/herdr-splits.nvim",
	"persiyanov/herdr-reviewr",
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "herdr"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("herdr", false); err != nil {
		return fmt.Errorf("failed to install herdr: %w", err)
	}

	if err := installClaudeIntegration(); err != nil {
		return fmt.Errorf("failed to install claude integration: %w", err)
	}

	for _, plugin := range plugins {
		if err := installPlugin(plugin); err != nil {
			return fmt.Errorf("failed to install plugin %s: %w", plugin, err)
		}
	}

	return nil
}

func installClaudeIntegration() error {
	cmd := exec.Command("herdr", "integration", "install", "claude")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func installPlugin(repo string) error {
	cmd := exec.Command("herdr", "plugin", "install", repo, "--yes")
	output, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	if strings.Contains(string(output), "already installed") {
		slog.Debug("herdr plugin already installed, skipping", "plugin", repo)
		return nil
	}

	return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	colors := ctx.Styles.Theme.Colors
	data := map[string]any{
		"Surface":       colors.Surface,
		"TextOnSurface": colors.TextOnSurface,
		"Primary":       colors.Primary,
		"Error":         colors.Error,
		"AccentMint":    colors.AccentMint,
		"AccentBlue":    colors.AccentBlue,
		"AccentGold":    colors.AccentGold,
		"AccentPurple":  colors.AccentPurple,
		"AccentPeach":   colors.AccentPeach,
	}

	fileMap, genErr := app.GenerateAppFiles("herdr", contents, data, ctx.OutDir)
	if genErr != nil {
		return nil, fmt.Errorf("failed to generate herdr files: %w", genErr)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/herdr/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync herdr files: %w", err)
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		AllowedBashCommands: []string{
			"herdr status:*",
			"herdr api snapshot:*",
			"herdr config check:*",
			"herdr workspace list:*",
			"herdr worktree list:*",
			"herdr agent list:*",
			"herdr pane list:*",
			"herdr tab list:*",
			"herdr plugin list:*",
		},
		SandboxAllowWrite: []string{
			"~/.config/herdr",
			"~/herdr-sessions",
		},
		SandboxAllowedDomains: []string{
			"herdr.dev",
		},
	}
}
