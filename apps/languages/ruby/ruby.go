package ruby

import (
	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "ruby"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetLanguageEnabled(config.LanguageRuby)
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"ruby-lsp@claude-plugins-official",
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
