package git

import (
	"fmt"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

const gitCompletionInit = `fpath=(~/.zsh $fpath)

autoload -Uz compinit && compinit`

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "git"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	fileMap, err := shizukuapp.GenerateAppFiles("git", nil, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	cmd := exec.Command("git", "config", "--global", "core.excludesfile", "~/.gitignore_global")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set global excludesfile: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		InitScripts: []string{gitCompletionInit},
		Aliases: []shizukuapp.Alias{
			{Name: "gsu", Command: "git status -uno"},
			{Name: "gittouch", Command: "git pull --rebase && git commit -m 'touch' --allow-empty && git push"},
		},
	}, nil
}
