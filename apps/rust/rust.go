package rust

import (
	"path"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuenv"
)

func Env() (*shizukuenv.EnvSetup, error) {
	rustupPrefix, err := internal.GetBrewAppPrefix("rustup")
	if err != nil {
		return nil, err
	}

	rustupBin := path.Join(rustupPrefix, "bin")

	return &shizukuenv.EnvSetup{
		PathDirs: []shizukuenv.PathDir{
			{Path: rustupBin, Priority: 10},
			{Path: "$HOME/.cargo/bin", Priority: 10},
		},
		Variables: []shizukuenv.EnvVar{
			{Key: "RUSTUP_HOME", Value: "$HOME/.rustup"},
		},
	}, nil
}
