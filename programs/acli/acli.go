package acli

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
	return "acli"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		AllowedBashCommands: []string{
			"acli --version:*",
			"acli jira workitem search:*",
			"acli jira workitem view:*",
			"acli jira workitem attachment list:*",
			"acli jira workitem comment list:*",
			"acli jira workitem comment visibility:*",
			"acli jira workitem link list:*",
			"acli jira workitem link type:*",
			"acli jira workitem watcher list:*",
		},
		SandboxAllowedDomains: []string{
			"api.atlassian.com",
			"as.atlassian.com",
			"ingest.us.sentry.io",
		},
		SandboxAllowRead: []string{
			"~/.config/acli",
		},
		SandboxAllowWrite: []string{
			"~/.config/acli",
		},
	}
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.AddTap("atlassian/acli"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("atlassian/acli/acli", false); err != nil {
		return fmt.Errorf("failed to install acli: %w", err)
	}

	return nil
}
