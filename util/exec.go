package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/pkg"
)

func BinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func Run(dir string, name string, args ...string) error {
	_, err := pkg.Run(dir, nil, name, args...)
	return err
}

func CloneOrPull(url, branch, dest string) (string, error) {
	repoPath, err := NormalizeFilePath(dest)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		if err := EnsureDirExists(filepath.Dir(repoPath)); err != nil {
			return "", err
		}
		if err := Run("", "git", "clone", "--depth", "1", "--branch", branch, url, repoPath); err != nil {
			return "", fmt.Errorf("failed to clone %s: %w", url, err)
		}
	} else if err := Run(repoPath, "git", "pull", "origin", branch); err != nil {
		return "", fmt.Errorf("failed to pull %s: %w", url, err)
	}

	return repoPath, nil
}
