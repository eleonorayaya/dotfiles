package notion

import "github.com/eleonorayaya/shizuku/app"

const mcpServerName = "notion"

type Options struct {
	DisableClaudeMCP bool
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "notion"
}

func (a *App) AgentConfig() app.AgentConfig {
	cfg := app.AgentConfig{}
	if a.opts.DisableClaudeMCP {
		cfg.DisabledMcpJsonServers = []string{mcpServerName}
	}
	return cfg
}
