package terminalbrowser

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

const marketplaceName = "terminal-browser"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terminal-browser"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("terminal-browser", true); err != nil {
		return fmt.Errorf("failed to install terminal-browser: %w", err)
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Marketplaces: map[string]app.Marketplace{
			marketplaceName: {Repo: "zenbu-labs/terminal-browser"},
		},
		Plugins: []string{"terminal-browser@" + marketplaceName},
	}
}
