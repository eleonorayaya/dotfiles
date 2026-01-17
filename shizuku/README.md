# Shizuku

A Go-based configuration management tool for dotfiles.

## Overview

Shizuku manages application configurations by generating files from templates, downloading remote resources, and syncing everything to the appropriate destinations.

## Usage

```bash
# Sync all application configurations
task build
shizuku sync
```

## Architecture

### Apps Structure

Each app follows this pattern:

```
shizuku/apps/{appName}/
├── {appName}.go          # Exports Sync(outDir string) error
└── contents/             # Source files (optional for some apps)
    └── ...               # Files and directories to sync
```

### App Implementation

```go
func Sync(outDir string) error {
    data := map[string]any{} // Template data

    fileMap, err := internal.GenerateAppFiles("appName", data, outDir)
    if err != nil {
        return fmt.Errorf("failed to generate app files: %w", err)
    }

    if err := internal.SyncAppFiles(fileMap, "~/.config/appName/"); err != nil {
        return fmt.Errorf("failed to sync app files: %w", err)
    }

    return nil
}
```

## Managed Apps

- **sketchybar** - Menu bar with templates
- **aerospace** - Window manager config
- **fastfetch** - System info tool config
- **kitty** - Terminal emulator config
- **jankyborders** - Window borders config
- **zellij** - Terminal multiplexer with remote plugin downloads
- **nvim** - Neovim editor configuration (114 files)
- **desktoppr** - Wallpaper setter

## Features

### Template Support

Files ending in `.tmpl` are processed as Go templates with the provided data map.

### Remote File Fetching

Apps can fetch remote files (e.g., plugins) using `internal.FetchRemoteAppFiles()`:

```go
remoteFiles := map[string]string{
    "plugins/plugin.wasm": "https://example.com/plugin.wasm",
}

pluginMap, err := internal.FetchRemoteAppFiles(outDir, "appName", remoteFiles)
// Merge with fileMap before syncing
```

### Build Directory

Generated files are placed in `out/{timestamp}/` before being synced to their final destinations.

## Adding a New App

1. Create app directory: `shizuku/apps/{appName}/`
2. Create `{appName}.go` with a `Sync(outDir string) error` function
3. Add source files to `contents/` directory (if needed)
4. Import and register in `cmd/sync/sync.go`

