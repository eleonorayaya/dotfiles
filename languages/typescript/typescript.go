package typescript

import (
	"log/slog"

	"github.com/eleonorayaya/shizuku/app"
)

const yarnProxyEnv = `if [ -n "$HTTP_PROXY" ]; then export YARN_HTTP_PROXY="$HTTP_PROXY"; fi
if [ -n "$HTTPS_PROXY" ]; then export YARN_HTTPS_PROXY="$HTTPS_PROXY"; fi`

type App struct {
	npmrcPath  string
	yarnrcPath string
}

func New() *App {
	return &App{
		npmrcPath:  "~/.npmrc",
		yarnrcPath: "~/.yarnrc.yml",
	}
}

func (a *App) Name() string {
	return "typescript"
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		ZshenvScripts: []string{yarnProxyEnv},
	}, nil
}

func (a *App) AgentConfig() app.AgentConfig {
	registries, err := registryHosts(a.npmrcPath, a.yarnrcPath)
	if err != nil {
		slog.Warn("failed to read package registry hosts", "error", err)
	}

	return app.AgentConfig{
		Plugins: []string{
			"typescript-lsp@claude-plugins-official",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
		},
		AllowedBashCommands: []string{
			"npm install",
		},
		SandboxAllowedDomains: append([]string{
			"registry.npmjs.org",
			"registry.yarnpkg.com",
			"repo.yarnpkg.com",
			"codeload.github.com",
		}, registries...),
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
