package claude

import (
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
