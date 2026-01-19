package theme

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/theme/monade"
)

type ThemeProvider interface {
	GetName() string
	GetType() string
	GetColors() map[string]string
}

var providers = map[string]ThemeProvider{
	"monade": monade.Provider,
}

func LoadTheme(themeName string) (*Theme, error) {
	provider, exists := providers[themeName]
	if !exists {
		return nil, fmt.Errorf("unknown theme '%s'", themeName)
	}

	colors := convertColors(provider.GetColors())
	if err := colors.Validate(); err != nil {
		return nil, fmt.Errorf("theme '%s' validation failed: %w", themeName, err)
	}

	return &Theme{
		Name:   provider.GetName(),
		Type:   provider.GetType(),
		Colors: colors,
	}, nil
}

func convertColors(m map[string]string) ThemeColors {
	return ThemeColors{
		Surface:              m["surface"],
		SurfaceVariant:       m["surfaceVariant"],
		SurfaceHighlight:     m["surfaceHighlight"],
		SurfaceBorder:        m["surfaceBorder"],
		TextOnSurface:        m["textOnSurface"],
		TextOnSurfaceVariant: m["textOnSurfaceVariant"],
		TextOnSurfaceMuted:   m["textOnSurfaceMuted"],
		TextOnSurfaceEmphasis: m["textOnSurfaceEmphasis"],
		Primary:              m["primary"],
		PrimaryVariant:       m["primaryVariant"],
		TextOnPrimary:        m["textOnPrimary"],
		Secondary:            m["secondary"],
		TextOnSecondary:      m["textOnSecondary"],
		Tertiary:             m["tertiary"],
		TertiaryVariant:      m["tertiaryVariant"],
		TextOnTertiary:       m["textOnTertiary"],
		Accent:               m["accent"],
		AccentPeach:          m["accentPeach"],
		AccentSalmon:         m["accentSalmon"],
		AccentPurple:         m["accentPurple"],
		AccentLavender:       m["accentLavender"],
		AccentGold:           m["accentGold"],
		AccentYellow:         m["accentYellow"],
		AccentMint:           m["accentMint"],
		AccentBlue:           m["accentBlue"],
		Error:                m["error"],
		TextOnError:          m["textOnError"],
		Warning:              m["warning"],
		TextOnWarning:        m["textOnWarning"],
		Success:              m["success"],
		TextOnSuccess:        m["textOnSuccess"],
		Info:                 m["info"],
		TextOnInfo:           m["textOnInfo"],
		Selection:            m["selection"],
		SelectionForeground:  m["selectionForeground"],
		Cursor:               m["cursor"],
		CursorText:           m["cursorText"],
		Link:                 m["link"],
		LinkHover:            m["linkHover"],
		Comment:              m["comment"],
	}
}
