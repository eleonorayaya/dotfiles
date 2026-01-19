package util

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func GetBrewAppPrefix(appName string) (string, error) {
	prefix, err := runBrewCommand("--prefix", appName)
	if err != nil {
		return "", fmt.Errorf("brew --prefix failed: %w", err)
	}

	return prefix, nil
}

func InstallBrewPackage(name string, isCask bool) error {
	if BrewPackageExists(name, isCask) {
		slog.Debug("brew package already installed, skipping", "package", name)
		return nil
	}

	args := []string{
		"install",
		name,
	}

	if isCask {
		args = append(args, "--cask")
	}

	slog.Debug("installing brew package", "package", name, "isCask", isCask)

	_, err := runBrewCommand(args...)
	if err != nil {
		return fmt.Errorf("brew install %s failed: %w", name, err)
	}

	slog.Debug("brew package installed", "package", name, "isCask", isCask)

	return nil
}

func BrewPackageExists(name string, isCask bool) bool {
	args := []string{
		"list",
		name,
	}

	if isCask {
		args = append(args, "--cask")
	}

	_, err := runBrewCommand(args...)
	return err == nil
}

func AddTap(tapName string) error {
	slog.Debug("adding brew tap", "tap", tapName)

	_, err := runBrewCommand("tap", tapName)
	if err != nil {
		return fmt.Errorf("brew tap %s failed: %w", tapName, err)
	}

	slog.Debug("brew tap added", "tap", tapName)
	return nil
}

func runBrewCommand(args ...string) (string, error) {
	cmd := exec.Command("brew", args...)
	out, err := cmd.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("Command failed with stderr: %s\n", string(exitError.Stderr))
		} else {
			return "", err
		}
	}

	return strings.TrimSpace(string(out)), nil
}
