package monade

import "github.com/eleonorayaya/shizuku/internal/theme"

func init() {
	theme.Register("monade", LoadTheme)
}

func LoadTheme() *theme.Theme {
	return &theme.Theme{
		Name: "monade",
		Type: "light",
		Colors: theme.ThemeColors{
			Surface:               "#e8dfd5",
			SurfaceVariant:        "#dfd6cc",
			SurfaceHighlight:      "#edd5d0",
			SurfaceBorder:         "#e8d0c4",
			TextOnSurface:         "#4a4a4a",
			TextOnSurfaceVariant:  "#6a6a6a",
			TextOnSurfaceMuted:    "#8a7873",
			TextOnSurfaceEmphasis: "#2a2a2a",
			Primary:               "#e6537a",
			PrimaryVariant:        "#d16577",
			TextOnPrimary:         "#fef6f0",
			Secondary:             "#d97c6e",
			TextOnSecondary:       "#2a2a2a",
			Tertiary:              "#9370b9",
			TertiaryVariant:       "#a87bb7",
			TextOnTertiary:        "#fef6f0",
			Accent:                "#e6537a",
			AccentPeach:           "#d97c6e",
			AccentSalmon:          "#d16577",
			AccentPurple:          "#9370b9",
			AccentLavender:        "#a87bb7",
			AccentGold:            "#b8733f",
			AccentYellow:          "#b8873f",
			AccentMint:            "#5fafa5",
			AccentBlue:            "#6a9bc3",
			Error:                 "#d16577",
			TextOnError:           "#fef6f0",
			Warning:               "#b8733f",
			TextOnWarning:         "#2a2a2a",
			Success:               "#5fafa5",
			TextOnSuccess:         "#2a2a2a",
			Info:                  "#6a9bc3",
			TextOnInfo:            "#2a2a2a",
			Selection:             "#ffd5d5",
			SelectionForeground:   "#4a4a4a",
			Cursor:                "#e6537a",
			CursorText:            "#fef6f0",
			Link:                  "#a87bb7",
			LinkHover:             "#9370b9",
			Comment:               "#8a7873",
		},
	}
}
