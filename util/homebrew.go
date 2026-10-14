package util

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/eleonorayaya/shizuku/pkg"
)

func GetBrewAppPrefix(appName string) (string, error) {
	prefix, err := runBrewCommand("--prefix", appName)
	if err != nil {
		return "", fmt.Errorf("brew --prefix failed: %w", err)
	}

	return prefix, nil
}

func InstallBrewPackage(name string, isCask bool) error {
	return pkg.Install(pkg.Spec{Brew: name, BrewCask: isCask})
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

func EnsureBrewService(name string) (started bool, err error) {
	out, err := runBrewCommand("services", "info", "--json", name)
	if err != nil {
		return false, fmt.Errorf("brew services info %s failed: %w", name, err)
	}

	var services []struct {
		Running bool `json:"running"`
	}
	if err := json.Unmarshal([]byte(out), &services); err != nil {
		return false, fmt.Errorf("failed to parse brew services info for %s: %w", name, err)
	}
	if len(services) > 0 && services[0].Running {
		return false, nil
	}

	slog.Debug("starting brew service", "service", name)
	if _, err := runBrewCommand("services", "start", name); err != nil {
		return false, fmt.Errorf("brew services start %s failed: %w", name, err)
	}

	return true, nil
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
