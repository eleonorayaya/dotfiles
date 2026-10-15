package claude

import (
	"fmt"
	"math"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/util"
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
		"codeFlagColor":    c.Removed,
		"codeStringColor":  c.String,
		"codeCommentColor": c.Comment,
		"linkColor":        c.Link,
		"pathColor":        c.Path,
		"numberColor":      diffNumberColor(c),
		"quoteColor":       c.TextOnSurfaceVariant,
		"ruleColor":        diffRuleColor(c),
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

const (
	ruleLighten   = 0.6
	prismantisMix = 0.2
)

func diffRuleColor(c styles.ThemeColors) string {
	return util.MixHex(util.GrayHex(c.SurfaceBorder), "#ffffff", ruleLighten)
}

func diffNumberColor(c styles.ThemeColors) string {
	rule, errRule := util.HexToRGB(diffRuleColor(c))
	target, errTarget := util.HexToRGB(diffBackground(c, c.Added))
	if errRule != nil || errTarget != nil {
		return c.Added
	}
	var out [3]int
	for i := range out {
		v := (float64(target[i]) - (1-prismantisMix)*float64(rule[i])) / prismantisMix
		out[i] = int(math.Max(0, math.Min(255, math.Round(v))))
	}
	return fmt.Sprintf("#%02x%02x%02x", out[0], out[1], out[2])
}
