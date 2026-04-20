package terminal

import (
	"embed"
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

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

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("antigen", false); err != nil {
		return fmt.Errorf("failed to install antigen: %w", err)
	}

	if err := util.AddTap("jandedobbeleer/oh-my-posh"); err != nil {
		return fmt.Errorf("failed to add tap: %w", err)
	}

	if err := util.InstallBrewPackage("jandedobbeleer/oh-my-posh/oh-my-posh", false); err != nil {
		return fmt.Errorf("failed to install oh-my-posh: %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	data := map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}

	fileMap, err := app.GenerateAppFiles("terminal", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/ohmyposh/",
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
	return &app.EnvSetup{
		PathDirs: []app.PathDir{
			{Path: "$HOME/.local/bin", Priority: 5},
		},
		InitScripts: []string{antigenInit, ohmyposhInit},
		Aliases: []app.Alias{
			{Name: "c", Command: "clear"},
			{Name: "curltime", Command: "curl -o /dev/null -s -w 'Total: %{time_total}s\\n'"},
		},
		Functions: []app.ShellFunction{
			{Name: "colormap", Body: colormapFunction},
		},
	}, nil
}

const colormapFunction = `    for i in {0..255}; do
        printf "\x1b[38;5;${i}mcolour${i}\x1b[0m\n"
    done`
