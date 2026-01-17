package shizukuconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDefaultLanguageConfig(t *testing.T) {
	langConfig := createDefaultLanguageConfig()

	if langConfig == nil {
		t.Fatal("Expected non-nil language config")
	}

	if len(langConfig) == 0 {
		t.Error("Expected at least one language in default config")
	}

	for lang, config := range langConfig {
		if config.Enabled {
			t.Errorf("Expected language %s to be disabled by default", lang)
		}

		if config.Config == nil {
			t.Errorf("Expected Config map to be initialized for language %s", lang)
		}
	}

	if _, exists := langConfig["rust"]; !exists {
		t.Error("Expected rust to be in default language config")
	}
}

func TestValidateLanguageConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    string
		shouldErr bool
	}{
		{
			name: "valid rust language",
			config: `
languages:
  rust:
    enabled: true
`,
			shouldErr: false,
		},
		{
			name: "invalid language",
			config: `
languages:
  python:
    enabled: true
`,
			shouldErr: true,
		},
		{
			name:      "no languages section",
			config:    ``,
			shouldErr: false,
		},
		{
			name: "nil languages map",
			config: `
other_config: value
`,
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "shizuku.yml")
			if err := os.WriteFile(configPath, []byte(tt.config), 0644); err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			_, err := newConfigFromPath(configPath)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestValidateLanguageConfigDirect(t *testing.T) {
	t.Run("accepts valid languages", func(t *testing.T) {
		langConfig := map[string]LanguageConfig{
			"rust": {Enabled: true},
		}

		if err := validateLanguageConfig(langConfig); err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("rejects invalid languages", func(t *testing.T) {
		langConfig := map[string]LanguageConfig{
			"python": {Enabled: true},
		}

		if err := validateLanguageConfig(langConfig); err == nil {
			t.Error("Expected error for invalid language")
		}
	})

	t.Run("accepts nil config", func(t *testing.T) {
		if err := validateLanguageConfig(nil); err != nil {
			t.Errorf("Expected nil config to be valid, got error: %v", err)
		}
	})

	t.Run("accepts empty config", func(t *testing.T) {
		langConfig := map[string]LanguageConfig{}

		if err := validateLanguageConfig(langConfig); err != nil {
			t.Errorf("Expected empty config to be valid, got error: %v", err)
		}
	})
}
