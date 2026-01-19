package shizukuapp

import "github.com/eleonorayaya/shizuku/internal/shizukuconfig"

type App interface {
	Name() string
	Enabled(config *shizukuconfig.Config) bool
}
