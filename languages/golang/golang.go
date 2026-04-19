package golang

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "golang"
}

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetLanguageEnabled(config.LanguageGo)
}

func (a *App) Install(cfg *config.Config) error {
	if err := util.InstallBrewPackage("go-task", false); err != nil {
		return fmt.Errorf("failed to install go-task: %w", err)
	}

	if !util.BinaryExists("gopls") {
		slog.Debug("installing gopls via go install")

		cmd := exec.Command("go", "install", "golang.org/x/tools/gopls@latest")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install gopls: %w\nOutput: %s", err, string(output))
		}

		slog.Debug("gopls installed successfully")
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"gopls-lsp@claude-plugins-official",
			"charm-dev@charm-dev-skills",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
			"charm-dev-skills":        {Repo: "williavs/charm-dev-skill-marketplace"},
		},
		AllowedCommands: []string{
			"Bash(go build:*)",
			"Bash(go vet:*)",
			"Bash(go mod tidy:*)",
		},
		SandboxAllowWrite: []string{
			"~/.cache/go-build",
			"~/.config/go",
			"~/.local/share/go",
			"~/go",
			"~/.cache/golangci-lint",
			"~/Library/Caches/go-build",
			"~/Library/Caches/golangci-lint",
		},
	}
}

func (a *App) Env() (*app.EnvSetup, error) {
	cmd := exec.Command("go", "env", "GOPATH")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GOPATH: %w", err)
	}

	gopath := strings.TrimSpace(string(output))

	return &app.EnvSetup{
		PathDirs: []app.PathDir{
			{Path: path.Join(gopath, "bin"), Priority: 20},
		},
	}, nil
}
