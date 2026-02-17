package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

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

func TestMergeMarketplaces(t *testing.T) {
	t.Run("adds missing marketplaces to existing file", func(t *testing.T) {
		outDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(outDir, "claude", "plugins"), 0755); err != nil {
			t.Fatal(err)
		}

		result, err := mergeMarketplaces(outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var marketplaces map[string]any
		if err := json.Unmarshal(data, &marketplaces); err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}

		for name := range desiredMarketplaces {
			entry, exists := marketplaces[name]
			if !exists {
				t.Errorf("expected marketplace %q to be present", name)
				continue
			}
			entryMap, ok := entry.(map[string]any)
			if !ok {
				t.Errorf("expected marketplace %q to be a map", name)
				continue
			}
			source, ok := entryMap["source"].(map[string]any)
			if !ok {
				t.Errorf("expected marketplace %q source to be a map", name)
				continue
			}
			if source["repo"] != desiredMarketplaces[name] {
				t.Errorf("expected marketplace %q repo to be %q, got %q", name, desiredMarketplaces[name], source["repo"])
			}
		}
	})

	t.Run("preserves existing marketplace entries", func(t *testing.T) {
		outDir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(outDir, "claude", "plugins"), 0755); err != nil {
			t.Fatal(err)
		}

		result, err := mergeMarketplaces(outDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var marketplaces map[string]any
		if err := json.Unmarshal(data, &marketplaces); err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}

		// Existing marketplaces from the real file should be preserved
		for name := range desiredMarketplaces {
			if _, exists := marketplaces[name]; !exists {
				t.Errorf("expected marketplace %q to be present", name)
			}
		}

		// Total count should be at least as many as desired (existing + desired)
		if len(marketplaces) < len(desiredMarketplaces) {
			t.Errorf("expected at least %d marketplaces, got %d", len(desiredMarketplaces), len(marketplaces))
		}
	})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
