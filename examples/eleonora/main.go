package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	shizuku "github.com/eleonorayaya/shizuku"
	"github.com/eleonorayaya/shizuku/apps/agents/claude"
	"github.com/eleonorayaya/shizuku/apps/languages/golang"
	"github.com/eleonorayaya/shizuku/apps/languages/lua"
	"github.com/eleonorayaya/shizuku/apps/languages/python"
	"github.com/eleonorayaya/shizuku/apps/languages/ruby"
	"github.com/eleonorayaya/shizuku/apps/languages/rust"
	"github.com/eleonorayaya/shizuku/apps/languages/typescript"
	"github.com/eleonorayaya/shizuku/apps/languages/zig"
	"github.com/eleonorayaya/shizuku/apps/programs/aerospace"
	"github.com/eleonorayaya/shizuku/apps/programs/bat"
	"github.com/eleonorayaya/shizuku/apps/programs/buildkite"
	"github.com/eleonorayaya/shizuku/apps/programs/desktoppr"
	"github.com/eleonorayaya/shizuku/apps/programs/fastfetch"
	"github.com/eleonorayaya/shizuku/apps/programs/git"
	"github.com/eleonorayaya/shizuku/apps/programs/glow"
	"github.com/eleonorayaya/shizuku/apps/programs/jankyborders"
	"github.com/eleonorayaya/shizuku/apps/programs/k9s"
	"github.com/eleonorayaya/shizuku/apps/programs/kitty"
	"github.com/eleonorayaya/shizuku/apps/programs/lsd"
	"github.com/eleonorayaya/shizuku/apps/programs/nvim"
	"github.com/eleonorayaya/shizuku/apps/programs/protonpass"
	"github.com/eleonorayaya/shizuku/apps/programs/protonvpn"
	"github.com/eleonorayaya/shizuku/apps/programs/sfsymbols"
	"github.com/eleonorayaya/shizuku/apps/programs/sketchybar"
	"github.com/eleonorayaya/shizuku/apps/programs/terminal"
	"github.com/eleonorayaya/shizuku/apps/programs/terraform"
	"github.com/eleonorayaya/shizuku/apps/programs/tmux"
	"github.com/eleonorayaya/shizuku/apps/programs/utena"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/examples/eleonora/data"
	"github.com/eleonorayaya/shizuku/util"
	"github.com/spf13/cobra"
)

var (
	verbose     bool
	showContent bool
	branch      string
)

func newBuilder() *shizuku.Builder {
	return shizuku.New(shizuku.Options{Verbose: verbose}).
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
		AddAgent(claude.New(data.ClaudeOptions()))
}

var rootCmd = &cobra.Command{
	Use: "shizuku",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if verbose {
			slog.SetLogLoggerLevel(slog.LevelDebug)
		}

		if _, err := os.Stat("apps"); os.IsNotExist(err) {
			sourceDir, err := util.NormalizeFilePath(config.SourceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve source directory: %w", err)
			}

			if err := os.Chdir(sourceDir); err != nil {
				return fmt.Errorf("failed to change to source directory %s: %w", sourceDir, err)
			}
		}

		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize shizuku configuration directory and create default config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		return newBuilder().Init()
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all application configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return newBuilder().Sync(context.Background())
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show what would change on next sync",
	RunE: func(cmd *cobra.Command, args []string) error {
		report, err := newBuilder().Diff(context.Background())
		if err != nil {
			return err
		}

		if report.TotalChanged == 0 {
			fmt.Println("No differences found.")
			return nil
		}

		for _, r := range report.Results {
			fmt.Printf("%s:\n", r.Name)
			for _, f := range r.Changed {
				fmt.Printf("  M %s\n", f)
			}
		}

		fmt.Printf("\n%d file(s) with differences. Diff files written to %s/\n", report.TotalChanged, report.OutDir)

		if showContent {
			fmt.Println()
			for _, r := range report.Results {
				for _, f := range r.Changed {
					diffPath := r.FileMap[f] + ".diff"
					content, err := os.ReadFile(diffPath)
					if err != nil {
						return fmt.Errorf("failed to read diff file %s: %w", diffPath, err)
					}
					fmt.Println(string(content))
				}
			}
		}

		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install application dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return newBuilder().Install(context.Background())
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available apps and their enabled status",
	RunE: func(cmd *cobra.Command, args []string) error {
		statuses, err := newBuilder().List()
		if err != nil {
			return err
		}

		fmt.Println("Available apps:")
		fmt.Println()
		for _, s := range statuses {
			status := "disabled"
			if s.Enabled {
				status = "enabled"
			}
			fmt.Printf("  %-20s %s\n", s.Name, status)
		}

		return nil
	},
}

var upgradeCmd = &cobra.Command{
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

func runExec(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	diffCmd.Flags().BoolVarP(&showContent, "print", "p", false, "Print diff contents to stdout")
	upgradeCmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to pull from")

	rootCmd.AddCommand(initCmd, syncCmd, diffCmd, installCmd, listCmd, upgradeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
