package bat

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "bat"
}

var spec = pkg.Spec{
	Brew:   "bat",
	Apt:    "bat",
	AptBin: "batcat",
	Bin:    "bat",
	Release: &pkg.Release{
		Repo:  "sharkdp/bat",
		Asset: `bat-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`,
	},
}

func (a *App) Install(ctx *app.Context) error {
	if err := pkg.Install(spec); err != nil {
		return fmt.Errorf("failed to install bat: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Aliases: []app.Alias{
			{Name: "cat", Command: "bat", Requires: "bat"},
		},
	}, nil
}
