package app

import "github.com/eleonorayaya/shizuku/config"

type Marketplace struct {
	Repo string
	Path string
}

type AgentConfig struct {
	Plugins             []string
	Marketplaces        map[string]Marketplace
	AllowedCommands     []string
	SandboxAllowedHosts []string
	SandboxAllowWrite   []string
}

type AgentConfigProvider interface {
	AgentConfig() AgentConfig
}

type SyncContext struct {
	AgentConfigs []AgentConfig
}

type ContextualSyncer interface {
	SyncWithContext(outDir string, cfg *config.Config, ctx SyncContext) error
}

type ContextualGenerator interface {
	GenerateWithContext(outDir string, cfg *config.Config, ctx SyncContext) (*GenerateResult, error)
}

func CollectAgentConfigs(apps []App) SyncContext {
	configs := []AgentConfig{}
	for _, app := range apps {
		if provider, ok := app.(AgentConfigProvider); ok {
			configs = append(configs, provider.AgentConfig())
		}
	}
	return SyncContext{AgentConfigs: configs}
}
