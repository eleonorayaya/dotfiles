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
	{file: "~/.zshenv", line: "source ~/.config/shizuku/shizuku.zshenv", marker: "shizuku/shizuku.zshenv"},
	{file: "~/.zshrc", line: "source ~/.config/shizuku/shizuku.sh", marker: "shizuku/shizuku.sh"},
}

var packages = []pkg.Spec{
	{Apt: "zsh", Bin: "zsh"},
	{Apt: "ca-certificates"},
	{Apt: "xz-utils", Bin: "xz"},
	{
		Brew: "jandedobbeleer/oh-my-posh/oh-my-posh",
		Bin:  "oh-my-posh",
		Release: &pkg.Release{
			Repo:  "JanDeDobbeleer/oh-my-posh",
			Asset: `posh-linux-{arch}`,
			Arch:  map[string]string{"amd64": "amd64", "arm64": "arm64"},
		},
	},
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terminal"
}

func (a *App) Install(ctx *app.Context) error {
	for _, spec := range packages {
		if err := pkg.Install(spec); err != nil {
			return fmt.Errorf("failed to install terminal package: %w", err)
		}
	}

	if err := ensureZshViMode(); err != nil {
		return fmt.Errorf("failed to install zsh-vi-mode: %w", err)
	}

	for _, rc := range rcLines {
		path, err := util.NormalizeFilePath(rc.file)
		if err != nil {
			return fmt.Errorf("failed to resolve %s: %w", rc.file, err)
		}
		if err := ensureLine(path, rc.line, rc.marker); err != nil {
			return fmt.Errorf("failed to update %s: %w", rc.file, err)
		}
	}

	if shell := os.Getenv("SHELL"); !strings.HasSuffix(shell, "zsh") {
		slog.Warn("login shell is not zsh; switch with: chsh -s \"$(command -v zsh)\"", "shell", shell)
	}

	return nil
}

func ensureZshViMode() error {
	path, err := util.NormalizeFilePath(zshViModePath)
	if err != nil {
		return fmt.Errorf("failed to resolve zsh-vi-mode path: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		slog.Debug("zsh-vi-mode already installed, skipping")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create plugin dir: %w", err)
	}

	output, err := exec.Command("git", "clone", "--depth", "1", zshViModeRepo, path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone zsh-vi-mode: %w\nOutput: %s", err, string(output))
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
