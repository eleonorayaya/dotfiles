package install

import (
	"fmt"
	"log/slog"

	"github.com/eleonorayaya/shizuku/apps"
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var InstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install application dependencies",
	RunE:  install,
}

func install(cmd *cobra.Command, args []string) error {
	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	registeredApps := apps.GetApps()

	for _, app := range registeredApps {
		if !app.Enabled(appConfig) {
			continue
		}

		if installer, ok := app.(shizukuapp.Installer); ok {
			slog.Info("installing app dependencies", "appName", app.Name())

			if err := installer.Install(appConfig); err != nil {
				return fmt.Errorf("failed to install %s: %w", app.Name(), err)
			}

			slog.Info("app dependencies installed", "appName", app.Name())
		}
	}

	return nil
}
