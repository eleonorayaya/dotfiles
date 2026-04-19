package typescript

import (
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "typescript"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageTypescript)
}

func (a *App) AgentConfig() shizukuapp.AgentConfig {
	return shizukuapp.AgentConfig{
		Plugins: []string{
			"typescript-lsp@claude-plugins-official",
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
