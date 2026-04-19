package shizuku

import (
	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/languages/golang"
	"github.com/eleonorayaya/shizuku/languages/lua"
	"github.com/eleonorayaya/shizuku/languages/python"
	"github.com/eleonorayaya/shizuku/languages/ruby"
	"github.com/eleonorayaya/shizuku/languages/rust"
	"github.com/eleonorayaya/shizuku/languages/typescript"
	"github.com/eleonorayaya/shizuku/languages/zig"
	"github.com/eleonorayaya/shizuku/programs/aerospace"
	"github.com/eleonorayaya/shizuku/programs/bat"
	"github.com/eleonorayaya/shizuku/programs/buildkite"
	"github.com/eleonorayaya/shizuku/programs/desktoppr"
	"github.com/eleonorayaya/shizuku/programs/fastfetch"
	"github.com/eleonorayaya/shizuku/programs/git"
	"github.com/eleonorayaya/shizuku/programs/glow"
	"github.com/eleonorayaya/shizuku/programs/jankyborders"
	"github.com/eleonorayaya/shizuku/programs/k9s"
	"github.com/eleonorayaya/shizuku/programs/kitty"
	"github.com/eleonorayaya/shizuku/programs/lsd"
	"github.com/eleonorayaya/shizuku/programs/nvim"
	"github.com/eleonorayaya/shizuku/programs/protonpass"
	"github.com/eleonorayaya/shizuku/programs/protonvpn"
	"github.com/eleonorayaya/shizuku/programs/sfsymbols"
	"github.com/eleonorayaya/shizuku/programs/sketchybar"
	"github.com/eleonorayaya/shizuku/programs/terminal"
	"github.com/eleonorayaya/shizuku/programs/terraform"
	"github.com/eleonorayaya/shizuku/programs/tmux"
	"github.com/eleonorayaya/shizuku/programs/utena"
)

func GetLanguages() []shizukuapp.App {
	return []shizukuapp.App{
		golang.New(),
		lua.New(),
		python.New(),
		ruby.New(),
		rust.New(),
		typescript.New(),
		zig.New(),
	}
}

func GetPrograms() []shizukuapp.App {
	return []shizukuapp.App{
		sketchybar.New(),
		aerospace.New(),
		fastfetch.New(),
		kitty.New(),
		jankyborders.New(),
		nvim.New(),
		bat.New(),
		git.New(),
		lsd.New(),
		protonpass.New(),
		protonvpn.New(),
		sfsymbols.New(),
		terminal.New(),
		terraform.New(),
		tmux.New(),
		desktoppr.New(),
		glow.New(),
		utena.New(),
		k9s.New(),
		buildkite.New(),
	}
}

func GetAgents() []shizukuapp.App {
	return []shizukuapp.App{
		claude.New(),
	}
}

func GetApps() []shizukuapp.App {
	all := []shizukuapp.App{}
	all = append(all, GetLanguages()...)
	all = append(all, GetPrograms()...)
	all = append(all, GetAgents()...)
	return all
}
