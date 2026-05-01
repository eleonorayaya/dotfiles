# Default Profile Per Device Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Allow a per-device default profile so `--profile` is not required on every command, with logging of the active profile and a warning when falling back to the base profile.

**Architecture:** Add a `~/.config/shizuku/profile` plain-text file as a third resolution tier (flag → env → file → base). Extract `ResolveProfile()` with injectable inputs for full test coverage of all priority tiers, add logging in `PersistentPreRunE`, and add a `shizuku profile` command group (`set`/`get`) to manage the file.

**Tech Stack:** Go, `log/slog`, `os`, cobra

---

## Profile resolution priority

```
--profile flag  →  SHIZUKU_PROFILE env var  →  ~/.config/shizuku/profile file  →  base (warn)
```

## Critical files

- Modify: `cli.go` — add exported `ResolveProfile(flag, env, profileFilePath string)`, update `PersistentPreRunE`, add `profileCmd()`
- Create: `cli_test.go` — unit tests covering all 4 resolution tiers + whitespace handling

---

### Task 1: Test `ResolveProfile`

**Files:**
- Create: `cli_test.go`

**Step 1: Write the failing tests**

```go
package shizuku_test

import (
	"os"
	"path/filepath"
	"testing"

	shizuku "github.com/eleonorayaya/shizuku"
)

func writeProfileFile(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, "profile")
	if err := os.WriteFile(path, []byte(name+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Flag takes priority over everything
func TestResolveProfile_FlagWins(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("work", "personal", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}

// Env var takes priority over file, loses to flag
func TestResolveProfile_EnvWinsOverFile(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("", "work", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}

// File is used when flag and env are empty
func TestResolveProfile_FileUsedWhenNoFlagOrEnv(t *testing.T) {
	dir := t.TempDir()
	profileFile := writeProfileFile(t, dir, "personal")

	got, err := shizuku.ResolveProfile("", "", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "personal" {
		t.Errorf("expected %q, got %q", "personal", got)
	}
}

// Returns empty string when nothing is configured (base profile)
func TestResolveProfile_EmptyWhenNothingSet(t *testing.T) {
	dir := t.TempDir()
	profileFile := filepath.Join(dir, "profile") // does not exist

	got, err := shizuku.ResolveProfile("", "", profileFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// Whitespace in the profile file is stripped
func TestResolveProfile_FileWhitespaceStripped(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile")
	if err := os.WriteFile(path, []byte("  work  \n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := shizuku.ResolveProfile("", "", path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "work" {
		t.Errorf("expected %q, got %q", "work", got)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `/task test`
Expected: FAIL — `shizuku.ResolveProfile` undefined

**Step 3: Implement `ResolveProfile` in `cli.go`**

Add these imports to `cli.go` (strings and path/filepath):
```go
import (
    "context"
    "fmt"
    "log/slog"
    "os"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)
```

Add the function:
```go
// ResolveProfile determines the active profile name using priority:
// flag > env > profile file > "" (base)
func ResolveProfile(flag, env, profileFilePath string) (string, error) {
    if flag != "" {
        return flag, nil
    }
    if env != "" {
        return env, nil
    }
    data, err := os.ReadFile(profileFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", nil
        }
        return "", fmt.Errorf("failed to read default profile: %w", err)
    }
    return strings.TrimSpace(string(data)), nil
}
```

**Step 4: Run tests to verify they pass**

Run: `/task test`
Expected: PASS

**Step 5: Commit**

```
feat: add ResolveProfile with full priority-tier test coverage
```

---

### Task 2: Wire profile resolution and logging into `PersistentPreRunE`

**Files:**
- Modify: `cli.go:15-23`

**Step 1: Update `PersistentPreRunE`**

Replace the existing `PersistentPreRunE` body (currently lines 15–22) with:

```go
PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if b.opts.Verbose {
        slog.SetLogLoggerLevel(slog.LevelDebug)
    }
    profileFile := filepath.Join(os.Getenv("HOME"), ".config", "shizuku", "profile")
    profile, err := ResolveProfile(b.opts.Profile, os.Getenv("SHIZUKU_PROFILE"), profileFile)
    if err != nil {
        return err
    }
    b.opts.Profile = profile
    if b.opts.Profile != "" {
        slog.Info("using profile", "profile", b.opts.Profile)
    } else {
        slog.Warn("no profile set, using base profile")
    }
    return nil
},
```

**Step 2: Build and verify it compiles**

Run: `/task build`
Expected: success

**Step 3: Smoke-test the logging**

Run: `/task run list`
Expected output includes `WARN no profile set, using base profile`

Run: `/task run list --profile work`
Expected output includes `INFO using profile profile=work`

**Step 4: Commit**

```
feat: wire default profile file and add profile logging
```

---

### Task 3: Add `shizuku profile set` / `shizuku profile get` commands

**Files:**
- Modify: `cli.go` — add `profileCmd()` function and register it

**Step 1: Add `profileCmd()` to `cli.go`**

Add this function at the bottom of `cli.go`:

```go
func profileCmd() *cobra.Command {
    profileFile := filepath.Join(os.Getenv("HOME"), ".config", "shizuku", "profile")

    cmd := &cobra.Command{
        Use:   "profile",
        Short: "Manage the default profile for this device",
    }

    setCmd := &cobra.Command{
        Use:   "set <name>",
        Short: "Set the default profile for this device",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            if err := os.MkdirAll(filepath.Dir(profileFile), 0755); err != nil {
                return fmt.Errorf("failed to create config dir: %w", err)
            }
            if err := os.WriteFile(profileFile, []byte(args[0]+"\n"), 0644); err != nil {
                return fmt.Errorf("failed to write profile: %w", err)
            }
            slog.Info("default profile set", "profile", args[0])
            return nil
        },
    }

    getCmd := &cobra.Command{
        Use:   "get",
        Short: "Show the default profile for this device",
        RunE: func(cmd *cobra.Command, args []string) error {
            name, err := ResolveProfile("", "", profileFile)
            if err != nil {
                return err
            }
            if name == "" {
                fmt.Println("(none - using base profile)")
            } else {
                fmt.Println(name)
            }
            return nil
        },
    }

    cmd.AddCommand(setCmd, getCmd)
    return cmd
}
```

**Step 2: Register `profileCmd` in `Command()`**

In `Command()`, add `profileCmd()` to `root.AddCommand(...)`:
```go
root.AddCommand(syncCmd, diffCmd, installCmd, listCmd, profileCmd())
```

**Step 3: Build**

Run: `/task build`
Expected: success

**Step 4: Smoke-test the new commands**

```bash
/task run profile set work
# Expected: INFO default profile set profile=work

/task run profile get
# Expected: work

/task run list
# Expected: INFO using profile profile=work

/task run profile set ""   # edge: empty arg is rejected by ExactArgs(1)

cat ~/.config/shizuku/profile
# Expected: work
```

**Step 5: Commit**

```
feat: add 'shizuku profile set/get' commands for per-device default
```

---

## Verification

End-to-end test sequence (no `--profile` flag anywhere):

1. `shizuku profile set work` → INFO log confirming
2. `shizuku list` → INFO showing `profile=work`, apps include work-profile additions (buildkite, datadog, k9s, mise, ruby)
3. `shizuku profile get` → prints `work`
4. Remove `~/.config/shizuku/profile`, run `shizuku list` → WARN about base profile
5. `SHIZUKU_PROFILE=personal shizuku list` → INFO `profile=personal`, env overrides absent file
6. `shizuku list --profile work` with file absent → INFO `profile=work`, flag overrides env and file
