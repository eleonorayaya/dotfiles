package monade_dark

type provider struct{}

var Provider = &provider{}

func (p *provider) GetName() string {
	return "monade-dark"
}

func (p *provider) GetType() string {
	return "dark"
}

func (p *provider) GetColors() map[string]string {
	return map[string]string{
		"surface":               "#2a2520",
		"surfaceVariant":        "#332e28",
		"surfaceHighlight":      "#3d2c2a",
		"surfaceBorder":         "#4a3830",
		"textOnSurface":         "#e8dfd5",
		"textOnSurfaceVariant":  "#c4b8ac",
		"textOnSurfaceMuted":    "#8a7873",
		"textOnSurfaceEmphasis": "#fef6f0",
		"primary":               "#e6537a",
		"primaryVariant":        "#d16577",
		"textOnPrimary":         "#fef6f0",
		"secondary":             "#d97c6e",
		"textOnSecondary":       "#2a2520",
		"tertiary":              "#9370b9",
		"tertiaryVariant":       "#a87bb7",
		"textOnTertiary":        "#fef6f0",
		"accent":                "#e6537a",
		"accentPeach":           "#d97c6e",
		"accentSalmon":          "#d16577",
		"accentPurple":          "#9370b9",
		"accentLavender":        "#a87bb7",
		"accentGold":            "#b8733f",
		"accentYellow":          "#b8873f",
		"accentMint":            "#5fafa5",
		"accentBlue":            "#6a9bc3",
		"error":                 "#d16577",
		"textOnError":           "#fef6f0",
		"warning":               "#b8733f",
		"textOnWarning":         "#2a2520",
		"success":               "#9370b9",
		"textOnSuccess":         "#2a2520",
		"info":                  "#6a9bc3",
		"textOnInfo":            "#2a2520",
		"selection":             "#4a2a2a",
		"selectionForeground":   "#e8dfd5",
		"cursor":                "#e6537a",
		"cursorText":            "#2a2520",
		"link":                  "#a87bb7",
		"linkHover":             "#9370b9",
		"comment":               "#8a7873",
	}
}
