# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shizuku is a Go library for managing dotfiles. It generates files from templates, downloads remote resources, and syncs everything to the appropriate destinations. Consumers compose their own personal binary by instantiating a `shizuku.Builder`, registering apps from the shared library (or their own), and supplying their own user data.

Apps are organized into three categories — `apps/languages/` (toolchains), `apps/programs/` (regular programs), and `apps/agents/` (agentic coding tools) — and synced in that order so later categories can consume outputs from earlier ones.

The canonical consumer lives at `examples/eleonora/` and compiles to the `shizuku` binary. Other users can `go get github.com/eleonorayaya/shizuku` and build their own entry point.

## Development Commands

**CRITICAL: ALWAYS use the `/task` skill for all build, test, lint, and run operations.**

### Task Skill Usage (MANDATORY)

The `/task` skill MUST be used for:
- **Building**: NEVER use `go build` or `make` directly
- **Testing**: NEVER use `go test` directly
- **Linting**: NEVER use `go fmt`, `golangci-lint`, or other linters directly
- **Running**: NEVER use `go run` or `./out/shizuku` directly

**Examples:**
- Build: `/task build` (NOT `go build`)
- Test: `/task test` (NOT `go test ./...`)
- Lint: `/task lint` (NOT `go fmt ./...`)
- Run: `/task run` (NOT `go run main.go` or `./out/shizuku`)

This ensures consistent build processes and proper dependency management.

### Running the CLI

**NEVER execute the built binary directly.** Always use `/task run` with arguments:

```bash
# Initialize default config at ~/.config/shizuku/shizuku.yml
/task run init

# Sync all application configurations
/task run sync

# Preview what would change on next sync
/task run diff

# Preview with full diff output
/task run diff -p

# Enable verbose logging
/task run sync --verbose
```

**FORBIDDEN**: Do NOT use `./out/shizuku`, `go run`, or any direct binary execution.

## Code Architecture

### Top-Level Structure

1. **shizuku.go** (root, `package shizuku`) - The `Builder` entry point. Consumers call `shizuku.New(Options{...}).AddLanguages(...).AddPrograms(...).AddAgent(...)` and then `Init()` / `Sync(ctx)` / `Diff(ctx)` / `Install(ctx)` / `List()`. The builder owns phase ordering (languages → programs → agents) and `SyncContext` assembly.

2. **app/** - Core types
   - `app.go`: `App`, `FileGenerator`, `FileSyncer` interfaces
   - `context.go`: `AgentConfig`, `AgentConfigProvider`, `SyncContext`, `ContextualSyncer`, `ContextualGenerator`
   - `files.go`: `GenerateAppFiles`, `SyncAppFiles`, diff helpers, remote resource fetching
   - `env.go`: Environment file generation

3. **config/** - `Config` struct, loading, validation, defaults, and the `Language` enum.

4. **util/** - File operations (copy, path normalization, directory creation, templates, homebrew helpers).

5. **apps/** - Pre-built library of apps
   - `apps/languages/` — language toolchains (e.g. `golang`, `rust`, `typescript`)
   - `apps/programs/` — regular programs (e.g. `nvim`, `kitty`, `git`)
   - `apps/agents/` — agentic coding tools (e.g. `claude`)

   Each app implements `Generate(outDir, cfg) (*app.GenerateResult, error)` and `Sync(outDir, cfg) error` (or the contextual variants — see below). Apps embed their `contents/` directory with `//go:embed all:contents` so the library is consumable without vendoring template files.

6. **examples/eleonora/** - The canonical consumer binary.
   - `main.go`: Cobra CLI (`init`, `sync`, `diff`, `install`, `list`, `upgrade`) that wires up the builder.
   - `data/`: Personal user data (marketplaces, allowed commands, sandbox paths, etc.).

### Data Flow

```
Consumer binary calls Builder.Sync(ctx)
  ↓
Load config from ~/.config/shizuku/shizuku.yml
  ↓
Create build directory: out/{timestamp}/
  ↓
Phase 1: Sync enabled languages (apps/languages/{name}/)
Phase 2: Sync enabled programs (apps/programs/{name}/)
Phase 3: Build SyncContext from all enabled languages + programs that
         implement AgentConfigProvider
Phase 4: Sync enabled agents (apps/agents/{name}/) — agents implementing
         ContextualSyncer receive the SyncContext
  ↓
For each app in a phase:
  ├─ Generate files from the app's embedded contents/ FS
  │  ├─ .tmpl files → Go template expansion
  │  └─ Regular files → Direct copy
  ├─ Download remote resources (if needed)
  └─ Sync to destination (e.g., ~/.config/{appName}/)
```

### App Implementation Pattern

Each app implements `FileGenerator` (for generation/diffing) and `FileSyncer` (for syncing). `Sync()` calls `Generate()` then syncs:

```go
//go:embed all:contents
var contents embed.FS

func (a *App) Generate(outDir string, cfg *config.Config) (*app.GenerateResult, error) {
    data := map[string]any{
        "key": "value",
    }

    fileMap, err := app.GenerateAppFiles("appName", contents, data, outDir)
    if err != nil {
        return nil, fmt.Errorf("failed to generate app files: %w", err)
    }

    return &app.GenerateResult{
        FileMap: fileMap,
        DestDir: "~/.config/appName/",
    }, nil
}

func (a *App) Sync(outDir string, cfg *config.Config) error {
    result, err := a.Generate(outDir, cfg)
    if err != nil {
        return err
    }

    if err := app.SyncAppFiles(result.FileMap, result.DestDir); err != nil {
        return fmt.Errorf("failed to sync app files: %w", err)
    }

    return nil
}
```

Side effects like remote resource fetching or exec commands belong in `Sync()` only, not in `Generate()`.

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

The builder collects every enabled app's `AgentConfig()` into a `SyncContext` before agents run. Agents that need this data implement `ContextualSyncer` / `ContextualGenerator` and receive the context as a parameter — see `apps/agents/claude/claude.go` for an example. This keeps language- and tool-specific data out of agent code.

### Configuration System

**Config file location:** `~/.config/shizuku/shizuku.yml`

**Structure:**
```yaml
languages:
  rust:
    enabled: true
    # Additional language-specific config inline
```

**Implementation details:**
- `Config` struct in `config/config.go`
- Language validation ensures only registered languages (defined in `config/language.go`) are allowed
- `CreateDefaultConfig()` generates a starter config with all languages disabled
- Config is loaded once and passed to each app's `Sync()` function

### Template Engine

- Uses Go's `html/template` package
- Files ending in `.tmpl` are processed with template data
- The `.tmpl` extension is stripped from output filenames
- Non-template files are copied directly
- Template data is passed as `map[string]any`

### Build Directory Structure

Generated files are placed in timestamped directories before syncing:
```
out/{unix_timestamp}/
├── appName1/
│   ├── generated_file1
│   └── generated_file2
└── appName2/
    └── ...
```

## Adding a New App

1. Pick a category for the new app:
   - `apps/languages/` — language toolchains
   - `apps/programs/` — regular programs
   - `apps/agents/` — agentic coding tools
2. Create directory: `apps/{category}/{appName}/`
3. Create `{appName}.go` implementing `Generate()` and `Sync()` (see App Implementation Pattern above). Embed the contents directory with `//go:embed all:contents` and pass the embedded FS to `app.GenerateAppFiles`.
4. Add source files to `contents/` directory (if needed).
5. Optionally implement `AgentConfig()` to declare LSP plugins, sandbox hosts, or sandbox write paths that agents should pick up.
6. Register the app in the consumer binary (`examples/eleonora/main.go`) via the builder:
   ```go
   shizuku.New(shizuku.Options{...}).
       AddPrograms(
           // ... existing programs
           appName.New(),
       )
   ```

## Adding a New Language

1. Add language constant to `config/language.go`:
   ```go
   const (
       LanguageRust Language = "rust"
       LanguageGo   Language = "go"  // New language
   )
   ```

2. Add to the `languages` slice:
   ```go
   var languages []Language = []Language{
       LanguageRust,
       LanguageGo,  // New language
   }
   ```

The validation and default config creation will automatically include the new language.

## Managing Shared Claude Code Settings

The `claude` shizuku app manages `~/.claude/settings.json` with additive merges for `enabledPlugins` and `permissions.allow`. To add allowed commands, plugins, or other managed fields, use the `/claude-config` skill.

## Coding Style

### Comments
Do not add comments unless explicitly asked. The code should be self-explanatory through clear naming and structure. Avoid obvious comments that simply restate what the code does.

### Use Existing Utilities
Always check for and use existing utility functions instead of reimplementing functionality. Before writing file operations or other common tasks, explore the `util` package to see if a utility function already exists that handles the task.

### Error Handling Pattern

All errors should be wrapped with context using `fmt.Errorf`:
```go
if err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

This creates a clear error chain showing where failures occurred in the call stack.

