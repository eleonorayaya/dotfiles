package install

import (
	"fmt"
	"log/slog"

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
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/spf13/cobra"
)

var InstallCommand = &cobra.Command{
	Use:   "install",
	Short: "Install application dependencies",
	RunE:  install,
}

type registeredApp struct {
	name string
	app  any
}

func install(cmd *cobra.Command, args []string) error {
	appConfig, err := shizukuconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
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
		{"python", python.New()},
		{"rust", rust.New()},
		{"terminal", terminal.New()},
		{"terraform", terraform.New()},
		{"desktoppr", desktoppr.New()},
	}

	for _, regApp := range apps {
		if installer, ok := regApp.app.(shizukuapp.Installer); ok {
			slog.Info("installing app dependencies", "appName", regApp.name)

			if err := installer.Install(appConfig); err != nil {
				return fmt.Errorf("failed to install %s: %w", regApp.name, err)
			}

			slog.Info("app dependencies installed", "appName", regApp.name)
		}
	}

	return nil
}
