package lsd

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
	return "lsd"
}

var spec = pkg.Spec{
	Brew: "lsd",
	Apt:  "lsd",
	Bin:  "lsd",
	Release: &pkg.Release{
		Repo:  "lsd-rs/lsd",
		Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`,
	},
}

func (a *App) Install(ctx *app.Context) error {
	if err := pkg.Install(spec); err != nil {
		return fmt.Errorf("failed to install lsd: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		Aliases: []app.Alias{
			{Name: "ls", Command: "lsd", Requires: "lsd"},
		},
	}, nil
}
