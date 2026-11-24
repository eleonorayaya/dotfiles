package data

import (
	"slices"
	"strings"
	"testing"
)

func TestCoderClaudeOptionsRegisterBundleMarketplaces(t *testing.T) {
	opts := CoderClaudeOptions()
	plugins := append([]string{}, opts.AlwaysOnPlugins...)
	for _, b := range Bundles() {
		plugins = append(plugins, b.Plugins...)
	}
	for _, p := range plugins {
		_, marketplace, ok := strings.Cut(p, "@")
		if !ok {
			t.Fatalf("plugin %q has no marketplace", p)
		}
		if _, ok := opts.Marketplaces[marketplace]; !ok {
			t.Errorf("marketplace %q for %s is not registered", marketplace, p)
		}
	}
}

func TestCoderClaudeOptionsAreRemote(t *testing.T) {
	opts := CoderClaudeOptions()
	if !opts.Remote {
		t.Error("expected Remote")
	}
	if opts.DisableClaudeAiConnectors || len(opts.GlobalConfig) > 0 || len(opts.SandboxAllowWrite) > 0 || len(opts.SandboxAllowedDomains) > 0 {
		t.Error("expected no connector, global config or sandbox settings")
	}
	for _, c := range opts.AllowedBashCommands {
		if strings.HasPrefix(c, "brew ") {
			t.Errorf("unexpected macOS-only command %q", c)
		}
	}
	if !slices.Contains(opts.AlwaysOnPlugins, "utena-claude@utena") {
		t.Error("expected utena-claude@utena to be always on")
	}
	if !slices.Contains(opts.DeniedBashCommands, "aws:*") {
		t.Error("expected the laptop deny rules to carry over")
	}
}

func TestCoderClaudeOptionsLeaveLaptopOptionsAlone(t *testing.T) {
	CoderClaudeOptions()
	opts := ClaudeOptions()
	if _, ok := opts.Marketplaces["claude-plugins-official"]; ok {
		t.Error("laptop options should get claude-plugins-official from the golang language, not data")
	}
	if !slices.Contains(opts.AllowedBashCommands, "brew --prefix:*") {
		t.Error("laptop options lost brew --prefix")
	}
}
