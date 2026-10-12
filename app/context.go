package app

import (
	"maps"

	"github.com/eleonorayaya/shizuku/styles"
)

type Marketplace struct {
	Repo string
	Path string
}

type Mod struct {
	Name         string
	Marketplace  string
	Repo         string
	Path         string
	Options      map[string]any
	ThemeOptions func(styles.Theme) map[string]any
}

func (m Mod) ResolveOptions(theme styles.Theme) map[string]any {
	resolved := map[string]any{}
	if m.ThemeOptions != nil {
		maps.Copy(resolved, m.ThemeOptions(theme))
	}
	maps.Copy(resolved, m.Options)
	return resolved
}

func (m Mod) MarketplaceName() string {
	if m.Marketplace != "" {
		return m.Marketplace
	}
	return m.Name
}

func (m Mod) PluginID() string {
	return m.Name + "@" + m.MarketplaceName()
}

type Hook struct {
	Event   string
	Matcher string
	Command string
}

type AgentConfig struct {
	Plugins                 []string
	Marketplaces            map[string]Marketplace
	DisabledMcpJsonServers  []string
	AllowedBashCommands     []string
	AllowedToolPermissions  []string
	SandboxAllowedDomains   []string
	SandboxAllowRead        []string
	SandboxAllowWrite       []string
	SandboxExcludedCommands []string
	Hooks                   []Hook
	BashCommandPrefix       string
}

type AgentConfigProvider interface {
	AgentConfig() AgentConfig
}

type AgentContext struct {
	AgentConfigs []AgentConfig
}

func CollectAgentConfigs(providers []AgentConfigProvider) AgentContext {
	configs := make([]AgentConfig, 0, len(providers))
	for _, p := range providers {
		configs = append(configs, p.AgentConfig())
	}
	return AgentContext{AgentConfigs: configs}
}
