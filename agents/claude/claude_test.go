package claude

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
)

func TestCollectPluginsEmptyContext(t *testing.T) {
	ctx := shizukuapp.SyncContext{}

	plugins := collectPlugins(ctx)

	for _, p := range alwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}

	if len(plugins) != len(alwaysOnPlugins) {
		t.Errorf("expected %d plugins, got %d", len(alwaysOnPlugins), len(plugins))
	}
}

func TestCollectPluginsWithRustConfig(t *testing.T) {
	ctx := shizukuapp.SyncContext{
		AgentConfigs: []shizukuapp.AgentConfig{
			{Plugins: []string{"rust-analyzer-lsp@claude-plugins-official"}},
		},
	}

	plugins := collectPlugins(ctx)

	if !contains(plugins, "rust-analyzer-lsp@claude-plugins-official") {
		t.Error("expected rust-analyzer-lsp to be present when rust config provided")
	}
	if contains(plugins, "gopls-lsp@claude-plugins-official") {
		t.Error("expected gopls-lsp to be absent when go config not provided")
	}
}

func TestCollectPluginsAggregatesAllConfigs(t *testing.T) {
	ctx := shizukuapp.SyncContext{
		AgentConfigs: []shizukuapp.AgentConfig{
			{Plugins: []string{"gopls-lsp@claude-plugins-official", "charm-dev@charm-dev-skills"}},
			{Plugins: []string{"rust-analyzer-lsp@claude-plugins-official"}},
			{Plugins: []string{"lua-lsp@claude-plugins-official"}},
		},
	}

	plugins := collectPlugins(ctx)

	expected := []string{
		"gopls-lsp@claude-plugins-official",
		"charm-dev@charm-dev-skills",
		"rust-analyzer-lsp@claude-plugins-official",
		"lua-lsp@claude-plugins-official",
	}
	for _, p := range expected {
		if !contains(plugins, p) {
			t.Errorf("expected plugin %q to be present", p)
		}
	}

	expectedCount := len(alwaysOnPlugins) + len(expected)
	if len(plugins) != expectedCount {
		t.Errorf("expected %d plugins, got %d", expectedCount, len(plugins))
	}
}

func TestCollectSandboxHostsAggregates(t *testing.T) {
	ctx := shizukuapp.SyncContext{
		AgentConfigs: []shizukuapp.AgentConfig{
			{SandboxAllowedHosts: []string{"crates.io"}},
			{SandboxAllowedHosts: []string{"registry.npmjs.org"}},
		},
	}

	hosts := collectSandboxHosts(ctx)

	if !contains(hosts, "crates.io") {
		t.Error("expected crates.io to be present")
	}
	if !contains(hosts, "registry.npmjs.org") {
		t.Error("expected registry.npmjs.org to be present")
	}
	for _, h := range desiredSandboxAllowedHosts {
		if !contains(hosts, h) {
			t.Errorf("expected default host %q to be present", h)
		}
	}
}

func TestCollectSandboxWriteAggregates(t *testing.T) {
	ctx := shizukuapp.SyncContext{
		AgentConfigs: []shizukuapp.AgentConfig{
			{SandboxAllowWrite: []string{"~/.cargo"}},
			{SandboxAllowWrite: []string{"~/.npm"}},
		},
	}

	paths := collectSandboxWrite(ctx)

	if !contains(paths, "~/.cargo") {
		t.Error("expected ~/.cargo to be present")
	}
	if !contains(paths, "~/.npm") {
		t.Error("expected ~/.npm to be present")
	}
	for _, p := range desiredSandboxAllowWrite {
		if !contains(paths, p) {
			t.Errorf("expected default path %q to be present", p)
		}
	}
}

func TestMergeSettingsMarketplaces(t *testing.T) {
	ctx := shizukuapp.SyncContext{}

	t.Run("adds all desired marketplaces", func(t *testing.T) {
		outDir := t.TempDir()

		result, err := mergeSettings(outDir, ctx)
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

		result, err := mergeSettings(outDir, ctx)
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
	return slices.Contains(slice, item)
}
