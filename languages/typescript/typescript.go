package typescript

import (
	"github.com/eleonorayaya/shizuku/app"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "typescript"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"typescript-lsp@claude-plugins-official",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
		},
		AllowedCommands: []string{
			"Bash(npm install)",
			"Bash(npx nx test:*)",
			"Bash(npx nx sync:*)",
		},
		SandboxAllowedHosts: []string{
			"registry.npmjs.org",
		},
		SandboxAllowWrite: []string{
			"~/.npm",
			"~/.node-gyp",
			"~/.cache/node-gyp",
			"~/.cache/npm",
			"~/.cache/node",
			"~/.cache/yarn",
			"~/.cache/node/corepack",
			"~/.config/npm",
			"~/.config/configstore",
			"~/.config/yarn",
			"~/.config/pnpm",
			"~/.pnpm-state",
			"~/.pnpm-store",
			"~/.yarn",
			"~/.yarnrc",
			"~/.yarnrc.yml",
			"~/.local/share/pnpm",
			"~/.local/state/pnpm",
			"~/Library/Caches/npm",
			"~/Library/Caches/Yarn",
			"~/Library/Caches/node/corepack",
			"~/Library/Caches/pnpm",
			"~/Library/Preferences/pnpm",
			"~/Library/pnpm",
			"~/Library/Caches/ms-playwright",
			"~/Library/Caches/Cypress",
			"~/.cache/puppeteer",
		},
	}
}
