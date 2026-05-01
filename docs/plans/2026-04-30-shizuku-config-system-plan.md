# Shizuku Config System Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the plain-text profile file and `ResolveProfile` with a proper YAML config package at `~/.config/shizuku/shizuku.yml`, with generalized env-override loading and a `shizuku config set/get` CLI.

**Architecture:** New `config/` package owns loading (YAML file → env override via reflection on struct tags), reading (field lookup by yaml tag name), and writing (load raw file → update field → marshal back). `cli.go` drops `ResolveProfile` and `profileCmd`, gains `configCmd` and a `defaultConfigPath` helper.

**Tech Stack:** Go, `gopkg.in/yaml.v3` (already in go.mod), `reflect`, cobra

---

## Files

| Action | Path |
|--------|------|
| Create | `config/config.go` |
| Create | `config/config_test.go` |
| Modify | `cli.go` |
| Delete | `cli_test.go` |

---

### Task 1: Config package tests

**Files:**
- Create: `config/config_test.go`

**Step 1: Write failing tests**

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/config"
)

// --- Load ---

func TestLoad_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "" {
		t.Errorf("expected empty profile, got %q", cfg.Profile)
	}
}

func TestLoad_ProfileFromFile(t *testing.T) {
	path := writeConfig(t, "profile: work\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	t.Setenv("SHIZUKU_PROFILE", "personal")
	path := writeConfig(t, "profile: work\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "personal" {
		t.Errorf("expected %q, got %q", "personal", cfg.Profile)
	}
}

func TestLoad_EnvWithNoFile(t *testing.T) {
	t.Setenv("SHIZUKU_PROFILE", "work")
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

// --- Get ---

func TestGet_KnownKey(t *testing.T) {
	cfg := &config.Config{Profile: "work"}
	val, err := config.Get(cfg, "profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "work" {
		t.Errorf("expected %q, got %q", "work", val)
	}
}

func TestGet_UnknownKey(t *testing.T) {
	cfg := &config.Config{}
	_, err := config.Get(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

// --- Set ---

func TestSet_WritesValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	if err := config.Set(path, "profile", "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != "work" {
		t.Errorf("expected %q, got %q", "work", cfg.Profile)
	}
}

func TestSet_UnknownKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	err := config.Set(path, "nonexistent", "val")
	if err == nil {
		t.Error("expected error for unknown key, got nil")
	}
}

func TestSet_DoesNotWriteEnvValues(t *testing.T) {
	// Set writes raw file values — env overrides must not bleed into file
	t.Setenv("SHIZUKU_PROFILE", "env-value")
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	// Set profile to something via file
	if err := config.Set(path, "profile", "file-value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Read raw file bytes — should contain file-value, not env-value
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "profile: file-value\n" {
		t.Errorf("unexpected file contents: %q", string(data))
	}
}

// --- YAML helper ---

func TestYAML_AllFields(t *testing.T) {
	cfg := &config.Config{Profile: "work"}
	out, err := cfg.YAML()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "profile: work\n" {
		t.Errorf("unexpected yaml: %q", out)
	}
}

// --- helpers ---

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "shizuku.yml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
```

**Step 2: Run tests to verify they fail**

Run: `task test -- ./config/...`
Expected: FAIL — package not found

---

### Task 2: Implement config package

**Files:**
- Create: `config/config.go`

**Step 1: Write the implementation**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Profile string `yaml:"profile" env:"SHIZUKU_PROFILE"`
}

// Load reads the YAML config at path (ok if missing), then overrides any field
// whose "env" tag names a non-empty environment variable.
func Load(path string) (*Config, error) {
	cfg, err := loadRaw(path)
	if err != nil {
		return nil, err
	}
	applyEnv(cfg)
	return cfg, nil
}

// Set loads the raw file (no env override), updates the field named by its yaml
// tag, and writes the result back.
func Set(path, key, value string) error {
	cfg, err := loadRaw(path)
	if err != nil {
		return err
	}
	if err := setField(cfg, key, value); err != nil {
		return err
	}
	return write(path, cfg)
}

// Get returns the value of the field whose yaml tag equals key.
func Get(cfg *Config, key string) (string, error) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("yaml") == key {
			return v.Field(i).String(), nil
		}
	}
	return "", fmt.Errorf("unknown config key %q", key)
}

// YAML marshals the config to a YAML string.
func (c *Config) YAML() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

func loadRaw(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

func applyEnv(cfg *Config) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if envKey := t.Field(i).Tag.Get("env"); envKey != "" {
			if val := os.Getenv(envKey); val != "" {
				v.Field(i).SetString(val)
			}
		}
	}
}

func setField(cfg *Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()
	for i := range t.NumField() {
		if t.Field(i).Tag.Get("yaml") == key {
			v.Field(i).SetString(value)
			return nil
		}
	}
	return fmt.Errorf("unknown config key %q", key)
}

func write(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}
```

Note: `for i := range t.NumField()` requires Go 1.22+. The go.mod already specifies `go 1.25.6`, so this is fine.

**Step 2: Run tests to verify they pass**

Run: `task test -- ./config/...`
Expected: PASS — all tests green

**Step 3: Commit**

```
feat(config): add YAML config package with env override and set/get
```

---

### Task 3: Update cli.go and delete cli_test.go

**Files:**
- Modify: `cli.go`
- Delete: `cli_test.go`

**Step 1: Replace cli.go**

Full replacement — every change explained inline:

```go
package shizuku

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/config"
	"github.com/spf13/cobra"
)

func defaultConfigPath() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "shizuku", "shizuku.yml")
}

func (b *Builder) Command() *cobra.Command {
	root := &cobra.Command{
		Use: "shizuku",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if b.opts.Verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
			}
			cfg, err := config.Load(defaultConfigPath())
			if err != nil {
				return err
			}
			if b.opts.Profile != "" {
				slog.Info("using profile", "profile", b.opts.Profile)
			} else if cfg.Profile != "" {
				b.opts.Profile = cfg.Profile
				slog.Info("using profile", "profile", b.opts.Profile)
			} else {
				slog.Warn("no profile set, using base profile")
			}
			return nil
		},
	}
	root.PersistentFlags().BoolVarP(&b.opts.Verbose, "verbose", "v", false, "Enable verbose output")
	root.PersistentFlags().StringVarP(&b.opts.Profile, "profile", "p", b.opts.Profile, "Profile to use (overlays on base)")

	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync all application configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Sync(context.Background())
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: "Show what would change on next sync",
		RunE: func(cmd *cobra.Command, args []string) error {
			report, err := b.Diff(context.Background())
			if err != nil {
				return err
			}

			if report.TotalChanged == 0 {
				fmt.Println("No differences found.")
				return nil
			}

			for _, r := range report.Results {
				fmt.Printf("%s:\n", r.Name)
				for _, f := range r.Changed {
					fmt.Printf("  M %s\n", f)
				}
			}

			fmt.Printf("\n%d file(s) with differences. Diff files written to %s/\n\n", report.TotalChanged, report.OutDir)

			for _, r := range report.Results {
				for _, f := range r.Changed {
					diffPath := r.FileMap[f] + ".diff"
					content, err := os.ReadFile(diffPath)
					if err != nil {
						return fmt.Errorf("failed to read diff file %s: %w", diffPath, err)
					}
					fmt.Println(string(content))
				}
			}
			return nil
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install application dependencies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Install(context.Background())
		},
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List apps active in the current profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			statuses := b.List()
			if b.opts.Profile != "" {
				fmt.Printf("Profile: %s\n\n", b.opts.Profile)
			} else {
				fmt.Println("Profile: (base)")
				fmt.Println()
			}
			for _, s := range statuses {
				fmt.Printf("  %-12s %s\n", s.Category, s.Name)
			}
			return nil
		},
	}

	root.AddCommand(syncCmd, diffCmd, installCmd, listCmd, configCmd())
	return root
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read and write shizuku configuration",
	}

	setCmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value in ~/.config/shizuku/shizuku.yml",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Set(defaultConfigPath(), args[0], args[1]); err != nil {
				return err
			}
			slog.Info("config updated", "key", args[0], "value", args[1])
			return nil
		},
	}

	getCmd := &cobra.Command{
		Use:   "get [key]",
		Short: "Get a config value, or show all as YAML",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(defaultConfigPath())
			if err != nil {
				return err
			}
			if len(args) == 0 {
				out, err := cfg.YAML()
				if err != nil {
					return err
				}
				fmt.Print(out)
				return nil
			}
			val, err := config.Get(cfg, args[0])
			if err != nil {
				return err
			}
			fmt.Println(val)
			return nil
		},
	}

	cmd.AddCommand(setCmd, getCmd)
	return cmd
}
```

**Step 2: Delete cli_test.go**

```bash
rm cli_test.go
```

**Step 3: Build and test**

Run: `task build`
Expected: success

Run: `task test`
Expected: PASS — config tests pass, no cli_test.go failures

**Step 4: Smoke-test**

```bash
task run -- config set profile work
# Expected: INFO config updated key=profile value=work

task run -- config get profile
# Expected: INFO using profile profile=work
#           work

task run -- config get
# Expected: INFO using profile profile=work
#           profile: work

task run -- list
# Expected: INFO using profile profile=work, work apps present
```

**Step 5: Commit**

```
refactor(cli): replace ResolveProfile+profileCmd with config package
```

---

## Verification

1. `shizuku config set profile work` → file written to `~/.config/shizuku/shizuku.yml`
2. `shizuku config get profile` → `work`
3. `shizuku config get` → `profile: work`
4. `shizuku list` (no flag) → INFO `profile=work`
5. `SHIZUKU_PROFILE=personal shizuku list` → INFO `profile=personal` (env wins)
6. `shizuku list --profile personal` → INFO `profile=personal` (flag wins)
7. `shizuku config set badkey val` → error `unknown config key "badkey"`
8. Remove `~/.config/shizuku/shizuku.yml`, run `shizuku list` → WARN base profile
