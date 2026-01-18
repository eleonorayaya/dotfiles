package git

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

const gitCompletionInit = `fpath=(~/.zsh $fpath)

autoload -Uz compinit && compinit`

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		InitScripts: []string{gitCompletionInit},
		Aliases: []shizukuenv.Alias{
			{Name: "gsu", Command: "git status -uno"},
			{Name: "gittouch", Command: "git pull --rebase && git commit -m 'touch' --allow-empty && git push"},
		},
	}, nil
}
