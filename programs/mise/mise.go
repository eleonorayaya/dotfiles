package mise

import (
	"github.com/eleonorayaya/shizuku/app"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "mise"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		SandboxAllowedHosts: []string{
			"mise.jdx.dev",
			"mise-versions.jdx.dev",
			"hk.jdx.dev",
		},
		SandboxAllowWrite: []string{
			"~/.cache/mise",
			"~/.config/mise",
			"~/.local/share/mise",
			"~/.local/state/mise",
			"~/Library/Caches/mise",
		},
	}
}
