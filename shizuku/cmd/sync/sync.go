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
	"github.com/spf13/cobra"
)

var SyncCommand = &cobra.Command{
	Use:   "sync [flags] configs_path",
	Short: "",
	RunE:  sync,
}

func sync(cmd *cobra.Command, args []string) error {

	buildId := fmt.Sprintf("%v", time.Now().Unix())

	outDir := path.Join("out", buildId)
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error created output dir: %w", err)
	}

	// Define apps to sync in order
	apps := []struct {
		name string
		fn   func(string) error
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

	// Sync each app
	for _, app := range apps {
		slog.Info("app synced", "appName", app.name)

		if err := app.fn(outDir); err != nil {
			return fmt.Errorf("could not sync %s: %w", app.name, err)
		}

		slog.Info("app synced", "appName", app.name)
	}

	fmt.Printf("\nBuild output: %s\n", outDir)
	return nil
}
