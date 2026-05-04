package data

import (
	"github.com/eleonorayaya/shizuku/agents/claude"
	"github.com/eleonorayaya/shizuku/app"
)

func ClaudeOptions() claude.Options {
	return claude.Options{
		Marketplaces: map[string]app.Marketplace{
			"superpowers-marketplace":   {Repo: "obra/superpowers-marketplace"},
			"claude-code-notion-plugin": {Repo: "makenotion/claude-code-notion-plugin"},
			"eleonorayaya-claude-code":  {Repo: "eleonorayaya/claude-plugins"},
			"utena":                     {Repo: "eleonorayaya/utena"},
		},
		AlwaysOnPlugins: []string{
			"superpowers@superpowers-marketplace",
		},
		Env: map[string]string{
			"CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING": "1",
			"CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS":  "1",
		},
		StatusLine: map[string]any{
			"type":    "command",
			"command": "npx -y ccstatusline@latest",
			"padding": 0,
		},
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
		},
		AllowedCommands: []string{
			"Bash(grep:*)",
			"Bash(find:*)",
			"Bash(ls:*)",
			"Bash(tree:*)",
			"Bash(cat:*)",
			"Bash(wc:*)",
			"Bash(xargs:*)",
			"Bash(echo:*)",
			"Bash(head:*)",
			"Bash(tail:*)",

			"Bash(brew --prefix:*)",

			"Bash(npx nx:*)",

			"Read(//tmp/**)",
			"Edit(//tmp/**)",
			"Write(//tmp/**)",
		},
		DefaultMode:  "plan",
		AdvisorModel: "opus",
	}
}
