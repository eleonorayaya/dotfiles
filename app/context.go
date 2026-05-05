package app

type Marketplace struct {
	Repo string
	Path string
}

type Hook struct {
	Event   string
	Matcher string
	Command string
}

type AgentConfig struct {
	Plugins                []string
	Marketplaces           map[string]Marketplace
	AllowedBashCommands    []string
	AllowedToolPermissions []string
	SandboxAllowedDomains  []string
	SandboxAllowWrite      []string
	Hooks                  []Hook
	BashCommandPrefix      string
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
