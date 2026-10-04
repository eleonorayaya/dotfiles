package helix

import (
	"embed"
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "helix"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("helix", false); err != nil {
		return fmt.Errorf("failed to install helix: %w", err)
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		SandboxAllowWrite: []string{
			"~/.cache/helix/",
		},
	}
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("helix", contents, map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/helix/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Variables: []app.EnvVar{
			{Key: "EDITOR", Value: "hx"},
			{Key: "VISUAL", Value: "hx"},
		},
	}, nil
}
