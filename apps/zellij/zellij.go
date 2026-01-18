package zellij

import (
	"fmt"
	"maps"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

var remotePlugins = map[string]string{
	"plugins/vim-zellij-navigator.wasm": "https://github.com/hiasr/vim-zellij-navigator/releases/latest/download/vim-zellij-navigator.wasm",
	"plugins/zjstatus.wasm":             "https://github.com/dj95/zjstatus/releases/latest/download/zjstatus.wasm",
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := shizukuapp.GenerateAppFiles("zellij", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	pluginMap, err := shizukuapp.FetchRemoteAppFiles(outDir, "zellij", remotePlugins)
	if err != nil {
		return fmt.Errorf("failed to fetch remote plugins: %w", err)
	}

	maps.Copy(fileMap, pluginMap)

	if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/zellij/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		Aliases: []shizukuapp.Alias{
			{Name: "zj", Command: "cd / && zellij -l welcome"},
		},
	}, nil
}
