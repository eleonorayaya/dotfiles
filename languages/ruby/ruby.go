package ruby

import (
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "ruby"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageRuby)
}

func (a *App) AgentConfig() shizukuapp.AgentConfig {
	return shizukuapp.AgentConfig{
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
