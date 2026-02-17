package claude

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type marketplace struct {
	Name string
	Repo string
}

var desiredMarketplaces = []marketplace{
	{Name: "claude-plugins-official", Repo: "anthropics/claude-plugins-official"},
	{Name: "superpowers-marketplace", Repo: "obra/superpowers-marketplace"},
	{Name: "subtask", Repo: "zippoxer/subtask"},
	{Name: "charm-dev-skills", Repo: "williavs/charm-dev-skill-marketplace"},
}

var alwaysOnPlugins = []string{
	"superpowers@superpowers-marketplace",
}

var optionalPlugins = []string{
	"lua-lsp@claude-plugins-official",
	"typescript-lsp@claude-plugins-official",
	"gopls-lsp@claude-plugins-official",
	"rust-analyzer-lsp@claude-plugins-official",
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

	return &shizukuapp.GenerateResult{
		FileMap: fileMap,
		DestDir: "~/.claude/",
	}, nil
}

func (a *App) Sync(outDir string, config *shizukuconfig.Config) error {
	if err := ensureMarketplaces(); err != nil {
		slog.Warn("failed to ensure marketplaces", "error", err)
	}

	result, err := a.Generate(outDir, config)
	if err != nil {
		return err
	}

	if err := shizukuapp.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}

func ensureMarketplaces() error {
	if !util.BinaryExists("claude") {
		return fmt.Errorf("claude CLI not found on PATH")
	}

	installed, err := loadInstalledMarketplaces()
	if err != nil {
		installed = map[string]bool{}
	}

	for _, m := range desiredMarketplaces {
		if installed[m.Name] {
			slog.Debug("marketplace already installed", "name", m.Name)
			continue
		}

		slog.Info("installing marketplace", "name", m.Name, "repo", m.Repo)
		cmd := exec.Command("claude", "plugin", "marketplace", "add", m.Repo)
		cmd.Env = filterEnv(os.Environ(), "CLAUDECODE")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install marketplace %s: %w\nOutput: %s", m.Name, err, output)
		}
	}

	return nil
}

func loadInstalledMarketplaces() (map[string]bool, error) {
	marketplacesPath, err := util.NormalizeFilePath("~/.claude/plugins/known_marketplaces.json")
	if err != nil {
		return nil, fmt.Errorf("failed to normalize marketplaces path: %w", err)
	}

	return loadInstalledMarketplacesFromPath(marketplacesPath)
}

func loadInstalledMarketplacesFromPath(path string) (map[string]bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read known_marketplaces.json: %w", err)
	}

	var marketplaces map[string]json.RawMessage
	if err := json.Unmarshal(data, &marketplaces); err != nil {
		return nil, fmt.Errorf("failed to parse known_marketplaces.json: %w", err)
	}

	installed := make(map[string]bool, len(marketplaces))
	for name := range marketplaces {
		installed[name] = true
	}

	return installed, nil
}

func filterEnv(env []string, exclude string) []string {
	prefix := exclude + "="
	filtered := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func getPlugins(config *shizukuconfig.Config) []string {
	plugins := make([]string, len(alwaysOnPlugins))
	copy(plugins, alwaysOnPlugins)
	if config.GetAppConfigBool("claude", "lsp_plugins", false) {
		plugins = append(plugins, optionalPlugins...)
	}
	if config.GetAppConfigBool("claude", "charm_dev", false) {
		plugins = append(plugins, "charm-dev@charm-dev-skills")
	}
	return plugins
}

func mergeSettings(outDir string, config *shizukuconfig.Config) (string, error) {
	settingsPath, err := util.NormalizeFilePath("~/.claude/settings.json")
	if err != nil {
		return "", fmt.Errorf("failed to normalize settings path: %w", err)
	}

	settings := map[string]any{}

	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return "", fmt.Errorf("failed to parse settings.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
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

	merged, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal settings: %w", err)
	}
	merged = append(merged, '\n')

	outPath := filepath.Join(outDir, "claude", "settings.json")
	if err := os.WriteFile(outPath, merged, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to write merged settings: %w", err)
	}

	return outPath, nil
}
