package lua

import (
	"github.com/eleonorayaya/shizuku/app"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "lua"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"lua-lsp@claude-plugins-official",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
		},
	}
}
