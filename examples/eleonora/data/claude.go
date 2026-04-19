package data

import "github.com/eleonorayaya/shizuku/apps/agents/claude"

func ClaudeOptions() claude.Options {
	return claude.Options{
		Marketplaces: map[string]claude.Marketplace{
			"claude-plugins-official":   {Repo: "anthropics/claude-plugins-official"},
			"superpowers-marketplace":   {Repo: "obra/superpowers-marketplace"},
			"charm-dev-skills":          {Repo: "williavs/charm-dev-skill-marketplace"},
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
		SandboxAllowedHosts: []string{
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
		},
		SandboxAllowWrite: []string{
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
		},
		DefaultMode: "plan",
	}
}
