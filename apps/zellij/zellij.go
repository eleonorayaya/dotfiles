package zellij

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/shizukustyle"
	"github.com/eleonorayaya/shizuku/internal/util"
)

var remotePlugins = map[string]string{
	"plugins/vim-zellij-navigator.wasm": "https://github.com/hiasr/vim-zellij-navigator/releases/latest/download/vim-zellij-navigator.wasm",
	"plugins/zjstatus.wasm":             "https://github.com/dj95/zjstatus/releases/latest/download/zjstatus.wasm",
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "zellij"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Install(config *shizukuconfig.Config) error {
	if err := util.InstallBrewPackage("zellij"); err != nil {
		return fmt.Errorf("failed to install zellij: %w", err)
	}

	return nil
}

func hexToRGB(hex string) string {
	hex = hex[1:]
	r, _ := strconv.ParseInt(hex[0:2], 16, 64)
	g, _ := strconv.ParseInt(hex[2:4], 16, 64)
	b, _ := strconv.ParseInt(hex[4:6], 16, 64)
	return fmt.Sprintf("%d %d %d", r, g, b)
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config, styles *shizukustyle.Styles) error {
	data := map[string]any{
		"ThemeName":            styles.Theme.Name,
		"Surface":              hexToRGB(styles.Theme.Colors.Surface),
		"SurfaceVariant":       hexToRGB(styles.Theme.Colors.SurfaceVariant),
		"SurfaceHighlight":     hexToRGB(styles.Theme.Colors.SurfaceHighlight),
		"TextOnSurface":        hexToRGB(styles.Theme.Colors.TextOnSurface),
		"TextOnSurfaceVariant": hexToRGB(styles.Theme.Colors.TextOnSurfaceVariant),
		"TextOnSurfaceMuted":   hexToRGB(styles.Theme.Colors.TextOnSurfaceMuted),
		"Primary":              hexToRGB(styles.Theme.Colors.Primary),
		"AccentSalmon":         hexToRGB(styles.Theme.Colors.AccentSalmon),
		"AccentBlue":           hexToRGB(styles.Theme.Colors.AccentBlue),
		"AccentMint":           hexToRGB(styles.Theme.Colors.AccentMint),
		"AccentLavender":       hexToRGB(styles.Theme.Colors.AccentLavender),
		"AccentPeach":          hexToRGB(styles.Theme.Colors.AccentPeach),
		"AccentGold":           hexToRGB(styles.Theme.Colors.AccentGold),
		"AccentPurple":         hexToRGB(styles.Theme.Colors.AccentPurple),
	}

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
