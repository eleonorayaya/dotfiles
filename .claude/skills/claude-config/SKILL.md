---
name: claude-config
description: Use when the user wants to add or remove allowed commands, plugins, bundles, env vars, MCP servers, or other fields in their shared Claude Code settings managed by shizuku. Use when they say things like "add X to my allowed commands", "add plugin Y", "put Z in a bundle", or "set env var W".
---

# Claude Config Manager

The shizuku `claude` app (`agents/claude/claude.go`) generates `~/.claude/settings.json` and parts of `~/.claude.json`. Never edit those files by hand; change the source below, then preview with `/task run diff -p` and apply with `/task run sync`.

| Change | Where |
|---|---|
| Always-on plugin, marketplace, mod | `examples/eleonora/data/claude.go` (`AlwaysOnPlugins`, `Marketplaces`, `Mods`) |
| Per-session plugin or MCP server | `examples/eleonora/data/bundles.go` (`Bundles`, `McpServers`); utena turns bundles on per session |
| Plugins, allow rules, sandbox paths that belong to a language or program | that app's `AgentConfig()` (e.g. `languages/golang/golang.go`) |
| Allowed / denied Bash commands, tool permissions | `AllowedBashCommands` (`"cmd sub:*"`, wrapped as `Bash(...)`), `AllowedToolPermissions`, `DeniedBashCommands` in `data/claude.go` |
| Env vars, default permission mode, advisor model, global config | `Env`, `DefaultMode`, `AdvisorModel`, `GlobalConfig` in `data/claude.go` |
| Profile differences | `PersonalClaudeOptions` / `CoderClaudeOptions` in `data/claude.go` |

## Removing things

Merges into `settings.json` are additive: `enabledPlugins`, `permissions.allow`/`deny` and `extraKnownMarketplaces` entries the data no longer declares stay in the live file. After removing one from the data:

- Plugin: `claude plugin uninstall <name>@<marketplace>`
- Marketplace: `claude plugin marketplace remove <name>`
- Allow/deny rule: delete it from `~/.claude/settings.json` by hand

## Verify

Run `/task test`, then `/task run diff -p` and check only the intended keys change.
