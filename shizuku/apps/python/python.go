package python

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		Aliases: []shizukuenv.Alias{
			{Name: "python", Command: "python3"},
		},
	}, nil
}
