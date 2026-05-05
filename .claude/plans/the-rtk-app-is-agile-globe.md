# RTK Allowed Command Decoration

## Context

RTK's `PreToolUse` hook intercepts Bash tool calls and prepends `rtk` (e.g. `git status` → `rtk git status`). Claude Code evaluates permissions *after* the hook rewrites the command, so `Bash(git:*)` no longer matches the rewritten form.

The fix is two-pronged:
1. **Refactor** `AllowedCommands` into `AllowedBashCommands` (bare patterns) and `AllowedToolPermissions` (non-bash entries). Claude owns the `Bash(...)` wrapping, making decoration trivial.
2. **Decorate**: when any `AgentConfig` has `BashCommandPrefix` set, emit both `Bash(X)` and `Bash(prefix X)` for every bash entry.

## Field Changes

### `app/context.go` — split + add prefix field

```go
type AgentConfig struct {
    Plugins                []string
    Marketplaces           map[string]app.Marketplace
    AllowedBashCommands    []string  // bare patterns: "git add:*", "grep:*"
    AllowedToolPermissions []string  // raw non-bash: "Read(//tmp/**)", "Skill(task)", "mcp__..."
    SandboxAllowedDomains  []string
    SandboxAllowWrite      []string
    Hooks                  []Hook
    BashCommandPrefix      string    // e.g. "rtk" — triggers duplicate decorated entries
}
```

### `agents/claude/claude.go` — same split on `Options`

```go
type Options struct {
    // ...
    AllowedBashCommands    []string
    AllowedToolPermissions []string
    // ...
}
```

`baselineAllowedCommands` (currently `["mcp__ide__getDiagnostics"]`) becomes `baselineAllowedToolPermissions`.

## `collectAllowedCommands` rewrite

```go
func (a *App) collectAllowedCommands(agents app.AgentContext) []string {
    // collect bash command patterns
    bashSources := [][]string{a.opts.AllowedBashCommands}
    for _, ac := range agents.AgentConfigs {
        bashSources = append(bashSources, ac.AllowedBashCommands)
    }
    bashCmds := dedupeStrings(bashSources...)

    // collect non-bash tool permissions
    toolSources := [][]string{baselineAllowedToolPermissions, a.opts.AllowedToolPermissions}
    for _, ac := range agents.AgentConfigs {
        toolSources = append(toolSources, ac.AllowedToolPermissions)
    }
    toolPerms := dedupeStrings(toolSources...)

    // collect bash prefixes (e.g. "rtk")
    prefixes := collectBashPrefixes(agents)

    // build final permissions: wrap each bash cmd, add prefixed copies if needed
    var permissions []string
    for _, cmd := range bashCmds {
        permissions = append(permissions, "Bash("+cmd+")")
        for _, p := range prefixes {
            permissions = append(permissions, "Bash("+p+" "+cmd+")")
        }
    }
    permissions = append(permissions, toolPerms...)
    return dedupeStrings(permissions)
}

func collectBashPrefixes(agents app.AgentContext) []string {
    seen := map[string]bool{}
    var out []string
    for _, ac := range agents.AgentConfigs {
        if ac.BashCommandPrefix != "" && !seen[ac.BashCommandPrefix] {
            seen[ac.BashCommandPrefix] = true
            out = append(out, ac.BashCommandPrefix)
        }
    }
    return out
}
```

## Callsite Updates

All `AllowedCommands` fields get split. Every `"Bash(X)"` entry strips its wrapper into `AllowedBashCommands`; non-bash entries go into `AllowedToolPermissions`.

| File | `AllowedBashCommands` | `AllowedToolPermissions` |
|---|---|---|
| `languages/golang/golang.go` | `"go build:*"`, `"go vet:*"`, `"go mod tidy:*"`, `"task:*"` | `"Skill(task)"` |
| `languages/typescript/typescript.go` | `"npm install"` | — |
| `languages/swift/swift.go` | `"swift build:*"` | — |
| `programs/git/git.go` | `"git add:*"`, `"git commit:*"`, … all 14 git/gh entries | — |
| `programs/rtk/rtk.go` | — | — (sets `BashCommandPrefix: "rtk"` only) |
| `examples/eleonora/data/claude.go` | `"grep:*"`, `"find:*"`, `"ls:*"`, `"tree:*"`, `"cat:*"`, `"wc:*"`, `"xargs:*"`, `"echo:*"`, `"head:*"`, `"tail:*"`, `"brew --prefix:*"`, `"npx nx:*"` | `"Read(//tmp/**)"`、`"Edit(//tmp/**)"`、`"Write(//tmp/**)"` |

## Verification

1. `/task build` — confirm compilation
2. `/task test` — confirm no regressions
3. `/task run diff` — inspect generated `settings.json`; verify `permissions.allow` contains both `"Bash(git add:*)"` and `"Bash(rtk git add:*)"`, and that `"Read(//tmp/**)"` is present without duplication
4. `/task run sync` — apply and confirm Claude Code stops prompting for allowlisted commands after RTK rewrites them
