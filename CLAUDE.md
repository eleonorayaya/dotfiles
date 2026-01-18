# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shizuku is a Go-based configuration management tool for dotfiles. It generates files from templates, downloads remote resources, and syncs everything to the appropriate destinations. The tool manages 8 different application configurations (sketchybar, nvim, aerospace, zellij, kitty, fastfetch, jankyborders, desktoppr).

## Development Commands

**CRITICAL: Always use the `/task` skill for all build, test, lint, and run operations. NEVER invoke `go build`, `go run`, `go fmt`, or `go test` directly.**

### Using the CLI
```bash
# Initialize default config at ~/.config/shizuku/shizuku.yml
./out/shizuku init

# Sync all application configurations
./out/shizuku sync

# Enable verbose logging
./out/shizuku sync --verbose
```

## Code Architecture

### Three-Layer Structure

1. **cmd/** - CLI commands (Cobra-based)
   - `main.go`: Root command with `--verbose` flag
   - `init/`: Creates default config file
   - `sync/`: Orchestrates all app syncing

2. **apps/** - Application-specific handlers
   - Each app exports `Sync(outDir string, config *shizukuconfig.Config) error`
   - Must be manually registered in `cmd/sync/sync.go`

3. **internal/** - Shared utilities
   - `generate.go`: Template processing and file generation
   - `apps.go`: File syncing and remote resource fetching
   - `util/`: File operations (copy, path normalization, directory creation)
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

Each app follows this standard structure:

```go
func Sync(outDir string, config *shizukuconfig.Config) error {
    // 1. Create template data
    data := map[string]any{
        "key": "value",
    }

    // 2. Generate files from contents/
    fileMap, err := shizukuapp.GenerateAppFiles("appName", data, outDir)
    if err != nil {
        return fmt.Errorf("failed to generate app files: %w", err)
    }

    // 3. (Optional) Download remote resources
    remoteFiles := map[string]string{
        "plugins/file.wasm": "https://example.com/file.wasm",
    }
    pluginMap, err := internal.FetchRemoteAppFiles(outDir, "appName", remoteFiles)
    if err != nil {
        return fmt.Errorf("failed to fetch remote files: %w", err)
    }
    maps.Copy(fileMap, pluginMap)

    // 4. Sync all files to destination
    if err := shizukuapp.SyncAppFiles(fileMap, "~/.config/appName/"); err != nil {
        return fmt.Errorf("failed to sync app files: %w", err)
    }

    return nil
}
```

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
2. Create `{appName}.go` with `Sync(outDir string, config *shizukuconfig.Config) error` function
3. Add source files to `contents/` directory (if needed)
4. Import and register the app in `cmd/sync/sync.go`:
   ```go
   apps := []struct {
       name string
       fn   func(string, *shizukuconfig.Config) error
   }{
       // ... existing apps
       {"{appName}", appName.Sync},
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

