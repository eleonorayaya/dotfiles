package git

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		Aliases: []shizukuenv.Alias{
			{Name: "gsu", Command: "git status -uno"},
			{Name: "gittouch", Command: "git pull --rebase && git commit -m 'touch' --allow-empty && git push"},
		},
	}, nil
}
