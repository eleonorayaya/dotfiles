package apps

import (
	"github.com/eleonorayaya/shizuku/apps/aerospace"
	"github.com/eleonorayaya/shizuku/apps/bat"
	"github.com/eleonorayaya/shizuku/apps/claude"
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
)

func GetApps() []shizukuapp.App {
	return []shizukuapp.App{
		sketchybar.New(),
		aerospace.New(),
		fastfetch.New(),
		kitty.New(),
		jankyborders.New(),
		zellij.New(),
		nvim.New(),
		bat.New(),
		git.New(),
		golang.New(),
		lsd.New(),
		protonpass.New(),
		protonvpn.New(),
		python.New(),
		rust.New(),
		sfsymbols.New(),
		terminal.New(),
		terraform.New(),
		desktoppr.New(),
		claude.New(),
	}
}
