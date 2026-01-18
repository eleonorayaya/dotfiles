---
name: app
description: Create, modify, and manage shizuku app configurations. Use when the user wants to create a new app, add template files to an app, or modify existing app structure.
---

# Shizuku App Manager

This skill helps create and manage shizuku application configurations.

## What This Skill Does

When invoked, this skill helps you:
1. **Scaffold new apps** - Create complete directory structure and registration
2. **Add template files** - Create and manage .tmpl files in app contents
3. **Modify existing apps** - Update app configurations and template data

## App Structure Overview

Each shizuku app follows this pattern:

```
apps/{appName}/
├── {appName}.go           # Main sync function
└── contents/              # Template files and assets
    ├── file.conf          # Regular files (copied as-is)
    └── file.tmpl          # Go templates (processed)
```

## Creating a New App

### Step 1: Gather Requirements

Ask the user:
1. **App name** - What is the name of the application? (e.g., "tmux", "alacritty")
2. **Destination path** - Where should files be synced? (e.g., "~/.config/tmux/")
3. **Template data** - Does the app need any template variables? (y/n)
4. **Remote files** - Does the app need to download any remote resources like plugins? (y/n)

### Step 2: Create Directory Structure

```bash
mkdir -p apps/{appName}/contents
```

### Step 3: Create the Go File

Create `apps/{appName}/{appName}.go` with this template:

```go
package {appName}

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal"
	"github.com/eleonorayaya/shizuku/internal/shizukuconfig"
)

func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("{appName}", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	if err := internal.SyncAppFiles(fileMap, "{destinationPath}"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
```

**With remote files:**
```go
func Sync(outDir string, config *shizukuconfig.Config) error {
	data := map[string]any{}

	fileMap, err := internal.GenerateAppFiles("{appName}", data, outDir)
	if err != nil {
		return fmt.Errorf("failed to generate app files: %w", err)
	}

	remoteFiles := map[string]string{
		"plugins/file.wasm": "https://example.com/file.wasm",
	}
	pluginMap, err := internal.FetchRemoteAppFiles(outDir, "{appName}", remoteFiles)
	if err != nil {
		return fmt.Errorf("failed to fetch remote files: %w", err)
	}
	maps.Copy(fileMap, pluginMap)

	if err := internal.SyncAppFiles(fileMap, "{destinationPath}"); err != nil {
		return fmt.Errorf("failed to sync app files: %w", err)
	}

	return nil
}
```

**Note:** If using `maps.Copy`, add import: `"maps"`

### Step 4: Register the App

1. Open `cmd/sync/sync.go`
2. Add import: `"github.com/eleonorayaya/shizuku/apps/{appName}"`
3. Add to the apps slice in the sync function:
   ```go
   {"{appName}", {appName}.Sync, nil},
   ```

### Step 5: Build and Test

```bash
task build
./out/shizuku sync
```

## Adding Template Files

When the user wants to add configuration files to an app:

1. **Ask for file path** - Relative to contents/ directory (e.g., "config.conf")
2. **Ask if it's a template** - Does it need variable substitution? (y/n)
3. **Create the file** - If template, use `.tmpl` extension
4. **Update template data** - If template, add necessary variables to the `data` map in the Go file

### Template Example

If creating `contents/config.conf.tmpl`:
```
# {{ .AppName }} Configuration
theme = "{{ .Theme }}"
```

Update the Go file:
```go
data := map[string]any{
	"AppName": "MyApp",
	"Theme":   "monade",
}
```

## Modifying Existing Apps

When modifying an app:
1. **Read the current Go file** to understand existing structure
2. **List contents/** to see existing files
3. **Make requested changes** while preserving the app pattern

## Important Notes

- **No comments** - Don't add comments unless explicitly requested (per CLAUDE.md)
- **Error wrapping** - Always use `fmt.Errorf("context: %w", err)` pattern
- **Template extension** - `.tmpl` files are processed, extension is removed in output
- **File syncing** - `SyncAppFiles` handles directory creation and file copying

## Workflow

1. Use TodoWrite to track the scaffolding steps
2. Ask clarifying questions using AskUserQuestion
3. Create all necessary files
4. Update cmd/sync/sync.go registration
5. Suggest running `task build` and testing

## Examples

**Simple app (no templates):**
- App: alacritty
- Destination: ~/.config/alacritty/
- Contents: alacritty.yml (static file)

**App with templates:**
- App: tmux
- Destination: ~/.tmux/
- Contents: tmux.conf.tmpl (with {{ .Theme }} variable)

**App with remote files:**
- App: zellij
- Destination: ~/.config/zellij/
- Remote: plugins/zjstatus.wasm from GitHub releases
