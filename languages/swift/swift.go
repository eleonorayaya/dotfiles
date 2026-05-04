package swift

import (
	"github.com/eleonorayaya/shizuku/app"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "swift"
}

func (a *App) Install(ctx *app.Context) error {
	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		AllowedCommands: []string{
			"Bash(swift build:*)",
		},
		SandboxAllowWrite: []string{
			"~/Library/Developer/",
		},
	}
}
