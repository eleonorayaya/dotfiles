package terraform

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		Aliases: []shizukuenv.Alias{
			{Name: "tf", Command: "terraform"},
			{Name: "tfmt", Command: "terraform fmt -recursive"},
		},
	}, nil
}
