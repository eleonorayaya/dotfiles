package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

func TestGetPluginsNoLanguagesEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Languages: shizukuconfig.LanguageConfigs{},
	}

	plugins := getPlugins(config)

	for _, p := range alwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}

	for _, langPlugins := range languagePlugins {
		for _, plugin := range langPlugins {
			if contains(plugins, plugin) {
				t.Errorf("expected language plugin %q to be absent when no languages enabled", plugin)
			}
		}
	}
}

func TestGetPluginsWithRustEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Languages: shizukuconfig.LanguageConfigs{
			"rust": {Enabled: true},
		},
	}

	plugins := getPlugins(config)

	if !contains(plugins, "rust-analyzer-lsp@claude-plugins-official") {
		t.Error("expected rust-analyzer-lsp to be present when rust is enabled")
	}
	if contains(plugins, "gopls-lsp@claude-plugins-official") {
		t.Error("expected gopls-lsp to be absent when go is not enabled")
	}
}

func TestGetPluginsWithGoEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Languages: shizukuconfig.LanguageConfigs{
			"go": {Enabled: true},
		},
	}

	plugins := getPlugins(config)

	if !contains(plugins, "gopls-lsp@claude-plugins-official") {
		t.Error("expected gopls-lsp to be present when go is enabled")
	}
	if !contains(plugins, "charm-dev@charm-dev-skills") {
		t.Error("expected charm-dev to be present when go is enabled")
	}
}

func TestGetPluginsWithAllLanguagesEnabled(t *testing.T) {
	config := &shizukuconfig.Config{
		Languages: shizukuconfig.LanguageConfigs{
			"go":         {Enabled: true},
			"lua":        {Enabled: true},
			"rust":       {Enabled: true},
			"typescript": {Enabled: true},
		},
	}

	plugins := getPlugins(config)

	for _, langPlugins := range languagePlugins {
		for _, plugin := range langPlugins {
			if !contains(plugins, plugin) {
				t.Errorf("expected language plugin %q to be present when all languages enabled", plugin)
			}
		}
	}

	expectedCount := len(alwaysOnPlugins)
	for _, langPlugins := range languagePlugins {
		expectedCount += len(langPlugins)
	}
	if len(plugins) != expectedCount {
		t.Errorf("expected %d plugins, got %d", expectedCount, len(plugins))
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

		for name := range desiredMarketplaces {
			if _, exists := marketplaces[name]; !exists {
				t.Errorf("expected marketplace %q to be present", name)
			}
		}

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
