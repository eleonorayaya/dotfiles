package sync

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/eleonorayaya/shizuku/apps/aerospace"
	"github.com/eleonorayaya/shizuku/apps/bat"
	"github.com/eleonorayaya/shizuku/apps/desktoppr"
	"github.com/eleonorayaya/shizuku/apps/fastfetch"
	"github.com/eleonorayaya/shizuku/apps/git"
	"github.com/eleonorayaya/shizuku/apps/golang"
	"github.com/eleonorayaya/shizuku/apps/jankyborders"
	"github.com/eleonorayaya/shizuku/apps/kitty"
	"github.com/eleonorayaya/shizuku/apps/lsd"
	"github.com/eleonorayaya/shizuku/apps/nvim"
	"github.com/eleonorayaya/shizuku/apps/python"
	"github.com/eleonorayaya/shizuku/apps/rust"
	"github.com/eleonorayaya/shizuku/apps/sketchybar"
	"github.com/eleonorayaya/shizuku/apps/terminal"
	"github.com/eleonorayaya/shizuku/apps/terraform"
	"github.com/eleonorayaya/shizuku/apps/zellij"
	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/shizukuenv"
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
		env  func() (*shizukuenv.EnvSetup, error)
	}{
		{"sketchybar", sketchybar.Sync, nil},
		{"aerospace", aerospace.Sync, nil},
		{"fastfetch", fastfetch.Sync, fastfetch.Env},
		{"kitty", kitty.Sync, nil},
		{"jankyborders", jankyborders.Sync, nil},
		{"zellij", zellij.Sync, zellij.Env},
		{"nvim", nvim.Sync, nvim.Env},
		{"bat", nil, bat.Env},
		{"git", nil, git.Env},
		{"golang", nil, golang.Env},
		{"lsd", nil, lsd.Env},
		{"python", nil, python.Env},
		{"rust", nil, rust.Env},
		{"terminal", terminal.Sync, terminal.Env},
		{"terraform", nil, terraform.Env},
		{"desktoppr", desktoppr.Sync, nil},
	}

	for _, app := range apps {
		slog.Info("app syncing", "appName", app.name)

		if app.fn != nil {
			if err := app.fn(outDir, appConfig); err != nil {
				return fmt.Errorf("could not sync %s: %w", app.name, err)
			}

			slog.Info("app synced", "appName", app.name)
		}
	}

	envSetups := []*shizukuenv.EnvSetup{}
	for _, app := range apps {
		if app.env != nil {
			envSetup, err := app.env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", app.name, err)
			}
			envSetups = append(envSetups, envSetup)
		}
	}

	shizukuShPath := path.Join(outDir, "shizuku.sh")
	if err := shizukuenv.GenerateEnvFile(envSetups, shizukuShPath); err != nil {
		return fmt.Errorf("failed to generate env file: %w", err)
	}

	if err := internal.SyncAppFile("shizuku.sh", shizukuShPath, "~/.config/shizuku/"); err != nil {
		return fmt.Errorf("failed to sync env file: %w", err)
	}

	return nil
}
