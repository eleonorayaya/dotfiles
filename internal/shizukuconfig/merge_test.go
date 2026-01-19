package shizukuconfig

import (
	"testing"

	"github.com/eleonorayaya/shizuku/internal/util"
	"gopkg.in/yaml.v3"
)

func TestMergeLanguageConfigs(t *testing.T) {
	t.Run("adds new language from defaults", func(t *testing.T) {
		existing := LanguageConfigs{
			"rust": {Enabled: true, Config: map[string]any{"version": "1.75"}},
		}
		defaults := LanguageConfigs{
			"rust": {Enabled: false, Config: map[string]any{"version": "1.70"}},
		}

		existing.merge(defaults)

		if len(existing) != 1 {
			t.Errorf("Expected 1 language, got %d", len(existing))
		}

		if !existing["rust"].Enabled {
			t.Error("Expected Rust to remain enabled")
		}
		if existing["rust"].Config["version"] != "1.75" {
			t.Error("Expected Rust version to remain 1.75")
		}
	})

	t.Run("preserves existing language enabled state", func(t *testing.T) {
		existing := LanguageConfigs{
			"rust": {Enabled: true, Config: make(map[string]any)},
		}
		defaults := LanguageConfigs{
			"rust": {Enabled: false, Config: make(map[string]any)},
		}

		existing.merge(defaults)

		if !existing["rust"].Enabled {
			t.Error("Expected existing Enabled=true to be preserved")
		}
	})

	t.Run("adds missing config keys from defaults", func(t *testing.T) {
		existing := LanguageConfigs{
			"rust": {Enabled: true, Config: map[string]any{"version": "1.75"}},
		}
		defaults := LanguageConfigs{
			"rust": {Enabled: false, Config: map[string]any{"version": "1.70", "newKey": "newValue"}},
		}

		existing.merge(defaults)

		if existing["rust"].Config["version"] != "1.75" {
			t.Error("Expected existing version to be preserved")
		}
		if existing["rust"].Config["newKey"] != "newValue" {
			t.Error("Expected new key to be added from defaults")
		}
	})
}

func TestMergeStringAnyMap(t *testing.T) {
	t.Run("merges nested maps", func(t *testing.T) {
		existing := map[string]any{
			"app1": map[string]any{
				"enabled":  true,
				"setting1": "custom",
			},
		}
		defaults := map[string]any{
			"app1": map[string]any{
				"enabled":  false,
				"setting1": "default",
				"setting2": "new",
			},
			"app2": map[string]any{
				"enabled": false,
			},
		}

		result := util.MergeStringAnyMap(existing, defaults)

		app1 := result["app1"].(map[string]any)
		if app1["enabled"] != true {
			t.Error("Expected app1.enabled to keep existing value true")
		}
		if app1["setting1"] != "custom" {
			t.Error("Expected app1.setting1 to keep existing value")
		}
		if app1["setting2"] != "new" {
			t.Error("Expected app1.setting2 to be added from defaults")
		}

		if result["app2"] == nil {
			t.Error("Expected app2 to be added from defaults")
		}
	})

	t.Run("preserves non-map values", func(t *testing.T) {
		existing := map[string]any{
			"key1": "value1",
			"key2": 42,
		}
		defaults := map[string]any{
			"key1": "default1",
			"key3": "default3",
		}

		result := util.MergeStringAnyMap(existing, defaults)

		if result["key1"] != "value1" {
			t.Error("Expected existing string value to be preserved")
		}
		if result["key2"] != 42 {
			t.Error("Expected existing int value to be preserved")
		}
		if result["key3"] != "default3" {
			t.Error("Expected new key from defaults to be added")
		}
	})

	t.Run("handles nil existing map", func(t *testing.T) {
		defaults := map[string]any{
			"key1": "value1",
		}

		result := util.MergeStringAnyMap(nil, defaults)

		if result["key1"] != "value1" {
			t.Error("Expected default to be used when existing is nil")
		}
	})

	t.Run("handles nil defaults map", func(t *testing.T) {
		existing := map[string]any{
			"key1": "value1",
		}

		result := util.MergeStringAnyMap(existing, nil)

		if result["key1"] != "value1" {
			t.Error("Expected existing to be returned when defaults is nil")
		}
	})
}

func TestConfigMergeWithDefaults(t *testing.T) {
	t.Run("preserves existing theme", func(t *testing.T) {
		customTheme, _ := loadThemeFromRegistry("monade")
		defaultTheme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "custom-theme",
				WindowOpacity: 85,
				Theme:         customTheme,
			},
			Languages: make(LanguageConfigs),
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         defaultTheme,
			},
			Languages: make(LanguageConfigs),
		}

		existing.mergeWithDefaults(defaults)

		if existing.Styles.ThemeName != "custom-theme" {
			t.Error("Expected existing theme to be preserved")
		}
		if existing.Styles.WindowOpacity != 85 {
			t.Error("Expected existing windowOpacity to be preserved")
		}
	})

	t.Run("uses default theme when empty", func(t *testing.T) {
		defaultTheme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "",
				WindowOpacity: 85,
			},
			Languages: make(LanguageConfigs),
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         defaultTheme,
			},
			Languages: make(LanguageConfigs),
		}

		existing.mergeWithDefaults(defaults)

		if existing.Styles.ThemeName != "monade" {
			t.Error("Expected default theme when existing is empty")
		}
		if existing.Styles.WindowOpacity != 85 {
			t.Error("Expected existing windowOpacity to be preserved")
		}
	})

	t.Run("merges windowOpacity from defaults when zero", func(t *testing.T) {
		yamlData := []byte(`
styles:
  theme: monade
languages:
  rust:
    enabled: true
`)
		var existing Config
		if err := yaml.Unmarshal(yamlData, &existing); err != nil {
			t.Fatalf("Failed to unmarshal YAML: %v", err)
		}

		defaults := newConfig()

		existing.mergeWithDefaults(defaults)

		if existing.Styles.WindowOpacity != 100 {
			t.Errorf("Expected windowOpacity to be added from defaults (100), got %d", existing.Styles.WindowOpacity)
		}
		if !existing.Languages["rust"].Enabled {
			t.Error("Expected Rust to remain enabled from existing config")
		}
	})

	t.Run("merges languages", func(t *testing.T) {
		theme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: LanguageConfigs{
				"rust": {Enabled: true, Config: make(map[string]any)},
			},
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: LanguageConfigs{
				"rust": {Enabled: false, Config: make(map[string]any)},
			},
		}

		existing.mergeWithDefaults(defaults)

		if !existing.Languages["rust"].Enabled {
			t.Error("Expected Rust to remain enabled after merge")
		}
	})

	t.Run("merges apps", func(t *testing.T) {
		theme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: make(LanguageConfigs),
			Apps: map[string]any{
				"nvim": map[string]any{
					"enabled": true,
				},
			},
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: make(LanguageConfigs),
			Apps: map[string]any{
				"nvim": map[string]any{
					"enabled":    false,
					"newSetting": "value",
				},
				"kitty": map[string]any{
					"enabled": false,
				},
			},
		}

		existing.mergeWithDefaults(defaults)

		nvimConfig := existing.Apps["nvim"].(map[string]any)
		if nvimConfig["enabled"] != true {
			t.Error("Expected nvim.enabled to remain true")
		}
		if nvimConfig["newSetting"] != "value" {
			t.Error("Expected newSetting to be added from defaults")
		}
		if existing.Apps["kitty"] == nil {
			t.Error("Expected kitty to be added from defaults")
		}
	})

	t.Run("handles nil languages", func(t *testing.T) {
		theme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: nil,
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: LanguageConfigs{
				"rust": {Enabled: false, Config: make(map[string]any)},
			},
		}

		existing.mergeWithDefaults(defaults)

		if existing.Languages["rust"].Enabled != false {
			t.Error("Expected default languages to be used when existing is nil")
		}
	})

	t.Run("handles nil apps", func(t *testing.T) {
		theme, _ := loadThemeFromRegistry("monade")
		existing := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: make(LanguageConfigs),
			Apps:      nil,
		}
		defaults := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: make(LanguageConfigs),
			Apps: map[string]any{
				"nvim": map[string]any{
					"enabled": false,
				},
			},
		}

		existing.mergeWithDefaults(defaults)

		if existing.Apps["nvim"] == nil {
			t.Error("Expected default apps to be used when existing is nil")
		}
	})
}
