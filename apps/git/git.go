package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
	"gopkg.in/ini.v1"
)

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

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	fileMap, err := shizukuapp.GenerateAppFiles("git", nil, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	mergedPath, err := mergeGitConfig(outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to merge gitconfig: %w", err)
	}
	fileMap[".gitconfig"] = mergedPath

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

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		InitScripts: []string{gitCompletionInit},
		Aliases: []shizukuapp.Alias{
			{Name: "gsu", Command: "git status -uno"},
			{Name: "gittouch", Command: "git pull --rebase && git commit -m 'touch' --allow-empty && git push"},
		},
	}, nil
}
