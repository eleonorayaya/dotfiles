# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shizuku is a Go library for managing dotfiles. It generates files from templates, downloads remote resources, and syncs everything to the appropriate destinations. Consumers compose their own personal binary by instantiating a `shizuku.Builder`, registering apps from the shared library (or their own), and supplying their own user data.

Apps are organized into three categories — `languages/` (toolchains), `programs/` (regular programs), and `agents/` (agentic coding tools) — and synced in that order so later categories can consume outputs from earlier ones.

The canonical consumer lives at `examples/eleonora/` and compiles to the `shizuku` binary. Other users can `go get github.com/eleonorayaya/shizuku` and build their own entry point.

## Development Commands

Always use the `/task` skill (`task <target>`) to build, test, lint and run: never `go build`, `go test`, `go fmt`, `golangci-lint`, `go run` or `./out/shizuku`. Run the CLI through `/task run <args>`, e.g. `/task run diff -p` to preview and `/task run sync` to apply.

## Code Architecture

Side effects like remote resource fetching or exec commands belong in an app's `Sync()` only, not in `Generate()`. Files in an app's embedded `contents/` ending in `.tmpl` are expanded with Go `text/template` (no HTML escaping) and lose the extension; other files are copied as-is. Use the `/app` skill to add or change an app.

### Declaring Agent Requirements

Any app — language, program, or otherwise — may declare what agentic coding tools need to support it by implementing `AgentConfigProvider`:

```go
func (a *App) AgentConfig() app.AgentConfig {
    return app.AgentConfig{
        Plugins:               []string{"rust-analyzer-lsp@claude-plugins-official"},
        SandboxAllowedDomains: []string{"crates.io", "docs.rs"},
        SandboxAllowWrite:     []string{"~/.cargo", "~/.rustup"},
    }
}
```

`ShellInitScripts` lists shell init lines the agent's shells need (utena passes `eval "$(utena shell-init)"`); agents implementing `app.ContextualEnvProvider` (`Env(agents AgentContext)`) receive them when the env files are generated.

The builder collects every enabled app's `AgentConfig()` into a `SyncContext` before agents run. Agents that need this data implement `ContextualSyncer` / `ContextualGenerator` and receive the context as a parameter — see `agents/claude/claude.go` for an example. This keeps language- and tool-specific data out of agent code.

## Installing Packages

Use `pkg.Install(pkg.Spec{...})` in an app's `Install()`. A spec lists every source; the first that applies wins (brew on macOS; apt, then GitHub release on Linux):

```go
pkg.Spec{
    Brew: "lsd", Apt: "lsd", Bin: "lsd",
    Release: &pkg.Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
}
```

`shizuku upgrade` re-runs every app's `Install()` with the installer in upgrade mode (`pkg.SetUpgrade`). In this mode installed brew packages are upgraded, apt packages are upgraded with `--only-upgrade`, and binaries installed from GitHub releases are reinstalled from the latest release. Packages that aren't installed are skipped. Steps in `Install()` that don't go through `pkg` (e.g. `go install`, git clones) run exactly as they do for `install`, unless the app checks `pkg.Upgrading()`: the steel helix build skips when `~/.cargo/bin/hx` exists and only rebuilds on `upgrade`. `pkg.Run` is the shared command runner (`util.Run` wraps it), and `util.CloneOrPull` shallow-clones or pulls a source checkout.

`util.InstallBrewPackage` remains for macOS-only apps. Aliases that depend on an optional binary should set `Requires` so the shell stays usable if the install failed.

## Themes

`styles.ThemeColors` has a 16-color terminal `Palette` and role fields (`Keyword`, `String`, `Path`, `Added`, …). Terminal palettes (kitty, tuios) read `Palette`; everything else reads roles. `styles.WithTheme` fills empty roles from the palette. Every theme must be listed in `styles/themes/themes.go` (`themes.All`), which the tests check.

Import iTerm2 schemes with `task theme:import -- [-name X] [-primary magenta] <file-or-url>...` (flags before files). It writes `styles/themes/<name>.go`, registers it in `themes.All`, and prints low-contrast warnings.

## Profiles

The base profile is server-safe. `desktop` holds macOS apps, languages and agents; `work` and `personal` use `shizuku.Extends("desktop")`. Only add apps to the base if they work on Debian/Ubuntu.

`coder` is base plus remote-workspace variants: `tuios.Options{Remote: true}` (no laptop after-attach hook, no install), `utena.Options{Follower: true}` (no clone/build in `Install`) and `data.CoderClaudeOptions()` (`Remote: true` skips the sandbox, the `disableClaudeAiConnectors` deletion and `~/.claude.json`). The root `install.sh` bootstraps it on a Linux workspace.

The tuios config sets `preferred_shell = '/bin/zsh'` on every profile (`/bin/zsh` exists on macOS and on usrmerged Ubuntu; tuios falls back to `$SHELL` without it), so panes are zsh even when the tuios daemon started before `chsh`.

`tuios.Options.Picker` replaces the `prefix+p` popup (`UTENA_POPUP=1 utena`) with a `type = "shell"` binding running that command, so it runs on the laptop even in a remote session. `desktop` sets it to `data.UtenaPicker()` (`utena picker` opening a kitty window titled `utena-picker`, which the aerospace config floats); the base profile keeps the popup and `Remote` binds no `prefix+p`.

`kitty.Options.OpenURLWith` sets kitty's `open_url_with` (default `default`, the system opener). `desktop` sets it to `data.UtenaOpen()` (`utena-open --attached`), so ctrl+clicked links open in the attached utena session's browser window.

## Managing Shared Claude Code Settings

The `claude` shizuku app manages `~/.claude/settings.json` with additive merges for `enabledPlugins` and `permissions.allow`: removing an entry from the data leaves it in the live file, so a dropped plugin also needs `claude plugin uninstall <name>@<marketplace>` and a dropped allow rule a hand edit. To add allowed commands, plugins, or other managed fields, use the `/claude-config` skill.

Claude Code mods are plugins, so `claude.Options.Mods` takes `app.Mod` entries. Each one registers its marketplace (which defaults to the mod name), enables `<name>@<marketplace>`, and merges `Options` into `pluginConfigs.<id>.options`. Option keys the mod doesn't set keep whatever the user picked in `/config`. `ThemeOptions` derives options from the shizuku theme, and explicit `Options` override them. `claude.Prismantis()` uses it to map `ThemeColors` onto prismantis's color slots.

`utena.Options.Bundles` (`app.Bundle{Plugins, McpServers}`) defines utena plugin bundles, and `utena.Options.McpServers` holds MCP server definitions (with `${VAR}` placeholders for secrets). The utena app writes both into `~/.config/utena/daemon.json` and passes them to agents through its `AgentConfig()`. The claude app writes bundled plugins that aren't always-on or mods as `false` to `settings.json`; utena turns them on per session. On sync it removes the MCP servers from user scope in `~/.claude.json`, because utena writes the selected ones into each session's `.mcp.json`. The generated `shizuku.zshenv` (via `EnvSetup.ZshenvScripts`) sources `~/.config/shizuku/secrets.env` (untracked, `set -a`) if it exists, so the placeholders resolve in every zsh, including the non-interactive login shells utena launches. `DisableClaudeAiConnectors` sets `disableClaudeAiConnectors` per profile (the personal profile uses `data.PersonalClaudeOptions()`).

The claude app also generates a Claude Code custom theme at `~/.claude/themes/shizuku.json` (base `light`/`dark` from `Theme.Type`, token overrides from `ThemeColors`; see `agents/claude/theme.go`) and selects it with `"theme": "custom:shizuku"` in settings.json.

## Coding Style

### Comments
Do not add comments unless explicitly asked.

### Use Existing Utilities
Check the `util` package for an existing helper before writing file operations or other common tasks.

### Errors
Wrap every returned error with context: `fmt.Errorf("descriptive context: %w", err)`.
