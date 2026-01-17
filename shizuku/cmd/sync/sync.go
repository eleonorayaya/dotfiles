package sync

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/eleonorayaya/shizuku/apps/aerospace"
	"github.com/eleonorayaya/shizuku/apps/desktoppr"
	"github.com/eleonorayaya/shizuku/apps/fastfetch"
	"github.com/eleonorayaya/shizuku/apps/jankyborders"
	"github.com/eleonorayaya/shizuku/apps/kitty"
	"github.com/eleonorayaya/shizuku/apps/nvim"
	"github.com/eleonorayaya/shizuku/apps/sketchybar"
	"github.com/eleonorayaya/shizuku/apps/zellij"
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

	apps := []struct {
		name string
		fn   func(string, *shizukuconfig.Config) error
	}{
		{"sketchybar", sketchybar.Sync},
		{"aerospace", aerospace.Sync},
		{"fastfetch", fastfetch.Sync},
		{"kitty", kitty.Sync},
		{"jankyborders", jankyborders.Sync},
		{"zellij", zellij.Sync},
		{"nvim", nvim.Sync},
		{"desktoppr", desktoppr.Sync},
	}

	for _, app := range apps {
		slog.Info("app syncing", "appName", app.name)

		if err := app.fn(outDir, appConfig); err != nil {
			return fmt.Errorf("could not sync %s: %w", app.name, err)
		}

		slog.Info("app synced", "appName", app.name)
	}

	return nil
}
