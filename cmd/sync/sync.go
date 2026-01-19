package sync

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/eleonorayaya/shizuku/apps"
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var SyncCommand = &cobra.Command{
	Use:   "sync [flags] configs_path",
	Short: "",
	RunE:  sync,
}

func sync(cmd *cobra.Command, args []string) error {
	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	buildId := fmt.Sprintf("%v", time.Now().Unix())

	outDir := path.Join("out", buildId)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error created output dir: %w", err)
	}

	allApps := apps.GetApps()
	enabledApps := shizukuapp.FilterEnabledApps(allApps, appConfig)

	for _, app := range enabledApps {
		slog.Info("app syncing", "appName", app.Name())

		if syncer, ok := app.(shizukuapp.FileSyncer); ok {
			if err := syncer.Sync(outDir, appConfig); err != nil {
				return fmt.Errorf("could not sync %s: %w", app.Name(), err)
			}

			slog.Info("app synced", "appName", app.Name())
		}
	}

	envSetups := []*shizukuapp.EnvSetup{}
	for _, app := range enabledApps {
		if provider, ok := app.(shizukuapp.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", app.Name(), err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	shizukuShPath := path.Join(outDir, "shizuku.sh")
	if err := shizukuapp.GenerateEnvFile(envSetups, shizukuShPath); err != nil {
		return fmt.Errorf("failed to generate env file: %w", err)
	}

	if err := shizukuapp.SyncAppFile("shizuku.sh", shizukuShPath, "~/.config/shizuku/"); err != nil {
		return fmt.Errorf("failed to sync env file: %w", err)
	}

	return nil
}
