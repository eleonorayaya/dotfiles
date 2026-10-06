package terminal

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const zshViModeRepo = "https://github.com/jeffreytse/zsh-vi-mode"

const zshViModePath = "~/.local/share/shizuku/plugins/zsh-vi-mode"

const zshViModePlugin = "zsh-vi-mode.plugin.zsh"

const zshViModeInit = `if [[ -f ~/.local/share/shizuku/plugins/zsh-vi-mode/zsh-vi-mode.plugin.zsh ]]; then
  source ~/.local/share/shizuku/plugins/zsh-vi-mode/zsh-vi-mode.plugin.zsh
fi`

const ohmyposhInit = `if command -v oh-my-posh >/dev/null 2>&1; then
  eval "$(oh-my-posh init zsh --config ~/.config/ohmyposh/ohmyposh.json)"
fi`

type rcLine struct {
	file   string
	line   string
	marker string
}

var rcLines = []rcLine{
	{file: "~/.zshenv", line: "[[ -r ~/.config/shizuku/shizuku.zshenv ]] && source ~/.config/shizuku/shizuku.zshenv", marker: "shizuku/shizuku.zshenv"},
	{file: "~/.zshrc", line: "[[ -r ~/.config/shizuku/shizuku.sh ]] && source ~/.config/shizuku/shizuku.sh", marker: "shizuku/shizuku.sh"},
}

type terminalPackage struct {
	spec         pkg.Spec
	skipIfExists string
}

var packages = []terminalPackage{
	{spec: pkg.Spec{Apt: "zsh", Bin: "zsh"}},
	{spec: pkg.Spec{Apt: "ca-certificates"}, skipIfExists: "/etc/ssl/certs/ca-certificates.crt"},
	{spec: pkg.Spec{Apt: "xz-utils", Bin: "xz"}},
	{spec: pkg.Spec{
		Brew: "jandedobbeleer/oh-my-posh/oh-my-posh",
		Bin:  "oh-my-posh",
		Release: &pkg.Release{
			Repo:  "JanDeDobbeleer/oh-my-posh",
			Asset: `posh-linux-{arch}`,
			Arch:  map[string]string{"amd64": "amd64", "arm64": "arm64"},
		},
	}},
}

var installPkg = pkg.Install

var cloneRepo = func(repo, dest string) error {
	output, err := exec.Command("git", "clone", "--depth", "1", repo, dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %s: %w\nOutput: %s", repo, err, string(output))
	}
	return nil
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terminal"
}

func (a *App) Install(ctx *app.Context) error {
	var errs []error

	for _, rc := range rcLines {
		path, err := util.NormalizeFilePath(rc.file)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to resolve %s: %w", rc.file, err))
			continue
		}
		if err := ensureLine(path, rc.line, rc.marker); err != nil {
			errs = append(errs, fmt.Errorf("failed to update %s: %w", rc.file, err))
		}
	}

	if err := installPackages(packages); err != nil {
		errs = append(errs, err)
	}

	if err := ensureZshViMode(); err != nil {
		slog.Warn("failed to install zsh-vi-mode", "error", err)
	}

	if shell := os.Getenv("SHELL"); !strings.HasSuffix(shell, "zsh") {
		slog.Warn("login shell is not zsh; switch with: chsh -s \"$(command -v zsh)\"", "shell", shell)
	}

	return errors.Join(errs...)
}

func installPackages(pkgs []terminalPackage) error {
	var errs []error
	for _, p := range pkgs {
		if p.skipIfExists != "" {
			if _, err := os.Stat(p.skipIfExists); err == nil {
				continue
			}
		}
		if err := installPkg(p.spec); err != nil {
			errs = append(errs, fmt.Errorf("failed to install terminal package: %w", err))
		}
	}
	return errors.Join(errs...)
}

func zshViModeDir() (string, error) {
	path, err := util.NormalizeFilePath(zshViModePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve zsh-vi-mode path: %w", err)
	}
	return path, nil
}

func ensureZshViMode() error {
	path, err := zshViModeDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath.Join(path, zshViModePlugin)); err == nil {
		slog.Debug("zsh-vi-mode already installed, skipping")
		return nil
	}

	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove incomplete zsh-vi-mode dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create plugin dir: %w", err)
	}

	if err := cloneRepo(zshViModeRepo, path); err != nil {
		if rmErr := os.RemoveAll(path); rmErr != nil {
			return errors.Join(fmt.Errorf("failed to clone zsh-vi-mode: %w", err), fmt.Errorf("failed to clean up %s: %w", path, rmErr))
		}
		return fmt.Errorf("failed to clone zsh-vi-mode: %w", err)
	}

	return nil
}

func ensureLine(path, line, marker string) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	if bytes.Contains(data, []byte(marker)) {
		return nil
	}

	prefix := ""
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		prefix = "\n"
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(prefix + line + "\n"); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	data := map[string]any{
		"Colors": ctx.Styles.Theme.Colors,
	}

	fileMap, err := app.GenerateAppFiles("terminal", contents, data, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.config/ohmyposh/",
	}, nil
}

func (a *App) Sync(ctx *app.Context) error {
	result, err := a.Generate(ctx)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		PathDirs: []app.PathDir{
			{Path: "$HOME/.local/bin", Priority: 5},
		},
		InitScripts: []string{zshViModeInit, ohmyposhInit},
		Aliases: []app.Alias{
			{Name: "c", Command: "clear"},
			{Name: "curltime", Command: "curl -o /dev/null -s -w 'Total: %{time_total}s\\n'"},
		},
		Functions: []app.ShellFunction{
			{Name: "colormap", Body: colormapFunction},
		},
	}, nil
}

const colormapFunction = `    for i in {0..255}; do
        printf "\x1b[38;5;${i}mcolour${i}\x1b[0m\n"
    done`
