package shizukuconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitConfig(t *testing.T) {
	t.Run("creates config when it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")

		created, returnedPath, err := initConfigAtPath(configPath)
		if err != nil {
			t.Fatalf("InitConfig failed: %v", err)
		}

		if !created {
			t.Error("Expected config to be created")
		}

		if returnedPath != configPath {
			t.Errorf("Expected path %s, got %s", configPath, returnedPath)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Config file was not created")
		}

		config, err := newConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("Failed to load created config: %v", err)
		}

		if config.Languages == nil {
			t.Error("Expected Languages map to be initialized")
		}
	})

	t.Run("does not create config when it already exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")

		existingContent := `languages:
  rust:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(existingContent), 0644); err != nil {
			t.Fatalf("Failed to write existing config: %v", err)
		}

		created, returnedPath, err := initConfigAtPath(configPath)
		if err != nil {
			t.Fatalf("InitConfig failed: %v", err)
		}

		if created {
			t.Error("Expected config not to be created when it already exists")
		}

		if returnedPath != configPath {
			t.Errorf("Expected path %s, got %s", configPath, returnedPath)
		}

		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatalf("Failed to read config: %v", err)
		}

		if string(content) != existingContent {
			t.Error("Existing config was modified")
		}
	})
}

func TestLoadConfigMergesDefaults(t *testing.T) {
	defaults := newConfig()

	t.Run("merges default windowOpacity when missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		configContent := `
styles:
  theme: monade
languages:
  rust:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := loadConfigFromPathWithDefaults(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Styles.WindowOpacity != defaultWindowOpacity {
			t.Errorf("Expected default windowOpacity %d, got %d", defaultWindowOpacity, config.Styles.WindowOpacity)
		}
	})

	t.Run("preserves explicit windowOpacity", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		configContent := `
styles:
  theme: monade
  windowOpacity: 90
languages:
  rust:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := loadConfigFromPathWithDefaults(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Styles.WindowOpacity != 90 {
			t.Errorf("Expected windowOpacity 90, got %d", config.Styles.WindowOpacity)
		}
	})

	t.Run("merges missing languages from defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		configContent := `
styles:
  theme: monade
languages:
  rust:
    enabled: true
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := loadConfigFromPathWithDefaults(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if !config.Languages["rust"].Enabled {
			t.Error("Expected rust to remain enabled")
		}

		for lang := range defaults.Languages {
			if _, exists := config.Languages[lang]; !exists {
				t.Errorf("Expected default language %q to be present after merge", lang)
			}
		}
	})

	t.Run("merges all defaults for minimal config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "shizuku.yml")
		configContent := `
styles:
  theme: monade
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write test config: %v", err)
		}

		config, err := loadConfigFromPathWithDefaults(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if config.Styles.WindowOpacity != defaultWindowOpacity {
			t.Errorf("Expected default windowOpacity %d, got %d", defaultWindowOpacity, config.Styles.WindowOpacity)
		}
		if config.Styles.ThemeName != defaultThemeName {
			t.Errorf("Expected theme %q, got %q", defaultThemeName, config.Styles.ThemeName)
		}
		if config.Styles.Theme == nil {
			t.Error("Expected theme to be loaded")
		}
		for lang := range defaults.Languages {
			if _, exists := config.Languages[lang]; !exists {
				t.Errorf("Expected default language %q to be present", lang)
			}
		}
	})
}

// loadConfigFromPathWithDefaults mirrors LoadConfig but accepts a path for testing.
func loadConfigFromPathWithDefaults(configPath string) (*Config, error) {
	config, err := newConfigFromPath(configPath)
	if err != nil {
		return nil, err
	}

	config.mergeWithDefaults(newConfig())

	return config, nil
}

func initConfigAtPath(configPath string) (bool, string, error) {
	if _, err := os.Stat(configPath); err == nil {
		return false, configPath, nil
	}

	defaultConfig := newConfig()
	if err := defaultConfig.save(configPath); err != nil {
		return false, "", err
	}

	return true, configPath, nil
}
