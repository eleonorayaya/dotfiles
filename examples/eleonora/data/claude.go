package data

import (
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/app"
)

func ClaudeOptions() claude.Options {
	home, _ := os.UserHomeDir()

	return claude.Options{
		Marketplaces: map[string]app.Marketplace{
			"claude-plugins-official":  {Repo: "anthropics/claude-plugins-official"},
			"eleonorayaya-claude-code": {Repo: "eleonorayaya/claude-plugins"},
		},
		AlwaysOnPlugins: []string{
			"superpowers@claude-plugins-official",
			"claude-md-management@claude-plugins-official",
			"plugin-dev@claude-plugins-official",
			"skill-creator@claude-plugins-official",
			"playground@claude-plugins-official",
			"better-claude@eleonorayaya-claude-code",
			"git-workflows@eleonorayaya-claude-code",
		},
		Mods: []app.Mod{
			claude.Prismantis(),
			claude.PillFooter(claude.PillFooterOptions{
				ReportCommand: "tuios agent-statusline claude-code --integration 1 --turn-end",
			}),
		},
		Env: map[string]string{
			"CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING": "1",
			"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS":  "1",
			"JEST_WATCHMAN":                         "false",
			"GIT_CONFIG_GLOBAL":                     filepath.Join(home, ".config", "git", "claude.gitconfig"),
		},
		GlobalConfig: map[string]any{
			"leftArrowOpensAgents":  false,
			"prStatusFooterEnabled": false,
		},
		RemoveStatusLine: true,
		SandboxAllowedDomains: []string{
			"api.anthropic.com",
			"code.claude.com",
			"formulae.brew.sh",
		},
		SandboxAllowWrite: []string{
			"/tmp",
			"/private/tmp",
			"/dev/ptmx",
			"/dev/ttys*",
			"~/.docker",
			"~/.colima",
			"~/Library/Caches/dotslash",
		},
		AllowedBashCommands: []string{
			"grep:*",
			"find:*",
			"ls:*",
			"tree:*",
			"cat:*",
			"wc:*",
			"xargs:*",
			"echo:*",
			"head:*",
			"tail:*",
			"brew --prefix:*",
			"npx nx:*",
		},
		AllowedToolPermissions: []string{
			"Read(//tmp/**)",
			"Edit(//tmp/**)",
		},
		DeniedBashCommands: []string{
			"aws",
			"aws:*",
		},
		DefaultMode:  "plan",
		AdvisorModel: "opus",
		Settings: map[string]any{
			"model":                    "opus",
			"editorMode":               "vim",
			"tui":                      "fullscreen",
			"showThinkingSummaries":    true,
			"skipAutoPermissionPrompt": true,
			"skipWorkflowUsageWarning": true,
			"agentPushNotifEnabled":    false,
		},
	}
}
