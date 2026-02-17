package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

func TestLoadInstalledMarketplaces(t *testing.T) {
	t.Run("parses valid known_marketplaces.json", func(t *testing.T) {
		dir := t.TempDir()
		jsonContent := `{
  "claude-plugins-official": {
    "source": {"source": "github", "repo": "anthropics/claude-plugins-official"},
    "installLocation": "/tmp/test/claude-plugins-official",
    "lastUpdated": "2026-02-17T18:35:12.923Z"
  },
  "superpowers-marketplace": {
    "source": {"source": "github", "repo": "obra/superpowers-marketplace"},
    "installLocation": "/tmp/test/superpowers-marketplace",
    "lastUpdated": "2026-02-17T18:52:21.277Z"
  }
}`
		path := filepath.Join(dir, "known_marketplaces.json")
		if err := os.WriteFile(path, []byte(jsonContent), 0644); err != nil {
			t.Fatal(err)
		}

		installed, err := loadInstalledMarketplacesFromPath(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !installed["claude-plugins-official"] {
			t.Error("expected claude-plugins-official to be installed")
		}
		if !installed["superpowers-marketplace"] {
			t.Error("expected superpowers-marketplace to be installed")
		}
		if installed["subtask"] {
			t.Error("expected subtask to not be installed")
		}
		if len(installed) != 2 {
			t.Errorf("expected 2 installed marketplaces, got %d", len(installed))
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := loadInstalledMarketplacesFromPath("/nonexistent/path/known_marketplaces.json")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("returns error for invalid json", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "known_marketplaces.json")
		if err := os.WriteFile(path, []byte("not json"), 0644); err != nil {
			t.Fatal(err)
		}

		_, err := loadInstalledMarketplacesFromPath(path)
		if err == nil {
			t.Error("expected error for invalid json")
		}
	})

	t.Run("handles empty object", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "known_marketplaces.json")
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}

		installed, err := loadInstalledMarketplacesFromPath(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(installed) != 0 {
			t.Errorf("expected 0 installed marketplaces, got %d", len(installed))
		}
	})
}

func TestFilterEnv(t *testing.T) {
	t.Run("removes target variable", func(t *testing.T) {
		env := []string{"HOME=/home/user", "CLAUDECODE=1", "PATH=/usr/bin"}
		result := filterEnv(env, "CLAUDECODE")

		if len(result) != 2 {
			t.Fatalf("expected 2 env vars, got %d", len(result))
		}
		for _, e := range result {
			if e == "CLAUDECODE=1" {
				t.Error("CLAUDECODE should have been filtered out")
			}
		}
	})

	t.Run("preserves all when target absent", func(t *testing.T) {
		env := []string{"HOME=/home/user", "PATH=/usr/bin"}
		result := filterEnv(env, "CLAUDECODE")

		if len(result) != 2 {
			t.Fatalf("expected 2 env vars, got %d", len(result))
		}
	})

	t.Run("does not filter partial matches", func(t *testing.T) {
		env := []string{"CLAUDECODE_OTHER=1", "CLAUDECODE=1"}
		result := filterEnv(env, "CLAUDECODE")

		if len(result) != 1 {
			t.Fatalf("expected 1 env var, got %d", len(result))
		}
		if result[0] != "CLAUDECODE_OTHER=1" {
			t.Errorf("expected CLAUDECODE_OTHER=1, got %s", result[0])
		}
	})
}

func TestGetPluginsWithoutLspPlugins(t *testing.T) {
	config := &shizukuconfig.Config{}

	plugins := getPlugins(config)

	for _, p := range alwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}

	for _, p := range optionalPlugins {
		if contains(plugins, p) {
			t.Errorf("expected optional plugin %q to be absent when lsp_plugins is not set", p)
		}
	}
}

func TestGetPluginsWithLspPluginsEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Apps: map[string]any{
			"claude": map[string]any{
				"lsp_plugins": true,
			},
		},
	}

	plugins := getPlugins(config)

	for _, p := range alwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}

	for _, p := range optionalPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected optional plugin %q to be present when lsp_plugins is true", p)
		}
	}
}

func TestGetPluginsWithLspPluginsExplicitlyDisabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Apps: map[string]any{
			"claude": map[string]any{
				"lsp_plugins": false,
			},
		},
	}

	plugins := getPlugins(config)

	if len(plugins) != len(alwaysOnPlugins) {
		t.Errorf("expected %d plugins, got %d", len(alwaysOnPlugins), len(plugins))
	}

	for _, p := range optionalPlugins {
		if contains(plugins, p) {
			t.Errorf("expected optional plugin %q to be absent when lsp_plugins is false", p)
		}
	}
}

func TestGetPluginsWithCharmDevDisabled(t *testing.T) {
	config := &shizukuconfig.Config{}

	plugins := getPlugins(config)

	if contains(plugins, "charm-dev@charm-dev-skills") {
		t.Error("charm-dev plugin should not be included when charm_dev is false")
	}
}

func TestGetPluginsWithCharmDevEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Apps: map[string]any{
			"claude": map[string]any{
				"charm_dev": true,
			},
		},
	}

	plugins := getPlugins(config)

	if !contains(plugins, "charm-dev@charm-dev-skills") {
		t.Error("charm-dev plugin should be included when charm_dev is true")
	}
}

func TestGetPluginsWithBothEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Apps: map[string]any{
			"claude": map[string]any{
				"lsp_plugins": true,
				"charm_dev":   true,
			},
		},
	}

	plugins := getPlugins(config)

	expected := len(alwaysOnPlugins) + len(optionalPlugins) + 1
	if len(plugins) != expected {
		t.Errorf("expected %d plugins, got %d", expected, len(plugins))
	}

	for _, p := range alwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}
	for _, p := range optionalPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected optional plugin %q to be present", p)
		}
	}
	if !contains(plugins, "charm-dev@charm-dev-skills") {
		t.Error("expected charm-dev plugin to be present")
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
