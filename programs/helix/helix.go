package helix

import (
	"embed"
	"fmt"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const (
	helixForkURL    = "https://github.com/mattwparas/helix"
	helixForkBranch = "steel-event-system"
	helixForkDir    = "~/.local/src/helix"
	steelURL        = "https://github.com/mattwparas/steel.git"
	forestURL       = "https://github.com/Ra77a3l3-jar/forest.hx.git"
)

type Options struct {
	Steel bool
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "helix"
}

var spec = pkg.Spec{
	Brew: "helix",
	Apt:  "hx",
	Bin:  "hx",
	Release: &pkg.Release{
		Repo:  "helix-editor/helix",
		Asset: `helix-.*-{arch}-linux\.tar\.xz`,
		Dirs:  map[string]string{"runtime": "~/.config/helix/runtime"},
	},
}

func (a *App) Install(ctx *app.Context) error {
	if a.opts.Steel {
		return installSteel()
	}

	if err := pkg.Install(spec); err != nil {
		return fmt.Errorf("failed to install helix: %w", err)
	}

	return nil
}

func installSteel() error {
	repoPath, err := util.CloneOrPull(helixForkURL, helixForkBranch, helixForkDir)
	if err != nil {
		return fmt.Errorf("failed to update helix fork: %w", err)
	}

	forge, err := util.NormalizeFilePath("~/.cargo/bin/forge")
	if err != nil {
		return err
	}

	steps := [][]string{
		{"cargo", "install", "--git", steelURL, "steel-interpreter", "steel-language-server", "cargo-steel-lib", "steel-forge", "--locked"},
		{forge, "pkg", "refresh"},
		{"cargo", "xtask", "code-gen"},
		{"cargo", "install", "--path", "helix-term", "--features", "steel,git", "--locked"},
		{forge, "pkg", "install", "--git", forestURL},
	}
	for _, step := range steps {
		if err := util.Run(repoPath, step[0], step[1:]...); err != nil {
			return fmt.Errorf("failed to install helix with steel: %w", err)
		}
	}

	return nil
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		SandboxAllowWrite: []string{
			"~/.cache/helix/",
		},
	}
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("helix", contents, map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}
	if !a.opts.Steel {
		delete(fileMap, "init.scm")
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/helix/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	vars := []app.EnvVar{
		{Key: "EDITOR", Value: "hx"},
		{Key: "VISUAL", Value: "hx"},
	}
	if a.opts.Steel {
		vars = append(vars, app.EnvVar{Key: "HELIX_RUNTIME", Value: strings.Replace(helixForkDir, "~", "$HOME", 1) + "/runtime"})
	}

	return &app.EnvSetup{
		Variables: vars,
		Aliases: []app.Alias{
			{Name: "vim", Command: "hx", Requires: "hx"},
		},
	}, nil
}
