package claude

import (
	"fmt"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

var desiredMarketplaces = map[string]string{
	"claude-plugins-official": "anthropics/claude-plugins-official",
	"superpowers-marketplace": "obra/superpowers-marketplace",
	"subtask":                 "zippoxer/subtask",
	"charm-dev-skills":        "williavs/charm-dev-skill-marketplace",
}

var alwaysOnPlugins = []string{
	"superpowers@superpowers-marketplace",
	"subtask@subtask",
}

var languagePlugins = map[shizukuconfig.Language][]string{
	shizukuconfig.LanguageGo:         {"gopls-lsp@claude-plugins-official", "charm-dev@charm-dev-skills"},
	shizukuconfig.LanguageLua:        {"lua-lsp@claude-plugins-official"},
	shizukuconfig.LanguageRust:       {"rust-analyzer-lsp@claude-plugins-official"},
	shizukuconfig.LanguageTypescript: {"typescript-lsp@claude-plugins-official"},
}

var desiredEnv = map[string]string{
	"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS": "1",
}

var desiredStatusLine = map[string]any{
	"type":    "command",
	"command": "~/.local/bin/starship-claude",
}

var desiredAllowedCommands = []string{
	"Bash(grep:*)",
	"Bash(find:*)",
	"Bash(ls:*)",
	"Bash(tree:*)",
	"Bash(cat:*)",
	"Bash(wc:*)",
	"Bash(xargs:*)",
	"Bash(bash:*)",
	"Bash(task:*)",
	"Bash(git add:*)",
	"Bash(git commit:*)",
	"Bash(git --version:*)",
	"Bash(brew --prefix:*)",
	"Skill(task)",
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "claude"
}

func (a *App) Enabled(config *shizukuconfig.Config) bool {
	return config.GetAppConfigBool(a.Name(), "enabled", true)
}

func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
	fileMap, err := shizukuapp.GenerateAppFiles("claude", nil, outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate app files: %w", err)
	}

	mergedPath, err := mergeSettings(outDir, config)
	if err != nil {
		return nil, fmt.Errorf("failed to merge settings: %w", err)
	}
	fileMap["settings.json"] = mergedPath

	marketplacesPath, err := mergeMarketplaces(outDir)
	if err != nil {
		return nil, fmt.Errorf("failed to merge marketplaces: %w", err)
	}
	fileMap["plugins/known_marketplaces.json"] = marketplacesPath

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.claude/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func getPlugins(config *shizukuconfig.Config) []string {
	plugins := make([]string, len(alwaysOnPlugins))
	copy(plugins, alwaysOnPlugins)
	for lang, langPlugins := range languagePlugins {
		if config.Languages[string(lang)].Enabled {
			plugins = append(plugins, langPlugins...)
		}
	}
	return plugins
}

func mergeMarketplaces(outDir string) (string, error) {
	marketplaces, err := util.ReadJSONMap("~/.claude/plugins/known_marketplaces.json")
	if err != nil {
		return "", fmt.Errorf("failed to read known_marketplaces.json: %w", err)
	}

	installBase, err := util.NormalizeFilePath("~/.claude/plugins/marketplaces")
	if err != nil {
		return "", fmt.Errorf("failed to normalize install base path: %w", err)
	}

	for name, repo := range desiredMarketplaces {
		if _, exists := marketplaces[name]; exists {
			continue
		}
		marketplaces[name] = map[string]any{
			"source": map[string]any{
				"source": "github",
				"repo":   repo,
			},
			"installLocation": filepath.Join(installBase, name),
		}
	}

	outPath := filepath.Join(outDir, "claude", "plugins", "known_marketplaces.json")
	if err := util.WriteJSONMap(outPath, marketplaces); err != nil {
		return "", fmt.Errorf("failed to write merged marketplaces: %w", err)
	}

	return outPath, nil
}

func mergeSettings(outDir string, config *shizukuconfig.Config) (string, error) {
	settings, err := util.ReadJSONMap("~/.claude/settings.json")
	if err != nil {
		return "", fmt.Errorf("failed to read settings.json: %w", err)
	}

	plugins, _ := settings["enabledPlugins"].(map[string]any)
	if plugins == nil {
		plugins = map[string]any{}
	}

	for _, plugin := range getPlugins(config) {
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
	for _, cmd := range desiredAllowedCommands {
		if !existing[cmd] {
			allowRaw = append(allowRaw, cmd)
		}
	}
	permissions["allow"] = allowRaw
	settings["permissions"] = permissions

	if len(desiredEnv) > 0 {
		env, _ := settings["env"].(map[string]any)
		if env == nil {
			env = map[string]any{}
		}
		for k, v := range desiredEnv {
			env[k] = v
		}
		settings["env"] = env
	}

	settings["statusLine"] = desiredStatusLine

	outPath := filepath.Join(outDir, "claude", "settings.json")
	if err := util.WriteJSONMap(outPath, settings); err != nil {
		return "", fmt.Errorf("failed to write merged settings: %w", err)
	}

	return outPath, nil
}
