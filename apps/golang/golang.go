package golang

import "github.com/eleonorayaya/shizuku/internal/shizukuenv"

func Env() (*shizukuenv.EnvSetup, error) {
	return &shizukuenv.EnvSetup{
		PathDirs: []shizukuenv.PathDir{
			{Path: "$HOME/go/bin", Priority: 20},
		},
	}, nil
}
