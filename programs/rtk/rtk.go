package rtk

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
	return "rtk"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("rtk", false); err != nil {
		return fmt.Errorf("failed to install rtk: %w", err)
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Hooks: []app.Hook{
			{
				Event:   "PreToolUse",
				Matcher: "Bash",
				Command: "rtk hook claude",
			},
		},
	}
}
