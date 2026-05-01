package datadog

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
	return "datadog"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		SandboxAllowedDomains: []string{
			"api.datadoghq.com",
			"app.datadoghq.com",
		},
		Marketplaces: map[string]app.Marketplace{
			"datadog-pup": {Repo: "datadog-labs/pup"},
		},
		Plugins: []string{"pup@datadog-pup"},
	}
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("datadog-labs/pack"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("datadog-labs/pack/pup", false); err != nil {
		return fmt.Errorf("failed to install pup: %w", err)
	}

	return nil
}
