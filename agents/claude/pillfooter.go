package claude

import (
	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
)

func PillFooter() app.Mod {
	return app.Mod{
		Name:         "pill-footer",
		Marketplace:  "eleonorayaya-claude-code",
		Repo:         "eleonorayaya/claude-plugins",
		ThemeOptions: pillFooterThemeOptions,
	}
}

func pillFooterThemeOptions(theme styles.Theme) map[string]any {
	c := theme.Colors
	colors := map[string]string{
		"modelText":         c.TextOnSurface,
		"modelBackground":   c.SurfaceBorder,
		"contextBackground": c.SurfaceVariant,
		"barColor":          c.PrimaryVariant,
		"trackColor":        c.SurfaceBorder,
	}
	options := map[string]any{}
	for key, color := range colors {
		if color != "" {
			options[key] = color
		}
	}
	return options
}
