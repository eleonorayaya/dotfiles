package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
)

func testOptions() Options {
	return Options{
		Marketplaces: map[string]app.Marketplace{
			"test-marketplace": {Repo: "example/test-marketplace"},
		},
		AlwaysOnPlugins:       []string{"always-on@test-marketplace"},
		SandboxAllowedDomains: []string{"default.example.com"},
		SandboxAllowWrite:     []string{"~/default-path"},
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

func TestCollectSandboxDomainsAggregates(t *testing.T) {
	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{SandboxAllowedDomains: []string{"crates.io"}},
			{SandboxAllowedDomains: []string{"registry.npmjs.org"}},
		},
	}

	domains := a.collectSandboxDomains(ctx)

	if !contains(domains, "crates.io") {
		t.Error("expected crates.io to be present")
	}
	if !contains(domains, "registry.npmjs.org") {
		t.Error("expected registry.npmjs.org to be present")
	}
	for _, d := range opts.SandboxAllowedDomains {
		if !contains(domains, d) {
			t.Errorf("expected default domain %q to be present", d)
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

func TestCollectDisabledMcpServersAggregates(t *testing.T) {
	opts := testOptions()
	opts.DisabledMcpJsonServers = []string{"opts-server"}
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{DisabledMcpJsonServers: []string{"notion"}},
			{DisabledMcpJsonServers: []string{"notion"}},
		},
	}

	servers := a.collectDisabledMcpServers(ctx)

	if !contains(servers, "opts-server") {
		t.Error("expected opts-server to be present")
	}
	if !contains(servers, "notion") {
		t.Error("expected notion to be present")
	}
	if len(servers) != 2 {
		t.Errorf("expected 2 deduped servers, got %d: %v", len(servers), servers)
	}
}

func TestMergeSettingsWritesDisabledMcpServers(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{DisabledMcpJsonServers: []string{"notion"}},
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

	disabled, ok := settings["disabledMcpjsonServers"].([]any)
	if !ok {
		t.Fatal("expected disabledMcpjsonServers to be a slice")
	}
	found := false
	for _, d := range disabled {
		if s, _ := d.(string); s == "notion" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected disabledMcpjsonServers to contain notion, got %v", disabled)
	}
}

func TestMergeSettingsOmitsDisabledMcpServersWhenEmpty(t *testing.T) {
	a := New(testOptions())

	outDir := t.TempDir()
	resultPath, err := a.mergeSettings(outDir, app.AgentContext{})
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

	if v, exists := settings["disabledMcpjsonServers"]; exists {
		t.Errorf("expected disabledMcpjsonServers to be omitted when empty, got %#v", v)
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

func TestMergeSettingsWritesAllowedDomainsAndStripsLegacyKey(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	if err := os.MkdirAll(filepath.Join(fakeHome, ".claude"), 0o755); err != nil {
		t.Fatalf("failed to create fake .claude dir: %v", err)
	}
	legacy := []byte(`{"sandbox":{"network":{"allowedHosts":["legacy.example.com"]}}}`)
	if err := os.WriteFile(filepath.Join(fakeHome, ".claude", "settings.json"), legacy, 0o644); err != nil {
		t.Fatalf("failed to write fake settings.json: %v", err)
	}

	opts := testOptions()
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{SandboxAllowedDomains: []string{"crates.io"}},
		},
	}

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

	sandbox, ok := settings["sandbox"].(map[string]any)
	if !ok {
		t.Fatal("expected sandbox to be a map")
	}
	network, ok := sandbox["network"].(map[string]any)
	if !ok {
		t.Fatal("expected sandbox.network to be a map")
	}

	if _, hasLegacy := network["allowedHosts"]; hasLegacy {
		t.Error("legacy sandbox.network.allowedHosts key must be stripped on write")
	}

	domains, ok := network["allowedDomains"].([]any)
	if !ok {
		t.Fatal("expected sandbox.network.allowedDomains to be a slice")
	}
	for _, want := range []string{"crates.io", "default.example.com", "legacy.example.com"} {
		found := false
		for _, d := range domains {
			if s, _ := d.(string); s == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected sandbox.network.allowedDomains to contain %q, got %v", want, domains)
		}
	}
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

func TestCollectAllowedCommandsWrapsBashCommands(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{AllowedBashCommands: []string{"git add:*", "git status:*"}},
		},
	}

	cmds := a.collectAllowedCommands(ctx)

	for _, want := range []string{"Bash(git add:*)", "Bash(git status:*)"} {
		if !contains(cmds, want) {
			t.Errorf("expected %q to be present", want)
		}
	}
}

func TestCollectAllowedCommandsPassesThroughToolPermissions(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{AllowedToolPermissions: []string{"Read(//tmp/**)", "Skill(task)"}},
		},
	}

	cmds := a.collectAllowedCommands(ctx)

	for _, want := range []string{"Read(//tmp/**)", "Skill(task)"} {
		if !contains(cmds, want) {
			t.Errorf("expected %q to be present unchanged", want)
		}
	}
}

func TestCollectAllowedCommandsBaselineAlwaysPresent(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{}

	cmds := a.collectAllowedCommands(ctx)

	if !contains(cmds, "mcp__ide__getDiagnostics") {
		t.Error("expected baseline mcp__ide__getDiagnostics to always be present")
	}
}

func TestCollectAllowedCommandsWithBashPrefix(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{AllowedBashCommands: []string{"git add:*", "grep:*"}},
			{BashCommandPrefix: "rtk"},
		},
	}

	cmds := a.collectAllowedCommands(ctx)

	for _, want := range []string{
		"Bash(git add:*)", "Bash(rtk git add:*)",
		"Bash(grep:*)", "Bash(rtk grep:*)",
	} {
		if !contains(cmds, want) {
			t.Errorf("expected %q to be present", want)
		}
	}
}

func TestCollectAllowedCommandsPrefixDoesNotApplyToToolPermissions(t *testing.T) {
	a := New(testOptions())
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{AllowedToolPermissions: []string{"Read(//tmp/**)"}},
			{BashCommandPrefix: "rtk"},
		},
	}

	cmds := a.collectAllowedCommands(ctx)

	if !contains(cmds, "Read(//tmp/**)") {
		t.Error("expected Read(//tmp/**) to be present")
	}
	if contains(cmds, "Bash(rtk Read(//tmp/**)") {
		t.Error("prefix must not be applied to tool permissions")
	}
}

func TestCollectAllowedCommandsMergesOptsAndAgentConfigs(t *testing.T) {
	opts := testOptions()
	opts.AllowedBashCommands = []string{"grep:*"}
	opts.AllowedToolPermissions = []string{"Write(//tmp/**)"}
	a := New(opts)
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{AllowedBashCommands: []string{"git status:*"}},
		},
	}

	cmds := a.collectAllowedCommands(ctx)

	for _, want := range []string{"Bash(grep:*)", "Bash(git status:*)", "Write(//tmp/**)"} {
		if !contains(cmds, want) {
			t.Errorf("expected %q to be present", want)
		}
	}
}

func TestCollectBashPrefixesDedupes(t *testing.T) {
	ctx := app.AgentContext{
		AgentConfigs: []app.AgentConfig{
			{BashCommandPrefix: "rtk"},
			{BashCommandPrefix: "rtk"},
			{BashCommandPrefix: ""},
		},
	}

	prefixes := collectBashPrefixes(ctx)

	if len(prefixes) != 1 {
		t.Errorf("expected 1 unique prefix, got %d: %v", len(prefixes), prefixes)
	}
	if prefixes[0] != "rtk" {
		t.Errorf("expected prefix rtk, got %q", prefixes[0])
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

func TestCollectDeniedCommandsWrapsBashCommands(t *testing.T) {
	opts := testOptions()
	opts.DeniedBashCommands = []string{"aws", "aws:*"}
	a := New(opts)

	cmds := a.collectDeniedCommands()

	for _, want := range []string{"Bash(aws)", "Bash(aws:*)"} {
		if !contains(cmds, want) {
			t.Errorf("expected %q to be present", want)
		}
	}
}

func TestCollectDeniedCommandsEmpty(t *testing.T) {
	a := New(testOptions())

	cmds := a.collectDeniedCommands()

	if len(cmds) != 0 {
		t.Errorf("expected no denied commands, got %d: %v", len(cmds), cmds)
	}
}

func TestMergeSettingsWritesDeniedCommands(t *testing.T) {
	opts := testOptions()
	opts.DeniedBashCommands = []string{"aws", "aws:*"}
	a := New(opts)

	outDir := t.TempDir()
	resultPath, err := a.mergeSettings(outDir, app.AgentContext{})
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

	permissions, ok := settings["permissions"].(map[string]any)
	if !ok {
		t.Fatal("expected permissions to be a map")
	}
	denyRaw, ok := permissions["deny"].([]any)
	if !ok {
		t.Fatal("expected permissions.deny to be a slice")
	}
	deny := make([]string, 0, len(denyRaw))
	for _, entry := range denyRaw {
		if s, ok := entry.(string); ok {
			deny = append(deny, s)
		}
	}
	for _, want := range []string{"Bash(aws)", "Bash(aws:*)"} {
		if !contains(deny, want) {
			t.Errorf("expected %q in permissions.deny, got %v", want, deny)
		}
	}
}
