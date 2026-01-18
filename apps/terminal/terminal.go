package terminal

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/shizukuenv"
)

const antigenInit = `source $(brew --prefix)/share/antigen/antigen.zsh

antigen bundle jeffreytse/zsh-vi-mode > /dev/null

antigen apply`

const ohmyposhInit = `eval "$(oh-my-posh init zsh --config ~/.config/ohmyposh/ohmyposh.json)"`

func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("terminal", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/ohmyposh/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		InitScripts: []string{antigenInit, ohmyposhInit},
		Aliases: []shizukuenv.Alias{
			{Name: "c", Command: "clear"},
			{Name: "curltime", Command: "curl -o /dev/null -s -w 'Total: %{time_total}s\\n'"},
		},
		Functions: []shizukuenv.ShellFunction{
			{Name: "colormap", Body: colormapFunction},
		},
	}, nil
}

const colormapFunction = `    for i in {0..255}; do
        printf "\x1b[38;5;${i}mcolour${i}\x1b[0m\n"
    done`
