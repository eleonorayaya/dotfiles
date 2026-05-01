package rust

import (
	"fmt"
	"path"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "rust"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("rustup", false); err != nil {
		return fmt.Errorf("failed to install rustup: %w", err)
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		Plugins: []string{
			"rust-analyzer-lsp@claude-plugins-official",
		},
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
		},
		SandboxAllowedDomains: []string{
			"index.crates.io",
			"static.crates.io",
			"static.rust-lang.org",
			"crates.io",
			"docs.rs",
		},
		SandboxAllowWrite: []string{
			"~/.cargo",
			"~/.rustup",
			"~/Library/Caches/cargo",
		},
	}
}

func (a *App) Env() (*app.EnvSetup, error) {
	rustupPrefix, err := util.GetBrewAppPrefix("rustup")
	if err != nil {
		return nil, err
	}

	rustupBin := path.Join(rustupPrefix, "bin")

	return &app.EnvSetup{
		PathDirs: []app.PathDir{
			{Path: rustupBin, Priority: 10},
			{Path: "$HOME/.cargo/bin", Priority: 10},
		},
		Variables: []app.EnvVar{
			{Key: "RUSTUP_HOME", Value: "$HOME/.rustup"},
		},
	}, nil
}
