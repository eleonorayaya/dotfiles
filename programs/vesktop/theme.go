package vesktop

import (
	"fmt"
	"strings"

	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/util"
)

type cssVar struct {
	name  string
	value string
}

var brandSteps = []struct {
	step   int
	amount float64
}{
	{100, 0.45}, {130, 0.425}, {160, 0.40}, {200, 0.375}, {230, 0.35}, {260, 0.325},
	{300, 0.30}, {330, 0.25}, {360, 0.20}, {400, 0.15}, {430, 0.10}, {460, 0.05},
	{500, 0}, {530, -0.05}, {560, -0.10}, {600, -0.15}, {630, -0.20}, {660, -0.25},
	{700, -0.30}, {730, -0.325}, {760, -0.35}, {800, -0.375}, {830, -0.40}, {860, -0.425},
	{900, -0.45},
}

func alpha(hex string, a float64) string {
	if hex == "" {
		return ""
	}
	return fmt.Sprintf("%s%02x", hex, int(a*255+0.5))
}

func shade(hex string, amount float64) string {
	if amount >= 0 {
		return util.MixHex(hex, "#ffffff", amount)
	}
	return util.MixHex(hex, "#000000", -amount)
}

func discordTheme(theme styles.Theme, opacity float64) string {
	c := theme.Colors
	light := theme.Type == "light"

	frame := "#000000"
	if light {
		frame = c.SurfaceBorder
	}
	base := c.Surface
	mantle := util.MixHex(base, frame, 0.1)
	crust := util.MixHex(base, frame, 0.2)
	surface0 := c.SurfaceVariant
	surface1 := c.SurfaceHighlight
	surface2 := c.SurfaceBorder
	brand := c.Primary

	vars := []cssVar{}
	for _, s := range brandSteps {
		vars = append(vars, cssVar{fmt.Sprintf("brand-%d", s.step), shade(brand, s.amount)})
	}
	for i := 5; i <= 95; i += 5 {
		vars = append(vars, cssVar{fmt.Sprintf("brand-%02da", i), alpha(brand, float64(i)/100)})
	}

	vars = append(vars, []cssVar{
		{"__header-bar-background", mantle},
		{"text-default", c.TextOnSurface},
		{"text-normal", c.TextOnSurface},
		{"text-muted", c.TextOnSurfaceMuted},
		{"text-link", c.Link},
		{"text-brand", brand},
		{"text-strong", c.TextOnSurfaceEmphasis},
		{"text-subtle", c.TextOnSurfaceVariant},
		{"header-primary", c.TextOnSurfaceEmphasis},
		{"header-secondary", c.TextOnSurfaceVariant},
		{"interactive-normal", c.TextOnSurfaceVariant},
		{"interactive-hover", c.TextOnSurface},
		{"interactive-active", c.TextOnSurfaceEmphasis},
		{"interactive-muted", c.TextOnSurfaceMuted},
		{"channels-default", c.TextOnSurfaceVariant},
		{"text-feedback-positive", c.Success},
		{"text-feedback-critical", c.Error},
		{"text-feedback-warning", c.Warning},
		{"text-feedback-info", c.Info},
		{"mention-foreground", brand},
		{"mention-background", alpha(brand, 0.15)},
		{"message-reacted-background-default", alpha(brand, 0.3)},
		{"message-reacted-text-default", brand},
		{"opacity-blurple-8", alpha(brand, 0.08)},
		{"opacity-blurple-16", alpha(brand, 0.16)},
		{"opacity-blurple-24", alpha(brand, 0.24)},
		{"opacity-blurple-32", alpha(brand, 0.32)},
		{"opacity-blurple-60", alpha(brand, 0.60)},
		{"blurple-50", brand},
		{"blurple-60", shade(brand, -0.10)},
		{"app-frame-background", crust},
		{"background-primary", base},
		{"background-secondary", mantle},
		{"background-secondary-alt", mantle},
		{"background-tertiary", crust},
		{"background-floating", mantle},
		{"background-accent", surface1},
		{"background-surface-highest", surface1},
		{"background-surface-higher", surface0},
		{"background-surface-high", base},
		{"background-base-lowest", crust},
		{"background-base-lower", mantle},
		{"background-base-low", base},
		{"bg-surface-raised", mantle},
		{"background-gradient-highest", base},
		{"home-background", base},
		{"chat-background", base},
		{"chat-background-default", base},
		{"chat-border", crust},
		{"chat-text-muted", c.TextOnSurfaceMuted},
		{"channeltextarea-background", surface0},
		{"modal-background", base},
		{"modal-footer-background", mantle},
		{"input-background", surface0},
		{"checkbox-icon-active", base},
		{"checkbox-border-default", c.TextOnSurfaceMuted},
		{"radio-thumb-background-active", base},
		{"border-muted", surface1},
		{"border-strong", surface2},
		{"border-normal", surface2},
		{"border-subtle", surface1},
		{"background-modifier-hover", alpha(surface2, 0.25)},
		{"background-modifier-active", alpha(surface2, 0.35)},
		{"background-modifier-selected", alpha(surface2, 0.45)},
		{"background-mod-muted", alpha(surface2, 0.05)},
		{"background-mod-normal", alpha(surface2, 0.15)},
		{"background-mod-subtle", alpha(surface2, 0.25)},
		{"background-mod-strong", alpha(surface2, 0.45)},
		{"custom-channel-members-bg", mantle},
		{"custom-status-bubble-background", crust},
		{"custom-status-bubble-background-color", mantle},
		{"card-background-default", surface0},
		{"scrollbar-thin-thumb", surface2},
		{"scrollbar-auto-thumb", surface2},
		{"scrollbar-auto-track", "transparent"},
		{"background-feedback-positive", alpha(c.Success, 0.15)},
		{"icon-feedback-positive", c.Success},
		{"background-feedback-warning", alpha(c.Warning, 0.15)},
		{"icon-feedback-warning", c.Warning},
		{"background-feedback-critical", alpha(c.Error, 0.15)},
		{"icon-feedback-critical", c.Error},
		{"background-feedback-info", alpha(c.Info, 0.15)},
		{"icon-feedback-info", c.Info},
		{"background-feedback-notification", c.Error},
		{"icon-feedback-notification", c.Error},
		{"badge-notification-background", c.Error},
		{"status-positive", c.Success},
		{"status-positive-background", c.Success},
		{"status-positive-text", c.TextOnSuccess},
		{"status-warning", c.Warning},
		{"status-warning-background", c.Warning},
		{"status-warning-text", c.TextOnWarning},
		{"status-danger", c.Error},
		{"green-360", c.Success},
		{"green-300", c.Success},
		{"yellow-360", c.Warning},
		{"yellow-300", c.Warning},
		{"red-400", c.Error},
		{"red-430", shade(c.Error, -0.10)},
		{"red-500", shade(c.Error, -0.15)},
		{"textbox-markdown-syntax", c.TextOnSurfaceMuted},
	}...)

	if light {
		vars = append(vars, []cssVar{
			{"white", base},
			{"white-500", base},
			{"black-500", c.TextOnSurface},
		}...)
	} else {
		vars = append(vars, cssVar{"black-500", crust})
	}

	if opacity < 1 {
		frameBg := alpha(base, opacity)
		if light {
			frameBg = alpha(util.MixHex(base, "#ffffff", 0.8), opacity)
		}
		panel := func(string) string { return "transparent" }
		overrides := map[string]string{
			"app-frame-background":        frameBg,
			"background-base-lowest":      frameBg,
			"background-tertiary":         frameBg,
			"background-base-lower":       panel(mantle),
			"background-secondary":        panel(mantle),
			"background-secondary-alt":    panel(mantle),
			"bg-surface-raised":           panel(mantle),
			"custom-channel-members-bg":   panel(mantle),
			"__header-bar-background":     panel(mantle),
			"background-base-low":         panel(base),
			"background-primary":          panel(base),
			"background-surface-high":     panel(base),
			"background-gradient-highest": panel(base),
			"home-background":             panel(base),
			"chat-background":             panel(base),
			"chat-background-default":     panel(base),
		}
		for i, v := range vars {
			if o, ok := overrides[v.name]; ok {
				vars[i].value = o
			}
		}
	}

	var b strings.Builder
	b.WriteString("/**\n * @name Shizuku\n * @description Generated by shizuku from the " + theme.Name + " theme\n * @author shizuku\n */\n\n")
	b.WriteString(".visual-refresh.theme-light,\n.visual-refresh .theme-light,\n.visual-refresh.theme-dark,\n.visual-refresh .theme-dark {\n")
	for _, v := range vars {
		if v.value == "" {
			continue
		}
		fmt.Fprintf(&b, "  --%s: %s !important;\n", v.name, v.value)
	}
	b.WriteString("}\n")
	if opacity < 1 {
		b.WriteString("\nhtml,\nbody,\n#app-mount {\n  background: transparent !important;\n}\n")
	}
	if light {
		b.WriteString("\ndiv[class*=\"tooltipGrey_\"] {\n")
		fmt.Fprintf(&b, "  --white: %s !important;\n  --white-500: %s !important;\n", c.TextOnSurface, c.TextOnSurface)
		b.WriteString("}\n")
	}
	return b.String()
}
