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
