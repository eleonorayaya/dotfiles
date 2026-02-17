# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shizuku is a Go-based configuration management tool for dotfiles. It generates files from templates, downloads remote resources, and syncs everything to the appropriate destinations. The tool manages 8 different application configurations (sketchybar, nvim, aerospace, zellij, kitty, fastfetch, jankyborders, desktoppr).

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

### Three-Layer Structure

1. **cmd/** - CLI commands (Cobra-based)
   - `main.go`: Root command with `--verbose` flag
   - `init/`: Creates default config file
   - `sync/`: Orchestrates all app syncing
   - `diff/`: Previews what would change on next sync

2. **apps/** - Application-specific handlers
   - Each app implements `Generate(outDir, config) (*GenerateResult, error)` and `Sync(outDir, config) error`
   - `Generate()` returns a `GenerateResult` containing the `FileMap` and `DestDir`
   - `Sync()` calls `Generate()` then `SyncAppFiles()` (plus any side effects like remote fetches or exec commands)
   - Must be manually registered in `apps/apps.go`

3. **internal/** - Shared utilities
   - `shizukuapp/files.go`: File generation, syncing, diffing, and remote resource fetching
   - `shizukuapp/app.go`: `App`, `FileGenerator`, `FileSyncer` interfaces
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
For each registered app:
  ├─ Generate files from apps/{appName}/contents/
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

    fileMap, err := shizukuapp.GenerateAppFiles("appName", data, outDir)
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

1. Create directory: `apps/{appName}/`
2. Create `{appName}.go` implementing `Generate()` and `Sync()` (see App Implementation Pattern above)
3. Add source files to `contents/` directory (if needed)
4. Import and register the app in `apps/apps.go`:
   ```go
   func GetApps() []shizukuapp.App {
       return []shizukuapp.App{
           // ... existing apps
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

