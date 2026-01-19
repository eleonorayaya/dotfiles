package shizukustyle

import (
	"encoding/json"
	"testing"
)

func TestThemeColorsValidation(t *testing.T) {
	t.Run("validates complete theme colors", func(t *testing.T) {
		colors := ThemeColors{
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
		}

		if err := colors.Validate(); err != nil {
			t.Errorf("Expected valid colors, got error: %v", err)
		}
	})

	t.Run("rejects theme missing surface color", func(t *testing.T) {
		colors := ThemeColors{
			SurfaceVariant: "#dfd6cc",
		}

		err := colors.Validate()
		if err == nil {
			t.Error("Expected validation error for missing surface color")
		}
		if err.Error() != "missing required color: surface" {
			t.Errorf("Expected 'missing required color: surface', got: %v", err)
		}
	})

	t.Run("rejects theme missing primary color", func(t *testing.T) {
		colors := ThemeColors{
			Surface:               "#e8dfd5",
			SurfaceVariant:        "#dfd6cc",
			SurfaceHighlight:      "#edd5d0",
			SurfaceBorder:         "#e8d0c4",
			TextOnSurface:         "#4a4a4a",
			TextOnSurfaceVariant:  "#6a6a6a",
			TextOnSurfaceMuted:    "#8a7873",
			TextOnSurfaceEmphasis: "#2a2a2a",
		}

		err := colors.Validate()
		if err == nil {
			t.Error("Expected validation error for missing primary color")
		}
		if err.Error() != "missing required color: primary" {
			t.Errorf("Expected 'missing required color: primary', got: %v", err)
		}
	})

	t.Run("parses complete theme JSON correctly", func(t *testing.T) {
		jsonData := `{
			"name": "test",
			"type": "light",
			"colors": {
				"surface": "#e8dfd5",
				"surfaceVariant": "#dfd6cc",
				"surfaceHighlight": "#edd5d0",
				"surfaceBorder": "#e8d0c4",
				"textOnSurface": "#4a4a4a",
				"textOnSurfaceVariant": "#6a6a6a",
				"textOnSurfaceMuted": "#8a7873",
				"textOnSurfaceEmphasis": "#2a2a2a",
				"primary": "#e6537a",
				"primaryVariant": "#d16577",
				"textOnPrimary": "#fef6f0",
				"secondary": "#d97c6e",
				"textOnSecondary": "#2a2a2a",
				"tertiary": "#9370b9",
				"tertiaryVariant": "#a87bb7",
				"textOnTertiary": "#fef6f0",
				"accent": "#e6537a",
				"accentPeach": "#d97c6e",
				"accentSalmon": "#d16577",
				"accentPurple": "#9370b9",
				"accentLavender": "#a87bb7",
				"accentGold": "#b8733f",
				"accentYellow": "#b8873f",
				"accentMint": "#5fafa5",
				"accentBlue": "#6a9bc3",
				"error": "#d16577",
				"textOnError": "#fef6f0",
				"warning": "#b8733f",
				"textOnWarning": "#2a2a2a",
				"success": "#5fafa5",
				"textOnSuccess": "#2a2a2a",
				"info": "#6a9bc3",
				"textOnInfo": "#2a2a2a",
				"selection": "#ffd5d5",
				"selectionForeground": "#4a4a4a",
				"cursor": "#e6537a",
				"cursorText": "#fef6f0",
				"link": "#a87bb7",
				"linkHover": "#9370b9",
				"comment": "#8a7873"
			}
		}`

		var theme Theme
		if err := json.Unmarshal([]byte(jsonData), &theme); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if theme.Name != "test" {
			t.Errorf("Expected name 'test', got '%s'", theme.Name)
		}

		if theme.Colors.Surface != "#e8dfd5" {
			t.Errorf("Expected surface color '#e8dfd5', got '%s'", theme.Colors.Surface)
		}

		if err := theme.Colors.Validate(); err != nil {
			t.Errorf("Expected valid theme, got error: %v", err)
		}
	})

	t.Run("rejects incomplete theme JSON", func(t *testing.T) {
		jsonData := `{
			"name": "incomplete",
			"type": "light",
			"colors": {
				"surface": "#ffffff",
				"surfaceVariant": "#eeeeee"
			}
		}`

		var theme Theme
		if err := json.Unmarshal([]byte(jsonData), &theme); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		err := theme.Colors.Validate()
		if err == nil {
			t.Error("Expected validation error for incomplete theme")
		}
	})
}
