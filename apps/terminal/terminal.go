package terminal

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/theme"
	"github.com/eleonorayaya/shizuku/internal/util"
)

const antigenInit = `source $(brew --prefix)/share/antigen/antigen.zsh

antigen bundle jeffreytse/zsh-vi-mode > /dev/null

antigen apply`

const ohmyposhInit = `eval "$(oh-my-posh init zsh --config ~/.config/ohmyposh/ohmyposh.json)"`

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terminal"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("antigen"); err != nil {
		return fmt.Errorf("failed to install antigen: %w", err)
	}

	if err := util.AddTap("jandedobbeleer/oh-my-posh"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("jandedobbeleer/oh-my-posh/oh-my-posh"); err != nil {
		return fmt.Errorf("failed to install oh-my-posh: %w", err)
	}

	return nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config, themeData *theme.Theme) error {
	data := map[string]any{
		"Colors": themeData.Colors,
	}

	fileMap, err := shizukuapp.GenerateAppFiles("terminal", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/ohmyposh/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		InitScripts: []string{antigenInit, ohmyposhInit},
		Aliases: []shizukuapp.Alias{
			{Name: "c", Command: "clear"},
			{Name: "curltime", Command: "curl -o /dev/null -s -w 'Total: %{time_total}s\\n'"},
		},
		Functions: []shizukuapp.ShellFunction{
			{Name: "colormap", Body: colormapFunction},
		},
	}, nil
}

const colormapFunction = `    for i in {0..255}; do
        printf "\x1b[38;5;${i}mcolour${i}\x1b[0m\n"
    done`
