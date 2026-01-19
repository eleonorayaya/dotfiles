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

func InstallBrewPackage(packageName string) error {
	exists, err := BrewPackageExists(packageName)
	if err != nil {
		return fmt.Errorf("failed to check if package exists: %w", err)
	}

	if exists {
		slog.Debug("brew package already installed, skipping", "package", packageName)
		return nil
	}

	slog.Debug("installing brew package", "package", packageName)

	_, err = runBrewCommand("install", packageName)
	if err != nil {
		return fmt.Errorf("brew install %s failed: %w", packageName, err)
	}

	slog.Debug("brew package installed", "package", packageName)

	return nil
}

func BrewPackageExists(packageName string) (bool, error) {
	_, err := runBrewCommand("list", packageName)
	if err != nil {
		return false, fmt.Errorf("failed to check if brew package exists: %w", err)
	}

	return true, nil
}

func runBrewCommand(args ...string) (string, error) {
	out, err := exec.Command("brew", args...).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

