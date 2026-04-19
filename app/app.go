package app

import "github.com/eleonorayaya/shizuku/config"

type App interface {
	Name() string
	Enabled(cfg *config.Config) bool
}

func FilterEnabledApps(apps []App, cfg *config.Config) []App {
	enabled := make([]App, 0)
	for _, app := range apps {
		if app.Enabled(cfg) {
			enabled = append(enabled, app)
		}
	}
	return enabled
}
