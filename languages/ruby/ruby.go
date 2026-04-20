package ruby

import (
	"github.com/eleonorayaya/shizuku/app"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "ruby"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"ruby-lsp@claude-plugins-official",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
		},
		SandboxAllowWrite: []string{
			"~/.bundle",
			"~/.gem",
			"~/.rbenv",
			"~/.cache/bundler",
			"~/.cache/rubygems",
			"~/Library/Caches/bundle",
		},
	}
}
