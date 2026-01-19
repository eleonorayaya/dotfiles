package shizukuconfig

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateDefaultStyles(t *testing.T) {
	styles := createDefaultStyles()

	if styles.ThemeName != "monade" {
		t.Errorf("Expected default theme 'monade', got %q", styles.ThemeName)
	}

	if styles.WindowOpacity != 100 {
		t.Errorf("Expected default windowOpacity 100, got %d", styles.WindowOpacity)
	}

	if styles.Theme == nil {
		t.Error("Expected theme to be loaded")
	}
}

func TestValidateStyles(t *testing.T) {
	t.Run("valid styles", func(t *testing.T) {
		styles := Styles{
			ThemeName:     "monade",
			WindowOpacity: 85,
			Theme:         &Theme{},
		}

		if err := styles.validate(); err != nil {
			t.Errorf("Expected valid styles, got error: %v", err)
		}
	})

	t.Run("empty theme name", func(t *testing.T) {
		styles := Styles{
			ThemeName:     "",
			WindowOpacity: 100,
		}

		err := styles.validate()
		if err == nil {
			t.Error("Expected error for empty theme name")
		}
	})

	t.Run("windowOpacity out of range", func(t *testing.T) {
		tests := []int{-1, 101}

		for _, opacity := range tests {
			styles := Styles{
				ThemeName:     "monade",
				WindowOpacity: opacity,
				Theme:         &Theme{},
			}

			err := styles.validate()
			if err == nil {
				t.Errorf("Expected error for windowOpacity=%d", opacity)
			}
		}
	})

	t.Run("windowOpacity at boundaries", func(t *testing.T) {
		tests := []int{0, 100}

		for _, opacity := range tests {
			styles := Styles{
				ThemeName:     "monade",
				WindowOpacity: opacity,
				Theme:         &Theme{},
			}

			if err := styles.validate(); err != nil {
				t.Errorf("Expected windowOpacity=%d to be valid, got error: %v", opacity, err)
			}
		}
	})

	t.Run("nil theme", func(t *testing.T) {
		styles := Styles{
			ThemeName:     "monade",
			WindowOpacity: 100,
			Theme:         nil,
		}

		err := styles.validate()
		if err == nil {
			t.Error("Expected error for nil theme")
		}
	})
}

func TestStylesUnmarshalYAML(t *testing.T) {
	t.Run("loads theme automatically", func(t *testing.T) {
		yamlData := []byte(`
theme: monade
windowOpacity: 85
`)
		var styles Styles
		if err := yaml.Unmarshal(yamlData, &styles); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if styles.ThemeName != "monade" {
			t.Errorf("Expected theme name 'monade', got %q", styles.ThemeName)
		}

		if styles.WindowOpacity != 85 {
			t.Errorf("Expected windowOpacity 85, got %d", styles.WindowOpacity)
		}

		if styles.Theme == nil {
			t.Error("Expected theme to be loaded automatically")
		}
	})

	t.Run("defaults windowOpacity to 0 when missing", func(t *testing.T) {
		yamlData := []byte(`
theme: monade
`)
		var styles Styles
		if err := yaml.Unmarshal(yamlData, &styles); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if styles.WindowOpacity != 0 {
			t.Errorf("Expected windowOpacity 0, got %d", styles.WindowOpacity)
		}
	})

	t.Run("fails for invalid theme", func(t *testing.T) {
		yamlData := []byte(`
theme: invalid-theme
windowOpacity: 100
`)
		var styles Styles
		err := yaml.Unmarshal(yamlData, &styles)
		if err == nil {
			t.Error("Expected error for invalid theme")
		}
	})
}

func TestMergeStyles(t *testing.T) {
	defaultTheme, _ := loadThemeFromRegistry("monade")

	t.Run("preserves existing values", func(t *testing.T) {
		existing := Styles{
			ThemeName:     "monade",
			WindowOpacity: 85,
			Theme:         defaultTheme,
		}
		defaults := createDefaultStyles()

		existing.merge(defaults)

		if existing.WindowOpacity != 85 {
			t.Errorf("Expected windowOpacity 85 to be preserved, got %d", existing.WindowOpacity)
		}
	})

	t.Run("uses default for missing windowOpacity", func(t *testing.T) {
		existing := Styles{
			ThemeName:     "monade",
			WindowOpacity: 0,
			Theme:         defaultTheme,
		}
		defaults := createDefaultStyles()

		existing.merge(defaults)

		if existing.WindowOpacity != 100 {
			t.Errorf("Expected default windowOpacity 100, got %d", existing.WindowOpacity)
		}
	})

	t.Run("uses default theme when missing", func(t *testing.T) {
		existing := Styles{
			ThemeName:     "",
			WindowOpacity: 85,
		}
		defaults := createDefaultStyles()

		existing.merge(defaults)

		if existing.ThemeName != "monade" {
			t.Errorf("Expected default theme 'monade', got %q", existing.ThemeName)
		}
		if existing.Theme == nil {
			t.Error("Expected theme to be set from defaults")
		}
	})
}
