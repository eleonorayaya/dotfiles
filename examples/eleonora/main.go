package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	shizuku "github.com/eleonorayaya/shizuku"
	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/examples/eleonora/data"
	"github.com/eleonorayaya/shizuku/languages/golang"
	"github.com/eleonorayaya/shizuku/languages/lua"
	"github.com/eleonorayaya/shizuku/languages/python"
	"github.com/eleonorayaya/shizuku/languages/ruby"
	"github.com/eleonorayaya/shizuku/languages/rust"
	"github.com/eleonorayaya/shizuku/languages/swift"
	"github.com/eleonorayaya/shizuku/languages/typescript"
	"github.com/eleonorayaya/shizuku/languages/zig"
	"github.com/eleonorayaya/shizuku/pkg"
	"github.com/eleonorayaya/shizuku/programs/acli"
	"github.com/eleonorayaya/shizuku/programs/aerospace"
	"github.com/eleonorayaya/shizuku/programs/bat"
	"github.com/eleonorayaya/shizuku/programs/buildkite"
	"github.com/eleonorayaya/shizuku/programs/datadog"
	"github.com/eleonorayaya/shizuku/programs/desktoppr"
	"github.com/eleonorayaya/shizuku/programs/dock"
	"github.com/eleonorayaya/shizuku/programs/fastfetch"
	"github.com/eleonorayaya/shizuku/programs/git"
	"github.com/eleonorayaya/shizuku/programs/glow"
	"github.com/eleonorayaya/shizuku/programs/helix"
	"github.com/eleonorayaya/shizuku/programs/jankyborders"
	"github.com/eleonorayaya/shizuku/programs/k9s"
	"github.com/eleonorayaya/shizuku/programs/kitty"
	"github.com/eleonorayaya/shizuku/programs/lsd"
	"github.com/eleonorayaya/shizuku/programs/mise"
	"github.com/eleonorayaya/shizuku/programs/notion"
	"github.com/eleonorayaya/shizuku/programs/nvim"
	"github.com/eleonorayaya/shizuku/programs/protonpass"
	"github.com/eleonorayaya/shizuku/programs/protonvpn"
	"github.com/eleonorayaya/shizuku/programs/sfsymbols"
	"github.com/eleonorayaya/shizuku/programs/sketchybar"
	"github.com/eleonorayaya/shizuku/programs/terminal"
	"github.com/eleonorayaya/shizuku/programs/terminalbrowser"
	"github.com/eleonorayaya/shizuku/programs/terraform"
	"github.com/eleonorayaya/shizuku/programs/tmux"
	"github.com/eleonorayaya/shizuku/programs/tuios"
	"github.com/eleonorayaya/shizuku/programs/utena"
	"github.com/eleonorayaya/shizuku/programs/vesktop"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
	"github.com/eleonorayaya/shizuku/util"
	"github.com/spf13/cobra"
)

const (
	sourceDir   = "~/.local/src/shizuku"
	releaseRepo = "eleonorayaya/dotfiles"
)

var (
	desktopGaps = styles.Gaps{
		InnerHorizontal: 16,
		InnerVertical:   16,
		OuterLeft:       64,
		OuterBottom:     64,
		OuterTop:        64,
		OuterRight:      64,
	}
	laptopGaps = styles.Gaps{
		InnerHorizontal: 8,
		InnerVertical:   8,
		OuterLeft:       8,
		OuterBottom:     8,
		OuterTop:        8,
		OuterRight:      8,
	}
)

func main() {
	cmd := shizuku.New(
		shizuku.WithStyles(styles.New(
			styles.WithTheme(themes.Sylveon),
			styles.WithWindowOpacity(90),
			styles.WithGaps(desktopGaps),
			styles.WithGapOverride("built-in retina", laptopGaps),
		)),
		shizuku.WithPrograms(
			git.New(),
			terminal.New(),
			helix.New(),
			bat.New(),
			lsd.New(),
			tuios.New(),
		),
		shizuku.WithProfile("desktop",
			shizuku.WithLanguages(
				golang.New(),
				lua.New(),
				python.New(),
				rust.New(),
				typescript.New(),
			),
			shizuku.WithPrograms(
				aerospace.New(),
				desktoppr.New(),
				dock.New(),
				fastfetch.New(),
				glow.New(),
				jankyborders.New(),
				kitty.New(),
				notion.New(notion.Options{DisableClaudeMCP: true}),
				nvim.New(),
				sfsymbols.New(),
				sketchybar.New(sketchybar.Options{}),
				terminalbrowser.New(),
				terraform.New(),
				tmux.New(),
				utena.New(),
			),
			shizuku.WithAgents(
				claude.New(data.ClaudeOptions()),
			),
		),
		shizuku.WithProfile("work",
			shizuku.Extends("desktop"),
			shizuku.WithLanguages(
				ruby.New(),
			),
			shizuku.WithPrograms(
				acli.New(),
				buildkite.New(),
				datadog.New(),
				k9s.New(),
				mise.New(),
				sketchybar.New(sketchybar.Options{Badges: []string{"Mail", "Slack"}}),
			),
		),
		shizuku.WithProfile("personal",
			shizuku.Extends("desktop"),
			shizuku.WithLanguages(
				swift.New(),
				zig.New(),
			),
			shizuku.WithPrograms(
				protonpass.New(),
				protonvpn.New(),
				vesktop.New(),
				notion.New(notion.Options{DisableClaudeMCP: false}),
				sketchybar.New(sketchybar.Options{Badges: []string{"Proton Mail", "Vesktop"}}),
			),
		),
	).Command()

	cmd.AddCommand(updateCmd(), sketchybar.Command())

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func updateCmd() *cobra.Command {
	var (
		branch      string
		fromRelease bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update shizuku itself from the source checkout or the latest GitHub release (see 'upgrade' to upgrade packages)",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoDir, err := util.NormalizeFilePath(sourceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve source directory: %w", err)
			}

			if _, err := os.Stat(repoDir); fromRelease || os.IsNotExist(err) {
				slog.Info("updating from latest github release", "repo", releaseRepo)
				if err := pkg.SelfUpdate(releaseRepo, "shizuku"); err != nil {
					return fmt.Errorf("failed to update from release: %w", err)
				}
				slog.Info("update complete, run 'shizuku install' and 'shizuku sync' to apply changes")
				return nil
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

			slog.Info("update complete, run 'shizuku install' and 'shizuku sync' to apply changes")
			return nil
		},
	}
	cmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to pull from")
	cmd.Flags().BoolVar(&fromRelease, "from-release", false, "Download the latest release binary instead of building from source")
	return cmd
}

func runExec(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = dir
	return cmd.Run()
}
