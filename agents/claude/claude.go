package claude

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/config"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

type Options struct {
	Marketplaces        map[string]app.Marketplace
	AlwaysOnPlugins     []string
	Env                 map[string]string
	StatusLine          map[string]any
	SandboxAllowedHosts []string
	SandboxAllowWrite   []string
	AllowedCommands     []string
	DefaultMode         string
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

func (a *App) Enabled(cfg *config.Config) bool {
	return cfg.GetAppConfigBool(a.Name(), "enabled", true)
}

const utenaShellInit = `eval "$(utena shell-init)"`

func (a *App) Env() (*app.EnvSetup, error) {
	return &app.EnvSetup{
		InitScripts: []string{utenaShellInit},
	}, nil
}

func (a *App) GenerateWithContext(outDir string, cfg *config.Config, ctx app.SyncContext) (*app.GenerateResult, error) {
	fileMap, err := app.GenerateAppFiles("claude", contents, nil, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	mergedPath, err := a.mergeSettings(outDir, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}
	fileMap["settings.json"] = mergedPath

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.claude/",
	}, nil
}

func (a *App) SyncWithContext(outDir string, cfg *config.Config, ctx app.SyncContext) error {
	result, err := a.GenerateWithContext(outDir, cfg, ctx)
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

func (a *App) collectPlugins(ctx app.SyncContext) []string {
	sources := [][]string{a.opts.AlwaysOnPlugins}
	for _, ac := range ctx.AgentConfigs {
		sources = append(sources, ac.Plugins)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectAllowedCommands(ctx app.SyncContext) []string {
	sources := [][]string{a.opts.AllowedCommands}
	for _, ac := range ctx.AgentConfigs {
		sources = append(sources, ac.AllowedCommands)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectMarketplaces(ctx app.SyncContext) map[string]app.Marketplace {
	merged := make(map[string]app.Marketplace, len(a.opts.Marketplaces))
	for name, m := range a.opts.Marketplaces {
		merged[name] = m
	}
	for _, ac := range ctx.AgentConfigs {
		for name, m := range ac.Marketplaces {
			if _, exists := merged[name]; exists {
				continue
			}
			merged[name] = m
		}
	}
	return merged
}

func (a *App) collectSandboxHosts(ctx app.SyncContext) []string {
	sources := [][]string{a.opts.SandboxAllowedHosts}
	for _, ac := range ctx.AgentConfigs {
		sources = append(sources, ac.SandboxAllowedHosts)
	}
	return dedupeStrings(sources...)
}

func (a *App) collectSandboxWrite(ctx app.SyncContext) []string {
	sources := [][]string{a.opts.SandboxAllowWrite}
	for _, ac := range ctx.AgentConfigs {
		sources = append(sources, ac.SandboxAllowWrite)
	}
	return dedupeStrings(sources...)
}

func (a *App) mergeSettings(outDir string, ctx app.SyncContext) (string, error) {
	settings, err := util.ReadJSONMap("~/.claude/settings.json")
	if err != nil {
		return "", fmt.Errorf("failed to read settings.json: %w", err)
	}

	plugins, _ := settings["enabledPlugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}

	for _, plugin := range a.collectPlugins(ctx) {
		plugins[plugin] = true
	}
	settings["enabledPlugins"] = plugins

	permissions, _ := settings["permissions"].(map[string]any)
	if permissions == nil {
		permissions = map[string]any{}
	}

	allowRaw, _ := permissions["allow"].([]any)
	existing := map[string]bool{}
	for _, entry := range allowRaw {
		if s, ok := entry.(string); ok {
			existing[s] = true
		}
	}
	for _, cmd := range a.collectAllowedCommands(ctx) {
		if !existing[cmd] {
			allowRaw = append(allowRaw, cmd)
			existing[cmd] = true
		}
	}
	permissions["allow"] = allowRaw
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
	for name, src := range a.collectMarketplaces(ctx) {
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

	allowedHostsRaw, _ := network["allowedHosts"].([]any)
	existingHosts := map[string]bool{}
	for _, entry := range allowedHostsRaw {
		if s, ok := entry.(string); ok {
			existingHosts[s] = true
		}
	}
	for _, host := range a.collectSandboxHosts(ctx) {
		if !existingHosts[host] {
			allowedHostsRaw = append(allowedHostsRaw, host)
		}
	}
	network["allowedHosts"] = allowedHostsRaw
	sandbox["network"] = network

	filesystem, _ := sandbox["filesystem"].(map[string]any)
	if filesystem == nil {
		filesystem = map[string]any{}
	}
	allowWriteRaw, _ := filesystem["allowWrite"].([]any)
	existingPaths := map[string]bool{}
	for _, entry := range allowWriteRaw {
		if s, ok := entry.(string); ok {
			existingPaths[s] = true
		}
	}
	for _, path := range a.collectSandboxWrite(ctx) {
		if !existingPaths[path] {
			allowWriteRaw = append(allowWriteRaw, path)
		}
	}
	filesystem["allowWrite"] = allowWriteRaw
	sandbox["filesystem"] = filesystem
	settings["sandbox"] = sandbox

	outPath := filepath.Join(outDir, "claude", "settings.json")
	if err := util.WriteJSONMap(outPath, settings); err != nil {
		return "", fmt.Errorf("failed to write merged settings: %w", err)
	}

	return outPath, nil
}
