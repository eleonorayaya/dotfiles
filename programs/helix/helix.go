package helix

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
)

//go:embed all:contents
var contents embed.FS

const (
	steelRepoURL  = "https://github.com/mattwparas/helix"
	steelBranch   = "steel-event-system"
	steelCloneDir = ".local/src/helix"
	steelGitURL   = "https://github.com/mattwparas/steel.git"
	forestGitURL  = "https://github.com/Ra77a3l3-jar/forest.hx.git"
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	repoPath := filepath.Join(homeDir, steelCloneDir)

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(repoPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := run("", "git", "clone", "--branch", steelBranch, steelRepoURL, repoPath); err != nil {
			return fmt.Errorf("failed to clone helix: %w", err)
		}
	} else {
		if err := run(repoPath, "git", "pull", "origin", steelBranch); err != nil {
			return fmt.Errorf("failed to pull latest helix: %w", err)
		}
	}

	cargoBin := filepath.Join(homeDir, ".cargo", "bin")
	steps := [][]string{
		{"cargo", "install", "--git", steelGitURL, "steel-interpreter", "steel-language-server", "cargo-steel-lib", "--locked", "--force"},
		{"cargo", "install", "--git", steelGitURL, "steel-forge", "--locked", "--force"},
		{filepath.Join(cargoBin, "forge"), "pkg", "refresh"},
		{"cargo", "xtask", "code-gen"},
		{"cargo", "install", "--path", "helix-term", "--features", "steel,git", "--locked", "--force"},
		{filepath.Join(cargoBin, "forge"), "pkg", "install", "--git", forestGitURL},
	}
	for _, step := range steps {
		if err := run(repoPath, step[0], step[1:]...); err != nil {
			return fmt.Errorf("failed to install helix with steel: %w", err)
		}
	}

	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %v: %s: %w", name, args, string(output), err)
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
		vars = append(vars, app.EnvVar{Key: "HELIX_RUNTIME", Value: "$HOME/" + steelCloneDir + "/runtime"})
	}

	return &app.EnvSetup{
		Variables: vars,
		Aliases: []app.Alias{
			{Name: "vim", Command: "hx", Requires: "hx"},
		},
	}, nil
}
