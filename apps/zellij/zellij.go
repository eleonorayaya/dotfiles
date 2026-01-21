package zellij

import (
	"fmt"
	"maps"
	"strconv"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
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
	if err := util.InstallBrewPackage("zellij", false); err != nil {
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

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{
		"ThemeName":            config.Styles.Theme.Name,
		"Surface":              hexToRGB(config.Styles.Theme.Colors.Surface),
		"SurfaceVariant":       hexToRGB(config.Styles.Theme.Colors.SurfaceVariant),
		"SurfaceHighlight":     hexToRGB(config.Styles.Theme.Colors.SurfaceHighlight),
		"TextOnSurface":        hexToRGB(config.Styles.Theme.Colors.TextOnSurface),
		"TextOnSurfaceVariant": hexToRGB(config.Styles.Theme.Colors.TextOnSurfaceVariant),
		"TextOnSurfaceMuted":   hexToRGB(config.Styles.Theme.Colors.TextOnSurfaceMuted),
		"Primary":              hexToRGB(config.Styles.Theme.Colors.Primary),
		"AccentSalmon":         hexToRGB(config.Styles.Theme.Colors.AccentSalmon),
		"AccentBlue":           hexToRGB(config.Styles.Theme.Colors.AccentBlue),
		"AccentMint":           hexToRGB(config.Styles.Theme.Colors.AccentMint),
		"AccentLavender":       hexToRGB(config.Styles.Theme.Colors.AccentLavender),
		"AccentPeach":          hexToRGB(config.Styles.Theme.Colors.AccentPeach),
		"AccentGold":           hexToRGB(config.Styles.Theme.Colors.AccentGold),
		"AccentPurple":          hexToRGB(config.Styles.Theme.Colors.AccentPurple),
		"HexPrimary":            config.Styles.Theme.Colors.Primary,
		"HexAccentPeach":        config.Styles.Theme.Colors.AccentPeach,
		"HexAccentMint":         config.Styles.Theme.Colors.AccentMint,
		"HexAccentSalmon":       config.Styles.Theme.Colors.AccentSalmon,
		"HexTextOnSurfaceMuted": config.Styles.Theme.Colors.TextOnSurfaceMuted,
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
