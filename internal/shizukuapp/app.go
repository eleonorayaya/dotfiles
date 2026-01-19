package shizukuapp

import "github.com/eleonorayaya/shizuku/internal/shizukuconfig"

type App interface {
	Name() string
	Enabled(config *shizukuconfig.Config) bool
}

func FilterEnabledApps(apps []App, config *shizukuconfig.Config) []App {
	enabled := make([]App, 0)
	for _, app := range apps {
		if app.Enabled(config) {
			enabled = append(enabled, app)
		}
	}
	return enabled
}
