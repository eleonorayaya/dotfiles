package claude

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
)

func testOptions() Options {
	return Options{
		Marketplaces: map[string]app.Marketplace{
			"test-marketplace": {Repo: "example/test-marketplace"},
		},
		AlwaysOnPlugins:     []string{"always-on@test-marketplace"},
		SandboxAllowedHosts: []string{"default.example.com"},
		SandboxAllowWrite:   []string{"~/default-path"},
	}
}

func TestCollectPluginsEmptyContext(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{}

	plugins := a.collectPlugins(ctx)

	for _, p := range opts.AlwaysOnPlugins {
		if !contains(plugins, p) {
			t.Errorf("expected always-on plugin %q to be present", p)
		}
	}

	if len(plugins) != len(opts.AlwaysOnPlugins) {
		t.Errorf("expected %d plugins, got %d", len(opts.AlwaysOnPlugins), len(plugins))
	}
}

func TestCollectPluginsWithRustConfig(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{Plugins: []string{"rust-analyzer-lsp@claude-plugins-official"}},
		},
	}

	plugins := a.collectPlugins(ctx)

	if !contains(plugins, "rust-analyzer-lsp@claude-plugins-official") {
		t.Error("expected rust-analyzer-lsp to be present when rust config provided")
	}
	if contains(plugins, "gopls-lsp@claude-plugins-official") {
		t.Error("expected gopls-lsp to be absent when go config not provided")
	}
}

func TestCollectPluginsAggregatesAllConfigs(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{Plugins: []string{"gopls-lsp@claude-plugins-official", "charm-dev@charm-dev-skills"}},
			{Plugins: []string{"rust-analyzer-lsp@claude-plugins-official"}},
			{Plugins: []string{"lua-lsp@claude-plugins-official"}},
		},
	}

	plugins := a.collectPlugins(ctx)

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

	expectedCount := len(opts.AlwaysOnPlugins) + len(expected)
	if len(plugins) != expectedCount {
		t.Errorf("expected %d plugins, got %d", expectedCount, len(plugins))
	}
}

func TestCollectSandboxHostsAggregates(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{SandboxAllowedHosts: []string{"crates.io"}},
			{SandboxAllowedHosts: []string{"registry.npmjs.org"}},
		},
	}

	hosts := a.collectSandboxHosts(ctx)

	if !contains(hosts, "crates.io") {
		t.Error("expected crates.io to be present")
	}
	if !contains(hosts, "registry.npmjs.org") {
		t.Error("expected registry.npmjs.org to be present")
	}
	for _, h := range opts.SandboxAllowedHosts {
		if !contains(hosts, h) {
			t.Errorf("expected default host %q to be present", h)
		}
	}
}

func TestCollectSandboxWriteAggregates(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{SandboxAllowWrite: []string{"~/.cargo"}},
			{SandboxAllowWrite: []string{"~/.npm"}},
		},
	}

	paths := a.collectSandboxWrite(ctx)

	if !contains(paths, "~/.cargo") {
		t.Error("expected ~/.cargo to be present")
	}
	if !contains(paths, "~/.npm") {
		t.Error("expected ~/.npm to be present")
	}
	for _, p := range opts.SandboxAllowWrite {
		if !contains(paths, p) {
			t.Errorf("expected default path %q to be present", p)
		}
	}
}

func TestMergeSettingsMarketplaces(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{}

	t.Run("adds all desired marketplaces", func(t *testing.T) {
		outDir := t.TempDir()

		result, err := a.mergeSettings(outDir, ctx)
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

		for name, src := range opts.Marketplaces {
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
			if source["repo"] != src.Repo {
				t.Errorf("expected marketplace %q repo to be %q, got %q", name, src.Repo, source["repo"])
			}
		}
	})

	t.Run("marketplace count matches desired", func(t *testing.T) {
		outDir := t.TempDir()

		result, err := a.mergeSettings(outDir, ctx)
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

		if len(marketplaces) < len(opts.Marketplaces) {
			t.Errorf("expected at least %d marketplaces, got %d", len(opts.Marketplaces), len(marketplaces))
		}
	})
}

func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

func TestMergeHooksAddsNewHook(t *testing.T) {
	result := mergeHooks(nil, []app.Hook{
		{Event: "PreToolUse", Matcher: "Bash", Command: "rtk hook claude"},
	})

	entries, ok := result["PreToolUse"].([]any)
	if !ok || len(entries) != 1 {
		t.Fatalf("expected one PreToolUse entry, got %#v", result["PreToolUse"])
	}
	entry := entries[0].(map[string]any)
	if entry["matcher"] != "Bash" {
		t.Errorf("expected matcher Bash, got %v", entry["matcher"])
	}
	commands := entry["hooks"].([]any)
	if len(commands) != 1 {
		t.Fatalf("expected one command hook, got %d", len(commands))
	}
	cmd := commands[0].(map[string]any)
	if cmd["type"] != "command" || cmd["command"] != "rtk hook claude" {
		t.Errorf("unexpected command hook: %#v", cmd)
	}
}

func TestMergeHooksPreservesExistingAndDedupes(t *testing.T) {
	existing := map[string]any{
		"PreToolUse": []any{
			map[string]any{
				"matcher": "Bash",
				"hooks": []any{
					map[string]any{"type": "command", "command": "rtk hook claude"},
				},
			},
		},
	}

	result := mergeHooks(existing, []app.Hook{
		{Event: "PreToolUse", Matcher: "Bash", Command: "rtk hook claude"},
		{Event: "PreToolUse", Matcher: "Bash", Command: "other-hook"},
		{Event: "Stop", Matcher: "", Command: "stop-hook"},
	})

	preEntries := result["PreToolUse"].([]any)
	if len(preEntries) != 1 {
		t.Fatalf("expected one Bash matcher entry, got %d", len(preEntries))
	}
	commands := preEntries[0].(map[string]any)["hooks"].([]any)
	if len(commands) != 2 {
		t.Fatalf("expected 2 deduped commands, got %d", len(commands))
	}

	stopEntries := result["Stop"].([]any)
	if len(stopEntries) != 1 {
		t.Fatalf("expected one Stop entry, got %d", len(stopEntries))
	}
}

func TestMergeSettingsHooksFromAgentConfig(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{Hooks: []app.Hook{{Event: "PreToolUse", Matcher: "Bash", Command: "rtk hook claude"}}},
		},
	}

	outDir := t.TempDir()
	resultPath, err := a.mergeSettings(outDir, ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("expected hooks to be a map")
	}
	pre, ok := hooks["PreToolUse"].([]any)
	if !ok || len(pre) == 0 {
		t.Fatal("expected PreToolUse entry from agent config")
	}
	matched := false
	for _, entry := range pre {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		if entryMap["matcher"] != "Bash" {
			continue
		}
		commandHooks, _ := entryMap["hooks"].([]any)
		for _, ch := range commandHooks {
			chMap, _ := ch.(map[string]any)
			if chMap["command"] == "rtk hook claude" {
				matched = true
			}
		}
	}
	if !matched {
		t.Errorf("expected rtk hook claude to be merged into settings, got %#v", hooks)
	}
}
