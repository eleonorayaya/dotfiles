package buildkite

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "buildkite"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		SandboxAllowedDomains: []string{
			"api.buildkite.com",
			"buildkite.com",
		},
	}
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("buildkite/buildkite"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("buildkite/buildkite/bk@3", false); err != nil {
		return fmt.Errorf("failed to install bk: %w", err)
	}

	return nil
}
