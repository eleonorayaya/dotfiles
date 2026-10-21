package data

import (
	_ "embed"
	"maps"
	"slices"
	"strings"

	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/app"
)

//go:embed claude-global.md
var claudeMD string

func PersonalClaudeOptions() claude.Options {
	opts := ClaudeOptions()
	opts.DisableClaudeAiConnectors = true
	return opts
}

func CoderClaudeOptions() claude.Options {
	opts := ClaudeOptions()
	opts.Remote = true
	opts.GlobalConfig = nil
	opts.SandboxAllowedDomains = nil
	opts.SandboxAllowWrite = nil
	opts.AllowedBashCommands = slices.DeleteFunc(slices.Clone(opts.AllowedBashCommands), func(c string) bool {
		return strings.HasPrefix(c, "brew ")
	})
	opts.Marketplaces = maps.Clone(opts.Marketplaces)
	maps.Copy(opts.Marketplaces, map[string]app.Marketplace{
		"claude-plugins-official":   {Repo: "anthropics/claude-plugins-official"},
		"go-skills":                 {Repo: "spf13/go-skills"},
		"charm-dev-skills":          {Repo: "williavs/charm-dev-skill-marketplace"},
		"notion-plugin-marketplace": {Repo: "makenotion/claude-code-notion-plugin"},
	})
	return opts
}

func ClaudeOptions() claude.Options {
	return claude.Options{
		ClaudeMD: claudeMD,
		Marketplaces: map[string]app.Marketplace{
			"superpowers-marketplace":  {Repo: "obra/superpowers-marketplace"},
			"eleonorayaya-claude-code": {Repo: "eleonorayaya/claude-plugins"},
			"utena":                    {Repo: "eleonorayaya/utena"},
		},
		AlwaysOnPlugins: []string{
			"superpowers@superpowers-marketplace",
			"better-claude@eleonorayaya-claude-code",
			"git-workflows@eleonorayaya-claude-code",
			"utena-claude@utena",
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
			"versions.backstage.io",
		},
		SandboxAllowWrite: []string{
			"/tmp",
			"/private/tmp",
			"/dev/ptmx",
			"/dev/ttys*",
			"~/.docker",
			"~/.colima",
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
	}
}
