package git

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
	"gopkg.in/ini.v1"
)

//go:embed all:contents
var contents embed.FS

const gitCompletionInit = `fpath=(~/.zsh $fpath)

autoload -Uz compinit && compinit`

var desiredGitConfigs = map[string]string{
	"core.excludesfile":    "~/.gitignore_global",
	"push.autoSetupRemote": "true",
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "git"
}

func (a *App) AgentConfig() app.AgentConfig {
	return app.AgentConfig{
		AllowedCommands: []string{
			"Bash(git add:*)",
			"Bash(git commit:*)",
			"Bash(git --version:*)",
			"Bash(git status:*)",
			"Bash(git diff:*)",
			"Bash(git log:*)",
			"Bash(git fetch:*)",
			"Bash(git push:*)",
			"Bash(git rebase:*)",
			"Bash(git stash:*)",
			"Bash(git grep:*)",
		},
	}
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("git", contents, nil, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	mergedPath, err := mergeGitConfig(ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to merge gitconfig: %w", err)
	}
	fileMap[".gitconfig"] = mergedPath

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/",
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

func mergeGitConfig(outDir string) (string, error) {
	gitConfigPath, err := util.NormalizeFilePath("~/.gitconfig")
	if err != nil {
		return "", fmt.Errorf("failed to normalize gitconfig path: %w", err)
	}

	cfg, err := ini.LooseLoad(gitConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load gitconfig: %w", err)
	}

	for key, value := range desiredGitConfigs {
		lastDot := strings.LastIndex(key, ".")
		section := key[:lastDot]
		keyName := key[lastDot+1:]
		cfg.Section(section).Key(keyName).SetValue(value)
	}

	outPath := filepath.Join(outDir, "git", ".gitconfig")
	if err := cfg.SaveTo(outPath); err != nil {
		return "", fmt.Errorf("failed to write merged gitconfig: %w", err)
	}

	return outPath, nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		InitScripts: []string{gitCompletionInit},
		Aliases: []app.Alias{
			{Name: "gsu", Command: "git status -uno"},
			{Name: "gittouch", Command: "git pull --rebase && git commit -m 'touch' --allow-empty && git push"},
		},
	}, nil
}
