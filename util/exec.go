package util

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func Run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}

	return nil
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
		if err := Run("", "git", "clone", "--branch", branch, url, repoPath); err != nil {
			return "", fmt.Errorf("failed to clone %s: %w", url, err)
		}
	} else if err := Run(repoPath, "git", "pull", "origin", branch); err != nil {
		return "", fmt.Errorf("failed to pull %s: %w", url, err)
	}

	return repoPath, nil
}
