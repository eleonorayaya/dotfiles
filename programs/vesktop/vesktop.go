package vesktop

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

const themeFile = "shizuku.css"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "vesktop"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("vesktop", true); err != nil {
		return fmt.Errorf("failed to install vesktop: %w", err)
	}

	return nil
}

func (a *App) Generate(ctx *app.Context) (*app.GenerateResult, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user config dir: %w", err)
	}
	destDir := filepath.Join(configDir, "vesktop")
	appOutDir := filepath.Join(ctx.OutDir, "vesktop")

	themeRel := filepath.Join("themes", themeFile)
	themePath := filepath.Join(appOutDir, themeRel)
	if err := util.EnsureDirExists(filepath.Dir(themePath)); err != nil {
		return nil, fmt.Errorf("failed to create themes dir: %w", err)
	}
	if err := os.WriteFile(themePath, []byte(discordTheme(ctx.Styles.Theme, float64(ctx.Styles.WindowOpacity)/100)), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write theme: %w", err)
	}

	settingsRel := filepath.Join("settings", "settings.json")
	settings, err := util.ReadJSONMap(filepath.Join(destDir, settingsRel))
	if err != nil {
		return nil, fmt.Errorf("failed to read vencord settings: %w", err)
	}
	settings["enabledThemes"] = []any{themeFile}
	if ctx.Styles.WindowOpacity < 100 {
		settings["transparent"] = true
		if runtime.GOOS == "darwin" {
			settings["macosVibrancyStyle"] = "under-window"
		}
	}
	settingsPath := filepath.Join(appOutDir, settingsRel)
	if err := util.WriteJSONMap(settingsPath, settings); err != nil {
		return nil, fmt.Errorf("failed to write vencord settings: %w", err)
	}

	return &app.GenerateResult{
		FileMap: map[string]string{
			themeRel:    themePath,
			settingsRel: settingsPath,
		},
		DestDir: destDir,
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
