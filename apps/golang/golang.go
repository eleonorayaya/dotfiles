package golang

import (
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"strings"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "golang"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetLanguageEnabled(shizukuconfig.LanguageGo)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("go-task", false); err != nil {
		return fmt.Errorf("failed to install go-task: %w", err)
	}

	if !util.BinaryExists("gopls") {
		slog.Debug("installing gopls via go install")

		cmd := exec.Command("go", "install", "golang.org/x/tools/gopls@latest")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install gopls: %w\nOutput: %s", err, string(output))
		}

		slog.Debug("gopls installed successfully")
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	cmd := exec.Command("go", "env", "GOPATH")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get GOPATH: %w", err)
	}

	gopath := strings.TrimSpace(string(output))

	return &shizukuapp.EnvSetup{
		PathDirs: []shizukuapp.PathDir{
			{Path: path.Join(gopath, "bin"), Priority: 20},
		},
	}, nil
}
