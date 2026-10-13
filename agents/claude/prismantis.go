package claude

import (
	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
)

func Prismantis() app.Mod {
	return app.Mod{
		Name:         "prismantis",
		Repo:         "NahumLitvin/prismantis",
		ThemeOptions: prismantisThemeOptions,
	}
}

func prismantisThemeOptions(theme styles.Theme) map[string]any {
	c := theme.Colors
	base := "catppuccin-mocha"
	if theme.Type == "light" {
		base = "catppuccin-latte"
	}

	options := map[string]any{"theme": base}
	colors := map[string]string{
		"accentColor":      c.Accent,
		"headingColor":     c.Primary,
		"strongColor":      c.TextOnSurfaceEmphasis,
		"emphasisColor":    c.Emphasis,
		"inlineCodeColor":  c.InlineCode,
		"codeTextColor":    c.TextOnSurface,
		"codeCommandColor": c.Palette.Blue,
		"codeFlagColor":    c.Palette.Magenta,
		"codeStringColor":  c.String,
		"codeCommentColor": c.Comment,
		"linkColor":        c.Link,
		"pathColor":        c.Path,
		"numberColor":      c.Number,
		"quoteColor":       c.TextOnSurfaceVariant,
		"ruleColor":        c.SurfaceBorder,
		"tableHeaderColor": c.Secondary,
		"tableRuleColor":   c.SurfaceBorder,
		"bulletColor":      c.Secondary,
		"diagramColor":     c.Tertiary,
		"diagramTextColor": c.TextOnSurface,
	}
	for key, color := range colors {
		if color != "" {
			options[key] = color
		}
	}
	return options
}
