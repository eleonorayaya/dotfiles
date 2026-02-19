package claude

import (
	"encoding/json"
	"os"
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

func TestMergeSettingsMarketplaces(t *testing.T) {
	config := &shizukuconfig.Config{
		Languages: shizukuconfig.LanguageConfigs{},
	}

	t.Run("adds all desired marketplaces", func(t *testing.T) {
		outDir := t.TempDir()

		result, err := mergeSettings(outDir, config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}

		marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]any)
		if !ok {
			t.Fatal("expected extraKnownMarketplaces to be a map")
		}

		for name, src := range desiredMarketplaces {
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
			if source["repo"] != src.repo {
				t.Errorf("expected marketplace %q repo to be %q, got %q", name, src.repo, source["repo"])
			}
		}
	})

	t.Run("marketplace count matches desired", func(t *testing.T) {
		outDir := t.TempDir()

		result, err := mergeSettings(outDir, config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(result)
		if err != nil {
			t.Fatalf("failed to read output: %v", err)
		}

		var settings map[string]any
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}

		marketplaces, ok := settings["extraKnownMarketplaces"].(map[string]any)
		if !ok {
			t.Fatal("expected extraKnownMarketplaces to be a map")
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
