package zellij

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
)

var remotePlugins = map[string]string{
	"plugins/vim-zellij-navigator.wasm": "https://github.com/hiasr/vim-zellij-navigator/releases/latest/download/vim-zellij-navigator.wasm",
	"plugins/zjstatus.wasm":             "https://github.com/dj95/zjstatus/releases/latest/download/zjstatus.wasm",
}

func Sync(outDir string) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("zellij", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	pluginMap, err := internal.FetchRemoteAppFiles(outDir, "zellij", remotePlugins)
	if err != nil {
		return fmt.Errorf("failed to fetch remote plugins: %w", err)
	}

	for fileName, filePath := range pluginMap {
		fileMap[fileName] = filePath
	}

	if err := internal.SyncAppFiles(fileMap, "~/.config/zellij/"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
