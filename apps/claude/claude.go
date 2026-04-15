package claude

import (
	"fmt"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/internal/shizukuapp"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
	"github.com/eleonorayaya/shizuku/internal/util"
)

type marketplaceSource struct {
	repo string
	path string
}

var desiredMarketplaces = map[string]marketplaceSource{
	"claude-plugins-official":  {repo: "anthropics/claude-plugins-official"},
	"superpowers-marketplace":  {repo: "obra/superpowers-marketplace"},
	"charm-dev-skills":         {repo: "williavs/charm-dev-skill-marketplace"},
	"eleonorayaya-claude-code": {repo: "eleonorayaya/claude-plugins"},
	"utena":                    {repo: "eleonorayaya/utena"},
}

var alwaysOnPlugins = []string{
	"superpowers@superpowers-marketplace",
}

var languagePlugins = map[shizukuconfig.Language][]string{
	shizukuconfig.LanguageGo:         {"gopls-lsp@claude-plugins-official", "charm-dev@charm-dev-skills"},
	shizukuconfig.LanguageLua:        {"lua-lsp@claude-plugins-official"},
	shizukuconfig.LanguageRuby:       {"ruby-lsp@claude-plugins-official"},
	shizukuconfig.LanguageRust:       {"rust-analyzer-lsp@claude-plugins-official"},
	shizukuconfig.LanguageTypescript: {"typescript-lsp@claude-plugins-official"},
}

var desiredEnv = map[string]string{
	"CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING": "1",
	"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS":  "1",
}

var desiredStatusLine = map[string]any{
	"type":    "command",
	"command": "npx -y ccstatusline@latest",
	"padding": 0,
}

var desiredSandboxAllowedHosts = []string{
	"api.anthropic.com",
	"code.claude.com",
	"api.github.com",
	"docs.github.com",
	"github.com",
	"raw.githubusercontent.com",
	"formulae.brew.sh",
	"api.buildkite.com",
	"buildkite.com",
	"mise.jdx.dev",
	"mise-versions.jdx.dev",
	"hk.jdx.dev",
}

var languageSandboxAllowedHosts = map[shizukuconfig.Language][]string{
	shizukuconfig.LanguageTypescript: {"registry.npmjs.org"},
	shizukuconfig.LanguageRust:       {"index.crates.io", "static.crates.io", "static.rust-lang.org", "crates.io", "docs.rs"},
}

var desiredSandboxAllowWrite = []string{
	"/dev/ptmx",
	"/dev/ttys*",
	"~/.claude/plugins/cache",

	"~/.cache/mise",
	"~/.config/mise",
	"~/.local/share/mise",
	"~/.local/state/mise",
	"~/Library/Caches/mise",

	"~/.docker",
	"~/.colima",
	"~/.config/gh",
	"~/.cache/gh",
	"~/.local/share/gh",
	"~/.local/state/gh",
	"~/.cache/pre-commit",
	"~/.cache/nvim/",
	"~/.task",
	"~/Library/Caches/dotslash",
}

var languageSandboxAllowWrite = map[shizukuconfig.Language][]string{
	shizukuconfig.LanguageRuby: {
		"~/.bundle",
		"~/.gem",
		"~/.rbenv",
		"~/.cache/bundler",
		"~/.cache/rubygems",
		"~/Library/Caches/bundle",
	},
	shizukuconfig.LanguageGo: {
		"~/.cache/go-build",
		"~/.config/go",
		"~/.local/share/go",
		"~/go",
		"~/.cache/golangci-lint",
		"~/Library/Caches/go-build",
		"~/Library/Caches/golangci-lint",
	},
	shizukuconfig.LanguageRust: {
		"~/.cargo",
		"~/.rustup",
		"~/Library/Caches/cargo",
	},
	shizukuconfig.LanguageTypescript: {
		"~/.npm",
		"~/.node-gyp",
		"~/.cache/node-gyp",
		"~/.cache/npm",
		"~/.cache/node",
		"~/.cache/yarn",
		"~/.cache/node/corepack",
		"~/.config/npm",
		"~/.config/configstore",
		"~/.config/yarn",
		"~/.config/pnpm",
		"~/.pnpm-state",
		"~/.pnpm-store",
		"~/.yarn",
		"~/.yarnrc",
		"~/.yarnrc.yml",
		"~/.local/share/pnpm",
		"~/.local/state/pnpm",
		"~/Library/Caches/npm",
		"~/Library/Caches/Yarn",
		"~/Library/Caches/node/corepack",
		"~/Library/Caches/pnpm",
		"~/Library/Preferences/pnpm",
		"~/Library/pnpm",
		"~/Library/Caches/ms-playwright",
		"~/Library/Caches/Cypress",
		"~/.cache/puppeteer",
	},
}

var desiredAllowedCommands = []string{
	"Bash(grep:*)",
	"Bash(find:*)",
	"Bash(ls:*)",
	"Bash(tree:*)",
	"Bash(cat:*)",
	"Bash(wc:*)",
	"Bash(xargs:*)",
	"Bash(echo:*)",

	"Bash(brew --prefix:*)",

	"Bash(npm install)",
	"Bash(npx nx test:*)",
	"Bash(npx nx sync:*)",

	"Edit(//tmp/**)",
	"Write(//tmp/**)",
	"Bash(git add:*)",
	"Bash(git commit:*)",
	"Bash(git --version:*)",
	"Bash(git status:*)",
	"Bash(git diff:*)",
	"Bash(git log:*)",
	"Bash(git fetch:*)",
	"Bash(git push:*)",
	"Bash(git rebase:*)",
	"Bash(git stash:*)",
	"Bash(git grep:*)",

	"Bash(gh pr view:*)",
	"Bash(gh pr list:*)",
	"Bash(gh pr checks:*)",
	"Bash(gh run view:*)",
	"Bash(gh run list:*)",
	"Bash(gh run watch:*)",

	"Bash(go build:*)",
	"Bash(go vet:*)",
	"Bash(go mod tidy:*)",
	"Skill(task)",
	"Bash(task:*)",

	"mcp__ide__getDiagnostics",
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

const utenaShellInit = `eval "$(utena shell-init)"`

func (a *App) Env() (*shizukuapp.EnvSetup, error) {
	return &shizukuapp.EnvSetup{
		InitScripts: []string{utenaShellInit},
	}, nil
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

func getSandboxAllowWrite(config *shizukuconfig.Config) []string {
	paths := make([]string, len(desiredSandboxAllowWrite))
	copy(paths, desiredSandboxAllowWrite)
	for lang, langPaths := range languageSandboxAllowWrite {
		if config.Languages[string(lang)].Enabled {
			paths = append(paths, langPaths...)
		}
	}
	return paths
}

func getSandboxAllowedHosts(config *shizukuconfig.Config) []string {
	hosts := make([]string, len(desiredSandboxAllowedHosts))
	copy(hosts, desiredSandboxAllowedHosts)
	for lang, langHosts := range languageSandboxAllowedHosts {
		if config.Languages[string(lang)].Enabled {
			hosts = append(hosts, langHosts...)
		}
	}
	return hosts
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

	settings["defaultMode"] = "plan"
	settings["statusLine"] = desiredStatusLine

	knownMarketplaces, _ := settings["extraKnownMarketplaces"].(map[string]any)
	if knownMarketplaces == nil {
		knownMarketplaces = map[string]any{}
	}
	for name, src := range desiredMarketplaces {
		if _, exists := knownMarketplaces[name]; exists {
			continue
		}
		source := map[string]any{
			"source": "github",
			"repo":   src.repo,
		}
		if src.path != "" {
			source["path"] = src.path
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
	for _, host := range getSandboxAllowedHosts(config) {
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
	for _, path := range getSandboxAllowWrite(config) {
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
