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

const (
	repoURL  = "https://github.com/eleonorayaya/dotfiles"
	cloneDir = ".local/src/shizuku"
)

var UpgradeCommand = &cobra.Command{
	Use:   "upgrade",
	Short: "Pull latest changes, rebuild, install, and sync",
	RunE:  upgrade,
}

func upgrade(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	repoDir := filepath.Join(homeDir, cloneDir)

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		slog.Info("cloning shizuku repo")
		if err := os.MkdirAll(filepath.Dir(repoDir), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		if err := run("", "git", "clone", repoURL, repoDir); err != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
	} else {
		slog.Info("pulling latest changes")
		if err := run(repoDir, "git", "pull", "origin", "main"); err != nil {
			return fmt.Errorf("failed to pull latest changes: %w", err)
		}
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
