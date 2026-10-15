package pkg

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func (i *Installer) upgrade(s Spec) error {
	switch i.GOOS {
	case "darwin":
		if s.Brew == "" {
			return nil
		}
		return i.brewUpgrade(s.Brew, s.BrewCask)
	case "linux":
		if s.Release != nil && i.releaseInstalled(s.Bin) {
			return i.releaseInstall(s)
		}
		if s.Apt != "" && i.hasApt() {
			return i.aptUpgrade(s.Apt)
		}
		slog.Debug("no upgradable install found, skipping", "package", s.name())
		return nil
	default:
		return fmt.Errorf("unsupported OS %s for %s", i.GOOS, s.name())
	}
}

func (i *Installer) releaseInstalled(bin string) bool {
	if bin == "" {
		return false
	}
	info, err := os.Lstat(filepath.Join(i.BinDir(), bin))
	return err == nil && info.Mode().IsRegular()
}

func (i *Installer) brewUpdate() error {
	if i.brewReady {
		return nil
	}
	if i.brewErr != nil {
		return i.brewErr
	}
	if _, err := i.run(nil, "brew", "update"); err != nil {
		i.brewErr = fmt.Errorf("brew update failed: %w", err)
		return i.brewErr
	}
	i.brewReady = true
	return nil
}

func (i *Installer) brewUpgrade(name string, cask bool) error {
	args := []string{"list", name}
	if cask {
		args = append(args, "--cask")
	}
	if _, err := i.run(nil, "brew", args...); err != nil {
		slog.Debug("brew package not installed, skipping upgrade", "package", name)
		return nil
	}

	if err := i.brewUpdate(); err != nil {
		return err
	}

	args[0] = "upgrade"
	if _, err := i.run([]string{"HOMEBREW_NO_AUTO_UPDATE=1"}, "brew", args...); err != nil {
		return fmt.Errorf("brew upgrade %s failed: %w", name, err)
	}
	slog.Debug("brew package upgraded", "package", name)
	return nil
}

func (i *Installer) aptUpgrade(name string) error {
	if err := i.aptUpdate(); err != nil {
		return err
	}
	out, err := i.run([]string{"LC_ALL=C"}, "apt-cache", "policy", name)
	if err != nil {
		return fmt.Errorf("apt-cache policy %s failed: %w", name, err)
	}
	if !parseInstalled(string(out)) {
		slog.Debug("apt package not installed, skipping upgrade", "package", name)
		return nil
	}
	if _, err := i.privileged("apt-get", "install", "--only-upgrade", "-y", name); err != nil {
		return fmt.Errorf("apt-get upgrade %s failed: %w", name, err)
	}
	return nil
}

func parseInstalled(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Installed:"); ok {
			v = strings.TrimSpace(v)
			return v != "" && v != "(none)"
		}
	}
	return false
}
