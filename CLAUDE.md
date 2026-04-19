# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shizuku is a Go-based configuration management tool for dotfiles. It generates files from templates, downloads remote resources, and syncs everything to the appropriate destinations. Apps are organized into three categories — `languages/` (toolchains), `programs/` (regular programs), and `agents/` (agentic coding tools) — and synced in that order so later categories can consume outputs from earlier ones.

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

1. **cmd/** - CLI commands (Cobra-based)
   - `main.go`: Root command with `--verbose` flag
   - `init/`: Creates default config file
   - `sync/`: Orchestrates all app syncing in phase order (languages → programs → agents)
   - `diff/`: Previews what would change on next sync

2. **languages/** - Language toolchains (e.g. `golang`, `rust`, `typescript`)
3. **programs/** - Regular programs (e.g. `nvim`, `kitty`, `git`)
4. **agents/** - Agentic coding tools (e.g. `claude`)

   Each app implements `Generate(outDir, config) (*GenerateResult, error)` and `Sync(outDir, config) error` (or the contextual variants — see below). Apps must be manually registered in the root `apps.go` under the matching `GetLanguages()` / `GetPrograms()` / `GetAgents()` function.

5. **apps.go** (root) - Category-aware registry: `GetLanguages()`, `GetPrograms()`, `GetAgents()`, `GetApps()`.

6. **internal/** - Shared utilities
   - `shizukuapp/files.go`: File generation, syncing, diffing, and remote resource fetching
   - `shizukuapp/app.go`: `App`, `FileGenerator`, `FileSyncer` interfaces
   - `shizukuapp/context.go`: `AgentConfig`, `AgentConfigProvider`, `SyncContext`, `ContextualSyncer`, `ContextualGenerator`
   - `shizukuapp/env.go`: Environment file generation
   - `util/`: File operations (copy, path normalization, directory creation, templates)
   - `shizukuconfig/`: Config loading, validation, and defaults

### Data Flow

```
CLI (sync command)
  ↓
Load config from ~/.config/shizuku/shizuku.yml
  ↓
Create build directory: out/{timestamp}/
  ↓
Phase 1: Sync enabled languages (languages/{name}/)
Phase 2: Sync enabled programs (programs/{name}/)
Phase 3: Build SyncContext from all enabled languages + programs that
         implement AgentConfigProvider
Phase 4: Sync enabled agents (agents/{name}/) — agents implementing
         ContextualSyncer receive the SyncContext
  ↓
For each app in a phase:
  ├─ Generate files from {category}/{appName}/contents/
  │  ├─ .tmpl files → Go template expansion
  │  └─ Regular files → Direct copy
  ├─ Download remote resources (if needed)
  └─ Sync to destination (e.g., ~/.config/{appName}/)
```

### App Implementation Pattern

Each app implements `FileGenerator` (for generation/diffing) and `FileSyncer` (for syncing). `Sync()` calls `Generate()` then syncs:

```go
func (a *App) Generate(outDir string, config *shizukuconfig.Config) (*shizukuapp.GenerateResult, error) {
    data := map[string]any{
        "key": "value",
    }

    fileMap, err := shizukuapp.GenerateAppFiles("programs/appName", data, outDir)
    if err != nil {
        return nil, fmt.Errorf("failed to generate app files: %w", err)
    }

    return &shizukuapp.GenerateResult{
        FileMap: fileMap,
        DestDir: "~/.config/appName/",
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
```

Side effects like remote resource fetching or exec commands belong in `Sync()` only, not in `Generate()`.

### Declaring Agent Requirements

Any app — language, program, or otherwise — may declare what agentic coding tools need to support it by implementing `AgentConfigProvider`:

```go
func (a *App) AgentConfig() shizukuapp.AgentConfig {
    return shizukuapp.AgentConfig{
        Plugins:             []string{"rust-analyzer-lsp@claude-plugins-official"},
        SandboxAllowedHosts: []string{"crates.io", "docs.rs"},
        SandboxAllowWrite:   []string{"~/.cargo", "~/.rustup"},
    }
}
```

The sync orchestrator collects every enabled app's `AgentConfig()` into a `SyncContext` before agents run. Agents that need this data implement `ContextualSyncer` / `ContextualGenerator` and receive the context as a parameter — see `agents/claude/claude.go` for an example. This keeps language- and tool-specific data out of agent code.

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
- `Config` struct in `internal/shizukuconfig/config.go`
- Language validation ensures only registered languages (defined in `language.go`) are allowed
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
   - `languages/` — language toolchains
   - `programs/` — regular programs
   - `agents/` — agentic coding tools
2. Create directory: `{category}/{appName}/`
3. Create `{appName}.go` implementing `Generate()` and `Sync()` (see App Implementation Pattern above). Pass `"{category}/{appName}"` as the first argument to `GenerateAppFiles`.
4. Add source files to `contents/` directory (if needed).
5. Optionally implement `AgentConfig()` to declare LSP plugins, sandbox hosts, or sandbox write paths that agents should pick up.
6. Import and register the app in the root `apps.go` under the matching category function:
   ```go
   func GetPrograms() []shizukuapp.App {
       return []shizukuapp.App{
           // ... existing programs
           appName.New(),
       }
   }
   ```

## Adding a New Language

1. Add language constant to `internal/shizukuconfig/language.go`:
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
Always check for and use existing utility functions instead of reimplementing functionality. Before writing file operations or other common tasks, explore the `internal/util` package to see if a utility function already exists that handles the task.

### Error Handling Pattern

All errors should be wrapped with context using `fmt.Errorf`:
```go
if err != nil {
    return fmt.Errorf("descriptive context: %w", err)
}
```

This creates a clear error chain showing where failures occurred in the call stack.

