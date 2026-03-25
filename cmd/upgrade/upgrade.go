package upgrade

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const cloneDir = ".local/src/shizuku"

var branch string

var UpgradeCommand = &cobra.Command{
	Use:   "upgrade",
	Short: "Pull latest changes, rebuild, install, and sync",
	RunE:  upgrade,
}

func init() {
	UpgradeCommand.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to pull from")
}

func upgrade(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	repoDir := filepath.Join(homeDir, cloneDir)

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

	shizukuBin, err := shizukuBinPath()
	if err != nil {
		return fmt.Errorf("failed to resolve shizuku binary path: %w", err)
	}

	slog.Info("running shizuku install")
	if err := run(repoDir, shizukuBin, "install"); err != nil {
		return fmt.Errorf("failed to run shizuku install: %w", err)
	}

	slog.Info("running shizuku sync")
	if err := run(repoDir, shizukuBin, "sync"); err != nil {
		return fmt.Errorf("failed to run shizuku sync: %w", err)
	}

	return nil
}

func shizukuBinPath() (string, error) {
	out, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOPATH: %w", err)
	}
	gopath := strings.TrimSpace(string(out))
	return filepath.Join(gopath, "bin", "shizuku"), nil
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
