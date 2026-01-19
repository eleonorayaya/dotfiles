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
	"github.com/eleonorayaya/shizuku/apps/protonpass"
	"github.com/eleonorayaya/shizuku/apps/protonvpn"
	"github.com/eleonorayaya/shizuku/apps/python"
	"github.com/eleonorayaya/shizuku/apps/rust"
	"github.com/eleonorayaya/shizuku/apps/sfsymbols"
	"github.com/eleonorayaya/shizuku/apps/sketchybar"
	"github.com/eleonorayaya/shizuku/apps/terminal"
	"github.com/eleonorayaya/shizuku/apps/terraform"
	"github.com/eleonorayaya/shizuku/apps/zellij"
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var SyncCommand = &cobra.Command{
	Use:   "sync [flags] configs_path",
	Short: "",
	RunE:  sync,
}

type registeredApp struct {
	name string
	app  any
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

	apps := []registeredApp{
		{"sketchybar", sketchybar.New()},
		{"aerospace", aerospace.New()},
		{"fastfetch", fastfetch.New()},
		{"kitty", kitty.New()},
		{"jankyborders", jankyborders.New()},
		{"zellij", zellij.New()},
		{"nvim", nvim.New()},
		{"bat", bat.New()},
		{"git", git.New()},
		{"golang", golang.New()},
		{"lsd", lsd.New()},
		{"protonpass", protonpass.New()},
		{"protonvpn", protonvpn.New()},
		{"python", python.New()},
		{"rust", rust.New()},
		{"sfsymbols", sfsymbols.New()},
		{"terminal", terminal.New()},
		{"terraform", terraform.New()},
		{"desktoppr", desktoppr.New()},
	}

	for _, regApp := range apps {
		slog.Info("app syncing", "appName", regApp.name)

		if syncer, ok := regApp.app.(shizukuapp.FileSyncer); ok {
			if err := syncer.Sync(outDir, appConfig); err != nil {
				return fmt.Errorf("could not sync %s: %w", regApp.name, err)
			}

			slog.Info("app synced", "appName", regApp.name)
		}
	}

	envSetups := []*shizukuapp.EnvSetup{}
	for _, regApp := range apps {
		if provider, ok := regApp.app.(shizukuapp.EnvProvider); ok {
			envSetup, err := provider.Env()
			if err != nil {
				return fmt.Errorf("failed to get env setup for %s: %w", regApp.name, err)
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
