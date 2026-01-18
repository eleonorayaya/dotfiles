package git

import "github.com/eleonorayaya/shizuku/internal/shizukuapp"

const gitCompletionInit = `fpath=(~/.zsh $fpath)

autoload -Uz compinit && compinit`

type App struct{}

func New() *App {
	return &App{}
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
