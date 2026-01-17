package shizukuconfig_test

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configContent := `
languages:
  rust:
    enabled: true
    version: "1.75"

sketchybar:
  Test: "Aayaya"
  bar_height: 36

aerospace:
  gap_h: 10
`

	configPath := filepath.Join(tmpDir, "shizuku.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load the config
	config, err := LoadConfigFromPath(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test language config
	if !shizukuconfig.IsLanguageEnabled(LanguageRust) {
		t.Error("Expected Rust to be enabled")
	}

	rustConfig, exists := shizukuconfig.GetLanguageConfig(LanguageRust)
	if !exists {
		t.Error("Expected Rust config to exist")
	}

	if !rustConfig.Enabled {
		t.Error("Expected Rust to be enabled")
	}

	// Test app config
	sketchybarConfig := shizukuconfig.GetAppConfig("sketchybar")
	if sketchybarConfig["Test"] != "Aayaya" {
		t.Errorf("Expected Test to be 'Aayaya', got %v", sketchybarConfig["Test"])
	}

	if sketchybarConfig["bar_height"] != 36 {
		t.Errorf("Expected bar_height to be 36, got %v", sketchybarConfig["bar_height"])
	}

	aerospaceConfig := shizukuconfig.GetAppConfig("aerospace")
	if aerospaceConfig["gap_h"] != 10 {
		t.Errorf("Expected gap_h to be 10, got %v", aerospaceConfig["gap_h"])
	}
}

func TestValidateLanguages(t *testing.T) {
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
			name: "no languages section",
			config: `
sketchybar:
  Test: "value"
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

			_, err := LoadConfigFromPath(configPath)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
