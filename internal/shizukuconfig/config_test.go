package shizukuconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := newConfig()

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.Languages == nil {
		t.Error("Expected Languages map to be initialized")
	}

	if len(config.Languages) == 0 {
		t.Error("Expected Languages map to have default languages")
	}

	for lang, langConfig := range config.Languages {
		if langConfig.Enabled {
			t.Errorf("Expected language %s to be disabled by default", lang)
		}
	}
}

func TestNewConfigFromPath(t *testing.T) {
	t.Run("loads valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
styles:
  theme: monade
languages:
  rust:
    enabled: true
    version: "1.75"
`
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := newConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("newConfigFromPath failed: %v", err)
		}

		rustConfig, exists := config.Languages["rust"]
		if !exists {
			t.Error("Expected Rust config to exist")
		}

		if !rustConfig.Enabled {
			t.Error("Expected Rust to be enabled")
		}

		if rustConfig.Config["version"] != "1.75" {
			t.Errorf("Expected version to be '1.75', got %v", rustConfig.Config["version"])
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nonexistent.yml")

		_, err := newConfigFromPath(configPath)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yml")
		if err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := newConfigFromPath(configPath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})

	t.Run("loads config with windowOpacity", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
styles:
  theme: monade
  windowOpacity: 85
languages:
  rust:
    enabled: true
`
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := newConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("newConfigFromPath failed: %v", err)
		}

		if config.Styles.WindowOpacity != 85 {
			t.Errorf("Expected windowOpacity 85, got %d", config.Styles.WindowOpacity)
		}

		if config.Styles.ThemeName != "monade" {
			t.Errorf("Expected theme name 'monade', got %q", config.Styles.ThemeName)
		}

		if config.Styles.Theme == nil {
			t.Error("Expected theme to be loaded automatically")
		}
	})

	t.Run("merges default windowOpacity when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
styles:
  theme: monade
languages:
  rust:
    enabled: true
`
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := newConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("newConfigFromPath failed: %v", err)
		}

		defaults := newConfig()
		config.mergeWithDefaults(defaults)

		if config.Styles.WindowOpacity != 100 {
			t.Errorf("Expected default windowOpacity 100, got %d", config.Styles.WindowOpacity)
		}
	})

	t.Run("rejects invalid windowOpacity", func(t *testing.T) {
		tmpDir := t.TempDir()
		configContent := `
styles:
  theme: monade
  windowOpacity: 150
languages:
  rust:
    enabled: true
`
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		_, err := newConfigFromPath(configPath)
		if err == nil {
			t.Error("Expected error for windowOpacity out of range")
		}
	})
}

func TestConfigSave(t *testing.T) {
	t.Run("saves config to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test.yml")

		config := newConfig()
		if err := config.save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		loadedConfig, err := newConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}

		if loadedConfig.Languages == nil {
			t.Error("Expected Languages map in loaded config")
		}
	})

	t.Run("creates directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nested", "dir", "config.yml")

		config := newConfig()
		if err := config.save(configPath); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created in nested directory")
		}
	})
}

func TestConfigValidate(t *testing.T) {
	t.Run("validates correct config", func(t *testing.T) {
		config := newConfig()
		if err := config.validate(); err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("rejects invalid language", func(t *testing.T) {
		theme, _ := loadThemeFromRegistry("monade")
		config := &Config{
			Styles: Styles{
				ThemeName:     "monade",
				WindowOpacity: 100,
				Theme:         theme,
			},
			Languages: LanguageConfigs{
				"invalid-lang": {Enabled: true},
			},
		}

		if err := config.validate(); err == nil {
			t.Error("Expected error for invalid language")
		}
	})
}
