package sync

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	shizuku "github.com/eleonorayaya/shizuku"
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
	cwd, _ := os.Getwd()
	slog.Debug("using source directory", "cwd", cwd)

	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	buildId := fmt.Sprintf("%v", time.Now().Unix())

	outDir := path.Join("out", buildId)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error created output dir: %w", err)
	}

	enabledLanguages := shizukuapp.FilterEnabledApps(shizuku.GetLanguages(), appConfig)
	enabledPrograms := shizukuapp.FilterEnabledApps(shizuku.GetPrograms(), appConfig)
	enabledAgents := shizukuapp.FilterEnabledApps(shizuku.GetAgents(), appConfig)

	if err := syncApps(enabledLanguages, outDir, appConfig); err != nil {
		return err
	}
	if err := syncApps(enabledPrograms, outDir, appConfig); err != nil {
		return err
	}

	ctx := shizukuapp.CollectAgentConfigs(append(enabledLanguages, enabledPrograms...))

	for _, app := range enabledAgents {
		slog.Info("app syncing", "appName", app.Name())

		if syncer, ok := app.(shizukuapp.ContextualSyncer); ok {
			if err := syncer.SyncWithContext(outDir, appConfig, ctx); err != nil {
				return fmt.Errorf("could not sync %s: %w", app.Name(), err)
			}
		} else if syncer, ok := app.(shizukuapp.FileSyncer); ok {
			if err := syncer.Sync(outDir, appConfig); err != nil {
				return fmt.Errorf("could not sync %s: %w", app.Name(), err)
			}
		}

		slog.Info("app synced", "appName", app.Name())
	}

	allEnabled := append(append(enabledLanguages, enabledPrograms...), enabledAgents...)

	envSetups := []*shizukuapp.EnvSetup{}
	for _, app := range allEnabled {
		if provider, ok := app.(shizukuapp.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", app.Name(), err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	envFileMap, err := shizukuapp.GenerateEnvFiles(envSetups, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate env files: %w", err)
	}

	if err := shizukuapp.SyncAppFiles(envFileMap, "~/.config/shizuku/"); err != nil {
		return fmt.Errorf("failed to sync env files: %w", err)
	}

	return nil
}

func syncApps(apps []shizukuapp.App, outDir string, config *shizukuconfig.Config) error {
	for _, app := range apps {
		slog.Info("app syncing", "appName", app.Name())

		if syncer, ok := app.(shizukuapp.FileSyncer); ok {
			if err := syncer.Sync(outDir, config); err != nil {
				return fmt.Errorf("could not sync %s: %w", app.Name(), err)
			}

			slog.Info("app synced", "appName", app.Name())
		}
	}
	return nil
}
