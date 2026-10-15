package claude

import (
	"cmp"
	"fmt"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/util"
)

const themeSlug = "shizuku"

const themeSetting = "custom:" + themeSlug

func claudeTheme(theme styles.Theme) map[string]any {
	c := theme.Colors
	p := c.Palette

	base := "dark"
	if theme.Type == "light" {
		base = "light"
	}

	shimmer := func(color string) string { return util.MixHex(color, "#ffffff", 0.3) }
	tint := func(color string, t float64) string { return util.MixHex(c.Surface, color, t) }

	autoMode := cmp.Or(c.TertiaryVariant, c.Warning)

	tokens := map[string]string{
		"claude":                     c.Primary,
		"claudeShimmer":              shimmer(c.Primary),
		"text":                       c.TextOnSurface,
		"inverseText":                c.Surface,
		"inactive":                   c.TextOnSurfaceMuted,
		"inactiveShimmer":            c.TextOnSurfaceVariant,
		"subtle":                     c.SurfaceBorder,
		"suggestion":                 c.Info,
		"permission":                 c.Secondary,
		"permissionShimmer":          shimmer(c.Secondary),
		"remember":                   c.Tertiary,
		"success":                    c.Success,
		"error":                      c.Error,
		"warning":                    autoMode,
		"warningShimmer":             shimmer(autoMode),
		"merged":                     c.Tertiary,
		"promptBorder":               c.TextOnSurfaceMuted,
		"promptBorderShimmer":        c.TextOnSurfaceVariant,
		"planMode":                   c.Info,
		"autoAccept":                 c.Tertiary,
		"bashBorder":                 c.Accent,
		"ide":                        c.Link,
		"fastMode":                   c.Warning,
		"fastModeShimmer":            shimmer(c.Warning),
		"diffAdded":                  tint(c.Added, 0.3),
		"diffRemoved":                tint(c.Removed, 0.3),
		"diffAddedDimmed":            tint(c.Added, 0.15),
		"diffRemovedDimmed":          tint(c.Removed, 0.15),
		"diffAddedWord":              tint(c.Added, 0.55),
		"diffRemovedWord":            tint(c.Removed, 0.55),
		"userMessageBackground":      c.SurfaceVariant,
		"userMessageBackgroundHover": c.SurfaceHighlight,
		"bashMessageBackgroundColor": c.SurfaceVariant,
		"memoryBackgroundColor":      tint(c.Tertiary, 0.1),
		"selectionBg":                c.Selection,
		"rate_limit_fill":            c.Primary,
		"rate_limit_empty":           c.SurfaceBorder,
		"briefLabelYou":              c.Secondary,
		"briefLabelClaude":           c.Primary,
		"red_FOR_SUBAGENTS_ONLY":     p.Red,
		"blue_FOR_SUBAGENTS_ONLY":    p.Blue,
		"green_FOR_SUBAGENTS_ONLY":   p.Green,
		"yellow_FOR_SUBAGENTS_ONLY":  p.Yellow,
		"purple_FOR_SUBAGENTS_ONLY":  p.BrightMagenta,
		"orange_FOR_SUBAGENTS_ONLY":  p.BrightYellow,
		"pink_FOR_SUBAGENTS_ONLY":    p.Magenta,
		"cyan_FOR_SUBAGENTS_ONLY":    p.Cyan,
	}

	overrides := map[string]any{}
	for token, color := range tokens {
		if color != "" {
			overrides[token] = color
		}
	}

	name := theme.Name
	if name == "" {
		name = themeSlug
	}

	return map[string]any{
		"name":      name,
		"base":      base,
		"overrides": overrides,
	}
}

func writeClaudeTheme(outDir string, theme styles.Theme) (string, string, error) {
	fileName := filepath.Join("themes", themeSlug+".json")
	outPath := filepath.Join(outDir, "claude", fileName)
	if err := util.WriteJSONMap(outPath, claudeTheme(theme)); err != nil {
		return "", "", fmt.Errorf("failed to write theme: %w", err)
	}
	return fileName, outPath, nil
}
