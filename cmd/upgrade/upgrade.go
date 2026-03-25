package upgrade

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
	"github.com/spf13/cobra"
)

var branch string

var UpgradeCommand = &cobra.Command{
	Use:   "upgrade",
	Short: "Pull latest changes and rebuild the shizuku binary",
	RunE:  upgrade,
}

func init() {
	UpgradeCommand.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to pull from")
}

func upgrade(cmd *cobra.Command, args []string) error {
	repoDir, err := util.NormalizeFilePath(shizukuconfig.SourceDir)
	if err != nil {
		return fmt.Errorf("failed to resolve source directory: %w", err)
	}

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		return fmt.Errorf("shizuku repo not found at %s, run 'shizuku install' first", repoDir)
	}

	slog.Info("pulling latest changes", "branch", branch)
	if err := run(repoDir, "git", "fetch", "origin", branch); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	if err := run(repoDir, "git", "checkout", branch); err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}
	if err := run(repoDir, "git", "pull", "origin", branch); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	slog.Info("building and installing shizuku")
	if err := run(repoDir, "task", "install"); err != nil {
		return fmt.Errorf("failed to build and install: %w", err)
	}

	slog.Info("upgrade complete, run 'shizuku install' and 'shizuku sync' to apply changes")

	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Run()
}
