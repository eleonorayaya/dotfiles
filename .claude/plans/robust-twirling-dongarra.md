# Plan: Add `shizuku diff` command

## Context

Currently `shizuku sync` generates config files from templates and copies them to their destinations (e.g. `~/.config/nvim/`). There's no way to preview what would change before syncing. The new `shizuku diff` command will generate files into the build directory, diff each one against the existing destination file, and write `.diff` files alongside the generated files in the build dir.

## Design

The core problem is that each app's destination directory (e.g. `~/.config/sketchybar/`) is hardcoded inside its `Sync()` method, bundled with the generation logic. We need to extract generation from syncing.

**Approach:** Add a `FileGenerator` interface. Each app implements `Generate()` which returns the fileMap + destination dir. `Sync()` is refactored to call `Generate()` then `SyncAppFiles()`. The diff command calls `Generate()` then diffs against the destination.

## Steps

### 1. Add types and diff function to `internal/shizukuapp/files.go`

Add after the `FileSyncer` interface (line 17):

- `GenerateResult` struct: `FileMap map[string]string` + `DestDir string`
- `FileGenerator` interface: `Generate(outDir string, config *Config) (*GenerateResult, error)`
- `DiffAppFiles(result *GenerateResult) ([]string, error)` — for each file in the fileMap:
  - Skip binary files (by extension: `.wasm`, `.png`, `.jpg`, etc.)
  - Resolve dest path via `util.NormalizeFilePath(path.Join(result.DestDir, fileName))`
  - If dest doesn't exist, diff against `/dev/null` (shows full new file)
  - Run `diff -u <dest> <generated>` via `os/exec`
  - If output is non-empty, write to `<generatedPath>.diff`
  - Return list of filenames that had diffs
- `isBinaryFile(fileName string) bool` helper

### 2. Refactor all 9 apps with `Sync()` to implement `FileGenerator`

Each app gets a `Generate()` method that extracts the generation logic from `Sync()`. `Sync()` then calls `Generate()` + `SyncAppFiles()`.

| App | File | DestDir | Notes |
|-----|------|---------|-------|
| sketchybar | `apps/sketchybar/sketchybar.go` | `~/.config/sketchybar/` | |
| aerospace | `apps/aerospace/aerospace.go` | `~/.config/aerospace/` | |
| fastfetch | `apps/fastfetch/fastfetch.go` | `~/.config/fastfetch/` | |
| kitty | `apps/kitty/kitty.go` | `~/.config/kitty/` | |
| jankyborders | `apps/jankyborders/jankyborders.go` | `~/.config/borders/` | |
| nvim | `apps/nvim/nvim.go` | `~/.config/nvim/` | |
| terminal | `apps/terminal/terminal.go` | `~/.config/ohmyposh/` | |
| zellij | `apps/zellij/zellij.go` | `~/.config/zellij/` | Remote plugin fetch stays in `Sync()` only |
| desktoppr | `apps/desktoppr/desktoppr.go` | `~/.config/desktoppr/` | `exec.Command` side effect stays in `Sync()` only |

### 3. Create `cmd/diff/diff.go`

New cobra command that mirrors the sync command structure:
- Load config, create timestamped build dir
- For each enabled app implementing `FileGenerator`: call `Generate()` then `DiffAppFiles()`
- Also generate and diff the env file (`shizuku.sh`) against `~/.config/shizuku/shizuku.sh`
- Print summary to stdout showing which files have diffs
- Print path to build dir containing `.diff` files

### 4. Register in `cmd/main.go`

Import `diffcmd "github.com/eleonorayaya/shizuku/cmd/diff"` and add `rootCmd.AddCommand(diffcmd.DiffCommand)`.

## Output format

```
sketchybar:
  M sketchybarrc
  M styles/theme.sh
kitty:
  M kitty.conf
shizuku (env):
  M shizuku.sh

4 file(s) with differences. Diff files written to out/1739800000/
```

Diff files are placed alongside generated files: `out/{timestamp}/sketchybar/sketchybarrc.diff`

## Verification

1. `/task build` — confirm it compiles
2. `/task run diff` — run the diff command, verify it generates `.diff` files in the build dir
3. `/task test` — run existing tests
