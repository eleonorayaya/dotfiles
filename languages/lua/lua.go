package lua

import (
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "lua"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageLua)
}

func (a *App) AgentConfig() shizukuapp.AgentConfig {
	return shizukuapp.AgentConfig{
		Plugins: []string{
			"lua-lsp@claude-plugins-official",
		},
	}
}
