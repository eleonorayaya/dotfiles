package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	shizuku "github.com/eleonorayaya/shizuku"
	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/examples/eleonora/data"
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
	"github.com/eleonorayaya/shizuku/util"
	"github.com/spf13/cobra"
)

func main() {
	cmd := shizuku.New(shizuku.Options{}).
		AddLanguages(
			golang.New(),
			lua.New(),
			python.New(),
			ruby.New(),
			rust.New(),
			typescript.New(),
			zig.New(),
		).
		AddPrograms(
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
		).
		AddAgent(claude.New(data.ClaudeOptions())).
		Command()

	cmd.AddCommand(upgradeCmd())

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func upgradeCmd() *cobra.Command {
	var branch string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Pull latest changes and rebuild the shizuku binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoDir, err := util.NormalizeFilePath(config.SourceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve source directory: %w", err)
			}

			if _, err := os.Stat(repoDir); os.IsNotExist(err) {
				return fmt.Errorf("shizuku repo not found at %s, run 'shizuku install' first", repoDir)
			}

			slog.Info("pulling latest changes", "branch", branch)
			if err := runExec(repoDir, "git", "fetch", "origin", branch); err != nil {
				return fmt.Errorf("failed to fetch: %w", err)
			}
			if err := runExec(repoDir, "git", "checkout", branch); err != nil {
				return fmt.Errorf("failed to checkout branch: %w", err)
			}
			if err := runExec(repoDir, "git", "pull", "origin", branch); err != nil {
				return fmt.Errorf("failed to pull latest changes: %w", err)
			}

			slog.Info("building and installing shizuku")
			if err := runExec(repoDir, "task", "install"); err != nil {
				return fmt.Errorf("failed to build and install: %w", err)
			}

			slog.Info("upgrade complete, run 'shizuku install' and 'shizuku sync' to apply changes")
			return nil
		},
	}
	cmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to pull from")
	return cmd
}

func runExec(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}
