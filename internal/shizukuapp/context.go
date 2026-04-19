package shizukuapp

import "github.com/eleonorayaya/shizuku/internal/shizukuconfig"

type AgentConfig struct {
	Plugins             []string
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
	SyncWithContext(outDir string, config *shizukuconfig.Config, ctx SyncContext) error
}

type ContextualGenerator interface {
	GenerateWithContext(outDir string, config *shizukuconfig.Config, ctx SyncContext) (*GenerateResult, error)
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
