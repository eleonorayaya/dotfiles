package lsd

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		Aliases: []shizukuenv.Alias{
			{Name: "ls", Command: "lsd"},
		},
	}, nil
}
