package pkg

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type runFunc func(env []string, name string, args ...string) ([]byte, error)

type Installer struct {
	GOOS     string
	GOARCH   string
	Home     string
	IsRoot   bool
	APIBase  string
	Token    string
	HTTP     *http.Client
	lookPath func(string) (string, error)
	run      runFunc
	aptReady bool
}

func NewInstaller() (*Installer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	return &Installer{
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,
		Home:     home,
		IsRoot:   os.Geteuid() == 0,
		APIBase:  "https://api.github.com",
		Token:    os.Getenv("GITHUB_TOKEN"),
		HTTP:     http.DefaultClient,
		lookPath: exec.LookPath,
		run:      runCommand,
	}, nil
}

var defaultInstaller = sync.OnceValues(NewInstaller)

func Install(s Spec) error {
	i, err := defaultInstaller()
	if err != nil {
		return err
	}
	return i.Install(s)
}

func runCommand(env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func (i *Installer) BinDir() string {
	return filepath.Join(i.Home, ".local", "bin")
}

func (i *Installer) installed(bin string) bool {
	if bin == "" {
		return false
	}
	if _, err := i.lookPath(bin); err == nil {
		return true
	}
	_, err := os.Stat(filepath.Join(i.BinDir(), bin))
	return err == nil
}

func (i *Installer) hasApt() bool {
	_, err := i.lookPath("apt-get")
	return err == nil
}

func (i *Installer) Install(s Spec) error {
	if i.installed(s.Bin) {
		slog.Debug("already installed, skipping", "bin", s.Bin)
		return nil
	}

	switch i.GOOS {
	case "darwin":
		if s.Brew == "" {
			return nil
		}
		return i.brewInstall(s.Brew, s.BrewCask)
	case "linux":
		if s.Apt != "" && i.hasApt() {
			ok, err := i.aptHasCandidate(s.Apt)
			if err != nil {
				return err
			}
			if ok {
				return i.aptInstall(s)
			}
		}
		if s.Release != nil {
			return i.releaseInstall(s)
		}
		return fmt.Errorf("no linux install method available for %s", s.name())
	default:
		return fmt.Errorf("unsupported OS %s for %s", i.GOOS, s.name())
	}
}

func (i *Installer) brewInstall(name string, cask bool) error {
	args := []string{"list", name}
	if cask {
		args = append(args, "--cask")
	}
	if _, err := i.run(nil, "brew", args...); err == nil {
		slog.Debug("brew package already installed, skipping", "package", name)
		return nil
	}

	args[0] = "install"
	if _, err := i.run(nil, "brew", args...); err != nil {
		return fmt.Errorf("brew install %s failed: %w", name, err)
	}
	return nil
}

func (i *Installer) privileged(name string, args ...string) ([]byte, error) {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	if i.IsRoot {
		return i.run(env, name, args...)
	}
	return i.run(env, "sudo", append([]string{"DEBIAN_FRONTEND=noninteractive", name}, args...)...)
}

func (i *Installer) aptUpdate() error {
	if i.aptReady {
		return nil
	}
	if _, err := i.privileged("apt-get", "update"); err != nil {
		return fmt.Errorf("apt-get update failed: %w", err)
	}
	i.aptReady = true
	return nil
}

func (i *Installer) aptHasCandidate(name string) (bool, error) {
	if err := i.aptUpdate(); err != nil {
		return false, err
	}
	out, err := i.run(nil, "apt-cache", "policy", name)
	if err != nil {
		return false, fmt.Errorf("apt-cache policy %s failed: %w", name, err)
	}
	return parseCandidate(string(out)), nil
}

func parseCandidate(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Candidate:"); ok {
			v = strings.TrimSpace(v)
			return v != "" && v != "(none)"
		}
	}
	return false
}

func (i *Installer) aptInstall(s Spec) error {
	if _, err := i.privileged("apt-get", "install", "-y", s.Apt); err != nil {
		return fmt.Errorf("apt-get install %s failed: %w", s.Apt, err)
	}
	if s.AptBin != "" && s.Bin != "" && s.AptBin != s.Bin {
		return i.linkBin(s.AptBin, s.Bin)
	}
	return nil
}

func (i *Installer) linkBin(from, to string) error {
	target, err := i.lookPath(from)
	if err != nil {
		return fmt.Errorf("%s not found after install: %w", from, err)
	}
	if err := os.MkdirAll(i.BinDir(), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", i.BinDir(), err)
	}
	link := filepath.Join(i.BinDir(), to)
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing %s: %w", link, err)
	}
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("failed to link %s to %s: %w", link, target, err)
	}
	return nil
}
