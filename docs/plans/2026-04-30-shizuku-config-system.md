# Shizuku Config System Design

**Goal:** Replace the plain-text `~/.config/shizuku/profile` file with a proper YAML config at `~/.config/shizuku/shizuku.yml`, backed by a generalized config loading system (file + env override) and a `shizuku config set/get` CLI.

---

## Config struct

New `config/` package. Fields carry both `yaml` and `env` tags:

```go
type Config struct {
    Profile string `yaml:"profile" env:"SHIZUKU_PROFILE"`
}
```

New fields are added here and become immediately available via `config set/get` and env override — no other code changes needed.

---

## Load

```go
func Load(path string) (*Config, error)
```

1. Read YAML from `path` into `Config` (silent no-op if file doesn't exist)
2. Walk struct fields via reflection; for each `env` tag, call `os.Getenv` and override if non-empty

Resolution order: **env var > YAML file > zero value**

The `--profile` CLI flag (handled in `PersistentPreRunE`) sits above all of these.

---

## Set / Get helpers

```go
func Set(path, key, value string) error  // load → update field → write YAML
func Get(cfg *Config, key string) (string, error)  // field lookup by yaml tag
```

Both use reflection to match the `yaml` tag name against `key`. Unknown keys return a descriptive error. `Set` reads the existing file first so other fields are preserved on write.

---

## CLI: `shizuku config`

Replaces `shizuku profile set/get`.

```
shizuku config set <key> <value>   # writes to ~/.config/shizuku/shizuku.yml
shizuku config get [key]           # prints one field or all fields as YAML
```

Examples:
```
shizuku config set profile work
shizuku config get profile         # → work
shizuku config get                 # → profile: work
```

---

## Integration with PersistentPreRunE

```go
cfgPath := defaultConfigPath()  // ~/.config/shizuku/shizuku.yml
cfg, err := config.Load(cfgPath)
// flag beats config
if b.opts.Profile != "" {
    slog.Info("using profile", "profile", b.opts.Profile)
} else if cfg.Profile != "" {
    b.opts.Profile = cfg.Profile
    slog.Info("using profile", "profile", b.opts.Profile)
} else {
    slog.Warn("no profile set, using base profile")
}
```

---

## What gets removed

- `ResolveProfile()` function in `cli.go`
- `profileCmd()` (the `shizuku profile set/get` commands)
- `cli_test.go` tests for `ResolveProfile` (replaced by `config/config_test.go`)
- `~/.config/shizuku/profile` plain-text file convention

---

## Files touched

| Action | Path |
|--------|------|
| Create | `config/config.go` |
| Create | `config/config_test.go` |
| Modify | `cli.go` — remove `ResolveProfile` + `profileCmd`, add `configCmd`, update `PersistentPreRunE` |
| Delete | `cli_test.go` |
