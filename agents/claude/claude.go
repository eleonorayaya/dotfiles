package claude

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

type Options struct {
	Marketplaces          map[string]app.Marketplace
	AlwaysOnPlugins       []string
	Env                   map[string]string
	StatusLine            map[string]any
	SandboxAllowedDomains []string
	SandboxAllowWrite     []string
	AllowedCommands       []string
	DefaultMode           string
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "claude"
}

const utenaShellInit = `eval "$(utena shell-init)"`

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		InitScripts: []string{utenaShellInit},
	}, nil
}

func (a *App) Generate(ctx *app.Context, agents app.AgentContext) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("claude", contents, nil, ctx.OutDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	mergedPath, err := a.mergeSettings(ctx.OutDir, agents)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}
	fileMap["settings.json"] = mergedPath

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.claude/",
	}, nil
}

func (a *App) Sync(ctx *app.Context, agents app.AgentContext) error {
	result, err := a.Generate(ctx, agents)
	if err != nil {
		return err
	}

	if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func dedupeStrings(sources ...[]string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, src := range sources {
		for _, s := range src {
			if seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func mergeStringsIntoAnySlice(existing []any, additions []string) []any {
	seen := map[string]bool{}
	for _, entry := range existing {
		if s, ok := entry.(string); ok {
			seen[s] = true
		}
	}
	for _, s := range additions {
		if seen[s] {
			continue
		}
		seen[s] = true
		existing = append(existing, s)
	}
	return existing
}

var baselineAllowedCommands = []string{
	"mcp__ide__getDiagnostics",
}

var baselineSandboxAllowWrite = []string{
	"~/.claude/plugins/cache",
}

func (a *App) collectPlugins(agents app.AgentContext) []string {
	sources := [][]string{a.opts.AlwaysOnPlugins}
	for _, ac := range agents.AgentConfigs {
		sources = append(sources, ac.Plugins)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectAllowedCommands(agents app.AgentContext) []string {
	sources := [][]string{baselineAllowedCommands, a.opts.AllowedCommands}
	for _, ac := range agents.AgentConfigs {
		sources = append(sources, ac.AllowedCommands)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectMarketplaces(agents app.AgentContext) map[string]app.Marketplace {
	merged := make(map[string]app.Marketplace, len(a.opts.Marketplaces))
	for name, m := range a.opts.Marketplaces {
		merged[name] = m
	}
	for _, ac := range agents.AgentConfigs {
		for name, m := range ac.Marketplaces {
			if _, exists := merged[name]; exists {
				continue
			}
			merged[name] = m
		}
	}
	return merged
}

func (a *App) collectSandboxDomains(agents app.AgentContext) []string {
	sources := [][]string{a.opts.SandboxAllowedDomains}
	for _, ac := range agents.AgentConfigs {
		sources = append(sources, ac.SandboxAllowedDomains)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectSandboxWrite(agents app.AgentContext) []string {
	sources := [][]string{baselineSandboxAllowWrite, a.opts.SandboxAllowWrite}
	for _, ac := range agents.AgentConfigs {
		sources = append(sources, ac.SandboxAllowWrite)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectHooks(agents app.AgentContext) []app.Hook {
	hooks := []app.Hook{}
	for _, ac := range agents.AgentConfigs {
		hooks = append(hooks, ac.Hooks...)
	}
	return hooks
}

func mergeHooks(existing map[string]any, hooks []app.Hook) map[string]any {
	if existing == nil {
		existing = map[string]any{}
	}
	for _, h := range hooks {
		eventEntries, _ := existing[h.Event].([]any)
		matcherIdx := -1
		for i, entry := range eventEntries {
			entryMap, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if m, _ := entryMap["matcher"].(string); m == h.Matcher {
				matcherIdx = i
				break
			}
		}

		if matcherIdx == -1 {
			eventEntries = append(eventEntries, map[string]any{
				"matcher": h.Matcher,
				"hooks": []any{
					map[string]any{"type": "command", "command": h.Command},
				},
			})
			existing[h.Event] = eventEntries
			continue
		}

		entryMap, _ := eventEntries[matcherIdx].(map[string]any)
		commandHooks, _ := entryMap["hooks"].([]any)
		exists := false
		for _, ch := range commandHooks {
			chMap, ok := ch.(map[string]any)
			if !ok {
				continue
			}
			if cmd, _ := chMap["command"].(string); cmd == h.Command {
				exists = true
				break
			}
		}
		if !exists {
			commandHooks = append(commandHooks, map[string]any{"type": "command", "command": h.Command})
			entryMap["hooks"] = commandHooks
			eventEntries[matcherIdx] = entryMap
			existing[h.Event] = eventEntries
		}
	}
	return existing
}

func (a *App) mergeSettings(outDir string, agents app.AgentContext) (string, error) {
	settings, err := util.ReadJSONMap("~/.claude/settings.json")
	if err != nil {
		return "", fmt.Errorf("failed to read settings.json: %w", err)
	}

	plugins, _ := settings["enabledPlugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}

	for _, plugin := range a.collectPlugins(agents) {
		plugins[plugin] = true
	}
	settings["enabledPlugins"] = plugins

	permissions, _ := settings["permissions"].(map[string]any)
	if permissions == nil {
		permissions = map[string]any{}
	}

	allowRaw, _ := permissions["allow"].([]any)
	permissions["allow"] = mergeStringsIntoAnySlice(allowRaw, a.collectAllowedCommands(agents))
	settings["permissions"] = permissions

	if len(a.opts.Env) > 0 {
		env, _ := settings["env"].(map[string]any)
		if env == nil {
			env = map[string]any{}
		}
		for k, v := range a.opts.Env {
			env[k] = v
		}
		settings["env"] = env
	}

	if a.opts.DefaultMode != "" {
		settings["defaultMode"] = a.opts.DefaultMode
	}
	if a.opts.StatusLine != nil {
		settings["statusLine"] = a.opts.StatusLine
	}

	knownMarketplaces, _ := settings["extraKnownMarketplaces"].(map[string]any)
	if knownMarketplaces == nil {
		knownMarketplaces = map[string]any{}
	}
	for name, src := range a.collectMarketplaces(agents) {
		if _, exists := knownMarketplaces[name]; exists {
			continue
		}
		source := map[string]any{
			"source": "github",
			"repo":   src.Repo,
		}
		if src.Path != "" {
			source["path"] = src.Path
		}
		knownMarketplaces[name] = map[string]any{
			"source": source,
		}
	}
	settings["extraKnownMarketplaces"] = knownMarketplaces

	sandbox, _ := settings["sandbox"].(map[string]any)
	if sandbox == nil {
		sandbox = map[string]any{}
	}
	sandbox["enabled"] = true
	sandbox["autoAllowBashIfSandboxed"] = true
	sandbox["enableWeakerNetworkIsolation"] = true

	network, _ := sandbox["network"].(map[string]any)
	if network == nil {
		network = map[string]any{}
	}
	network["allowAllUnixSockets"] = true
	network["allowLocalBinding"] = true

	allowedDomainsRaw, _ := network["allowedDomains"].([]any)
	if legacy, ok := network["allowedHosts"].([]any); ok {
		allowedDomainsRaw = append(allowedDomainsRaw, legacy...)
		delete(network, "allowedHosts")
	}
	network["allowedDomains"] = mergeStringsIntoAnySlice(allowedDomainsRaw, a.collectSandboxDomains(agents))
	sandbox["network"] = network

	filesystem, _ := sandbox["filesystem"].(map[string]any)
	if filesystem == nil {
		filesystem = map[string]any{}
	}
	allowWriteRaw, _ := filesystem["allowWrite"].([]any)
	filesystem["allowWrite"] = mergeStringsIntoAnySlice(allowWriteRaw, a.collectSandboxWrite(agents))
	sandbox["filesystem"] = filesystem
	settings["sandbox"] = sandbox

	hooks, _ := settings["hooks"].(map[string]any)
	settings["hooks"] = mergeHooks(hooks, a.collectHooks(agents))

	outPath := filepath.Join(outDir, "claude", "settings.json")
	if err := util.WriteJSONMap(outPath, settings); err != nil {
		return "", fmt.Errorf("failed to write merged settings: %w", err)
	}

	return outPath, nil
}
