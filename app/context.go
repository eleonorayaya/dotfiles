package app

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
