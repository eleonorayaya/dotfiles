# Plan: Shizuku as a Consumable Library

## Context

Shizuku is currently a single-binary CLI tool tightly coupled to my personal dotfile preferences. The repo straddles two roles: it ships a generic file-generation/sync engine *and* it hardcodes my apps, my marketplaces, my sandbox host list, my allowed git commands, etc. Anyone else who wanted to use shizuku to manage their own dotfiles would have to fork the repo and edit code throughout `agents/claude/claude.go`, `apps.go`, and the various app packages.

This refactor turns shizuku into a **library** that other users can `go get` and consume. Consumers compose their own personal binary by instantiating a builder, registering apps from the shared library (or their own), and supplying their own user data (marketplaces, allowed commands, etc.). My personal binary becomes one example consumer that lives inside the repo at `examples/eleonora/` — not the entry point of the package itself.

The two halves shizuku will provide:

1. **A toolkit** at the repo root (`shizuku.go` + `app/`, `config/`, `util/` subpackages) for defining apps, generating files from templates, syncing them to destinations, declaring agent requirements, and orchestrating phase-ordered sync via a builder.
2. **A library of pre-built apps** (`apps/languages/`, `apps/programs/`, `apps/agents/`) that consumers can pick from à la carte.

---

## Final Repo Layout

```
shizuku/                                # module root — package shizuku
├── shizuku.go                          # Builder: New(Options).AddLanguages(...).Execute()
├── app/                                # package app
│   ├── app.go                          # App, FileGenerator, FileSyncer interfaces
│   ├── context.go                      # AgentConfig, AgentConfigProvider, SyncContext,
│   │                                   # ContextualSyncer, ContextualGenerator
│   ├── files.go                        # GenerateAppFiles, SyncAppFiles, diff helpers (was internal/shizukuapp/files.go)
│   └── env.go                          # Env file generation (was internal/shizukuapp/env.go)
├── config/                             # package config
│   └── config.go                       # Config loading + validation (was internal/shizukuconfig)
├── util/                               # package util
│   └── ...                             # File ops, templates, JSON helpers (was internal/util)
│
├── apps/                               # pre-built library of apps
│   ├── languages/
│   │   ├── golang/, lua/, python/, ruby/, rust/, typescript/, zig/
│   ├── programs/
│   │   ├── aerospace/, bat/, buildkite/, desktoppr/, fastfetch/, git/,
│   │   ├── glow/, jankyborders/, k9s/, kitty/, lsd/, nvim/, protonpass/,
│   │   ├── protonvpn/, sfsymbols/, sketchybar/, terminal/, terraform/,
│   │   ├── tmux/, utena/
│   └── agents/
│       └── claude/                     # claude as a consumer-configurable app
│
├── examples/
│   └── eleonora/                       # my personal binary — the canonical example consumer
│       ├── main.go                     # ~25 lines: build options, call Execute()
│       └── data/                       # my marketplaces, my allowed commands, etc.
│
├── go.mod
├── Taskfile.yml
└── CLAUDE.md
```

Notes on the layout:

- `internal/shizukuapp` → `app/`. `internal/shizukuconfig` → `config/`. `internal/util` → `util/`. The `internal/` directory disappears so consumers can import these packages. No `pkg/` wrapper — the toolkit lives directly at the module root, so the most common imports are short: `import "github.com/eleonorayaya/shizuku"`, `import "github.com/eleonorayaya/shizuku/app"`.
- The top-level `apps.go` registry (the current `GetLanguages/GetPrograms/GetAgents` functions) is **removed** — registration moves into the consumer's `main.go` via the builder.
- `cmd/` (current `init`, `sync`, `diff` commands) is **removed** — Cobra moves into the example consumer. Library users wire up their own CLI however they like (or skip a CLI entirely).
- App `contents/` directories are embedded via `//go:embed` so consumers don't need to vendor template files.

---

## Builder API

The library's primary entry point is a builder in `shizuku.go` at the module root. Consumers construct one builder, register their apps, and call `Execute(ctx, action)` where action is `"sync"`, `"diff"`, or `"init"`.

```go
package shizuku

type Options struct {
    ConfigPath string                  // defaults to ~/.config/shizuku/shizuku.yml
    OutDir     string                  // defaults to out/{timestamp}
    Verbose    bool
}

type Builder struct { /* ... */ }

func New(opts Options) *Builder

func (b *Builder) AddLanguage(app app.App) *Builder
func (b *Builder) AddLanguages(apps ...app.App) *Builder
func (b *Builder) AddProgram(app app.App) *Builder
func (b *Builder) AddPrograms(apps ...app.App) *Builder
func (b *Builder) AddAgent(app app.App) *Builder
func (b *Builder) AddAgents(apps ...app.App) *Builder

// Execute runs the chosen action across all registered apps
// in phase order: languages → programs → agents.
func (b *Builder) Execute(ctx context.Context, action Action) error

type Action string
const (
    ActionSync Action = "sync"
    ActionDiff Action = "diff"
    ActionInit Action = "init"
)
```

The builder owns phase ordering, `SyncContext` assembly from `AgentConfigProvider` apps, and per-app dispatch (`ContextualSyncer` vs. `FileSyncer`). It does **not** define apps itself.

### Example consumer (`examples/eleonora/main.go`)

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/eleonorayaya/shizuku"
    "github.com/eleonorayaya/shizuku/apps/agents/claude"
    "github.com/eleonorayaya/shizuku/apps/languages/golang"
    "github.com/eleonorayaya/shizuku/apps/languages/rust"
    "github.com/eleonorayaya/shizuku/apps/programs/git"
    "github.com/eleonorayaya/shizuku/apps/programs/nvim"

    mydata "github.com/eleonorayaya/shizuku/examples/eleonora/data"
)

func main() {
    action := shizuku.Action(os.Args[1])

    err := shizuku.New(shizuku.Options{Verbose: true}).
        AddLanguages(golang.New(), rust.New()).
        AddPrograms(git.New(), nvim.New()).
        AddAgent(claude.New(mydata.ClaudeOptions())).
        Execute(context.Background(), action)

    if err != nil {
        log.Fatal(err)
    }
}
```

That's the entire personal binary — a thin assembly of library pieces.

---

## Extracting User Data

The current `agents/claude/claude.go` hardcodes my personal lists: `desiredMarketplaces`, `alwaysOnPlugins`, `desiredEnv`, `desiredStatusLine`, `desiredSandboxAllowedHosts`, `desiredSandboxAllowWrite`, `desiredAllowedCommands`. None of those belong in a library — they're my preferences.

### `apps/agents/claude` becomes consumer-configurable

```go
package claude

type Marketplace struct {
    Repo string
    Path string  // optional sub-path
}

type Options struct {
    Marketplaces        map[string]Marketplace
    AlwaysOnPlugins     []string
    Env                 map[string]string
    StatusLine          map[string]any
    SandboxAllowedHosts []string  // baseline; merged with AgentConfig contributions
    SandboxAllowWrite   []string  // baseline; merged with AgentConfig contributions
    AllowedCommands     []string
    DefaultMode         string    // e.g. "plan"
}

type App struct{ opts Options }

func New(opts Options) *App { return &App{opts: opts} }
```

`mergeSettings` reads from `a.opts` instead of package-level `desired*` vars. The aggregation logic from `SyncContext` is unchanged — it still folds in per-language plugins and sandbox additions.

### `examples/eleonora/data/claude.go`

This file holds my personal lists, lifted verbatim from the current `claude.go` `desired*` maps:

```go
package data

import "github.com/eleonorayaya/shizuku/apps/agents/claude"

func ClaudeOptions() claude.Options {
    return claude.Options{
        Marketplaces: map[string]claude.Marketplace{
            "claude-plugins-official": {Repo: "anthropics/claude-plugins-official"},
            "superpowers-marketplace": {Repo: "obra/superpowers-marketplace"},
            // ... rest of the current desiredMarketplaces map
        },
        AlwaysOnPlugins: []string{"superpowers@superpowers-marketplace"},
        Env: map[string]string{
            "CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING": "1",
            "CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS":  "1",
        },
        StatusLine: map[string]any{
            "type": "command", "command": "npx -y ccstatusline@latest", "padding": 0,
        },
        SandboxAllowedHosts: []string{ /* current desiredSandboxAllowedHosts */ },
        SandboxAllowWrite:   []string{ /* current desiredSandboxAllowWrite */ },
        AllowedCommands:     []string{ /* current desiredAllowedCommands */ },
        DefaultMode:         "plan",
    }
}
```

A consumer who doesn't want my marketplace/plugin choices simply supplies their own `claude.Options` — no fork required.

### Other apps with hardcoded user data

Most apps (`apps/programs/git`, `apps/programs/buildkite`, etc.) will need a similar audit during the migration: any value currently embedded that's clearly personal (git author email, buildkite org slug) becomes a field on a per-app `Options` struct. Apps with no user-specific data take no constructor arg (e.g. `bat.New()`).

---

## Migration Phases

The repo is large enough that a one-shot rewrite is risky. Six phases, each independently buildable and testable:

### Phase 1 — Promote internal packages to the module root

`git mv internal/shizukuapp app`, `internal/shizukuconfig config`, `internal/util util`. Rename the package declarations (`package shizukuapp` → `package app`, etc.). Update all imports throughout the repo. No behavior change. Gate: `task build && task test` clean.

### Phase 2 — Embed app contents

Add `//go:embed contents/*` to each app package and update `app.GenerateAppFiles` to accept an `fs.FS` instead of a directory path. This is a prerequisite for being importable — currently `GenerateAppFiles` reads from disk relative to the repo root. Gate: `task run -- diff` produces identical output to pre-phase.

### Phase 3 — Restructure to `apps/{category}/` and add Builder skeleton

`git mv languages/* apps/languages/`, same for programs and agents. Delete root `apps.go` (the registry). Create `shizuku.go` at the module root with the Builder (initially the Builder duplicates what the current `cmd/sync` and `cmd/diff` orchestrators do — `package shizuku`). The current `cmd/main.go` continues to work, just rewritten as ~30 lines that call `shizuku.New(...).AddX(...).Execute(...)`. Gate: `task run -- sync` produces byte-identical output to pre-phase.

### Phase 4 — Extract user data from `apps/agents/claude`

Convert claude to take an `Options` struct. Move `desired*` maps to `examples/eleonora/data/claude.go`. Audit other apps for hardcoded personal data (git, buildkite are likely candidates) and apply the same treatment. Gate: synced `~/.claude/settings.json` is byte-identical to pre-phase.

### Phase 5 — Move CLI to example consumer

`git mv cmd examples/eleonora/`. Rewrite `examples/eleonora/main.go` to import library packages (no more module-internal imports). Update `Taskfile.yml`: `task build` now compiles `examples/eleonora/main.go` to `out/shizuku`. Gate: `task build && task run -- sync` still works for me.

### Phase 6 — Polish for consumability (optional, can defer)

- Public `doc.go` files with package-level docs and short usage examples.
- `examples/minimal/main.go` — a 10-line example showing the smallest viable consumer.
- README rewrite focused on "how to use shizuku as a library" (not "how to install my dotfiles").
- CI workflow that builds `examples/eleonora` and runs `go vet ./...`.

---

## Critical Files

| File | Change |
|------|--------|
| `shizuku.go` (root, `package shizuku`) | **New** — Builder, Options, Action, Execute orchestrator (phase order + SyncContext assembly) |
| `app/app.go` | Move from `internal/shizukuapp/app.go`; package becomes `app` |
| `app/context.go` | Move from `internal/shizukuapp/context.go` |
| `app/files.go` | Move; refactor `GenerateAppFiles` to accept `fs.FS` |
| `app/env.go` | Move from `internal/shizukuapp/env.go` |
| `config/config.go` | Move from `internal/shizukuconfig/`; package becomes `config` |
| `util/...` | Move from `internal/util/` |
| `apps/{category}/{name}/` | Move from `{category}/{name}/`; add `//go:embed contents/*` |
| `apps/agents/claude/claude.go` | Add `Options` param to `New`; remove `desired*` package vars; read from `a.opts` in `mergeSettings` |
| `apps/agents/claude/claude_test.go` | Update tests to construct `App` with test `Options` |
| `examples/eleonora/main.go` | **New** — consumer entry point (~25 lines) |
| `examples/eleonora/data/claude.go` | **New** — my personal claude data (lifted from current `desired*` maps) |
| `examples/eleonora/data/*.go` | **New** — any other personal data extracted from apps |
| `apps.go` (root) | **Deleted** — registry moves into consumer code |
| `cmd/` | **Deleted** (moved into `examples/eleonora/`) |
| `Taskfile.yml` | Update `build`/`run` targets to point at `examples/eleonora/main.go` |
| `CLAUDE.md` | Rewrite around "this is a library; here's the example consumer" |
| `.claude/skills/app/Skill.md` | Update: new apps go in `apps/{category}/` and are registered via the consumer's builder, not `apps.go` |

---

## Verification

After **each phase**:

```
task build              # compiles cleanly
task test               # all tests pass
task run -- diff        # diff output unchanged from prior phase
```

After **Phase 5** (CLI relocation):

```
task run -- sync        # full sync completes; ~/.claude/settings.json byte-identical to pre-refactor snapshot
```

**Library consumability check** (after Phase 6, or earlier as a smoke test):

In a scratch directory outside this repo:

```bash
go mod init scratch
go get github.com/eleonorayaya/shizuku@<branch>
# write a 15-line main.go that imports github.com/eleonorayaya/shizuku and one app
go build .
```

If that compiles, the library is properly consumable. If `internal/` references leak through, Go's import rules will reject it — which is why Phase 1 promoting `internal/` → `pkg/` is non-negotiable for shipping.

**Manual check** (after Phase 4): with `golang` and `rust` enabled, the synced `settings.json` still contains `gopls-lsp@claude-plugins-official`, `charm-dev@charm-dev-skills`, `rust-analyzer-lsp@claude-plugins-official`, and all rust sandbox hosts/paths. Diff against a pre-refactor snapshot should be empty.
