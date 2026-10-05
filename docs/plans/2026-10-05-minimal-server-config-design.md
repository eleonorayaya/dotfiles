# Minimal Server Config & Multi-Platform Installs — Design

## Goal

Install a small, server-safe subset of the shizuku config on Debian/Ubuntu servers (git, helix, bat, lsd, zsh + vi-mode + prompt, tuios) without the full desktop environment, while keeping the Mac config unchanged. This requires installs that work beyond Homebrew, and a way to get the binary onto servers without a Go toolchain.

## Decisions

- Supported package managers: **brew** (macOS) and **apt** (Debian/Ubuntu). Others (dnf, pacman, apk) are out of scope but must be a single-file addition.
- Tools missing from apt fall back to **GitHub release binaries** installed into `~/.local/bin`.
- The **base profile becomes the minimal server set**; desktop apps move into a `desktop` profile that `work` and `personal` extend.
- Shell stays **zsh-only**.
- The binary is distributed via **GitHub releases cut automatically on every merge to main**; `shizuku upgrade` downloads from releases when there is no source checkout.

## 1. Package-manager layer (`pkg/`)

```go
type Spec struct {
    Brew     string
    BrewCask bool
    Apt      string
    AptBin   string   // binary name apt installs, if different (bat → batcat)
    Bin      string   // binary name used for idempotency checks
    Release  *Release
}

type Release struct {
    Repo  string            // "helix-editor/helix"
    Asset string            // regex with {arch}: `helix-.*-{arch}-linux\.tar\.xz`
    Arch  map[string]string // GOARCH → token; default amd64→x86_64, arm64→aarch64
    Bin   string            // binary inside archive, or the raw asset itself
    Dirs  map[string]string // extra archive dirs to copy, e.g. "runtime" → "~/.config/helix/runtime"
}

func Install(s Spec) error
```

Resolution order:

1. **darwin** → brew (existing behavior).
2. **linux + apt available** → if `Apt` is set and `apt-cache policy` shows a candidate, `apt-get install -y`.
3. **fallback** → if `Release` is set, fetch the latest GitHub release, match the asset for the current arch, extract the binary into `~/.local/bin`. Extraction shells out to `tar` (handles `.gz` and `.xz`); raw binaries are chmod'd.
4. Otherwise return an error naming the app.

Details:

- **Idempotency:** skip if `Bin` resolves via `exec.LookPath` **or** exists at `~/.local/bin/<bin>` (on first run `~/.local/bin` is not on PATH yet).
- **apt bootstrap:** once per install run, before any probe: `apt-get update`, with `DEBIAN_FRONTEND=noninteractive`. Use `sudo` only when uid ≠ 0 (containers run as root without sudo).
- **AptBin:** when apt installs under a different name, symlink `~/.local/bin/<Bin>` → the apt binary so aliases work identically everywhere.
- **Managers** are an internal interface (`brew`, `apt`, `release`) chosen once per run; adding dnf is one file.
- **Release downloads** use the GitHub API anonymously, sending `GITHUB_TOKEN` when set to avoid rate limits.
- `util.InstallBrewPackage` becomes a thin wrapper over `pkg.Install(pkg.Spec{Brew: ...})`; desktop-only apps keep calling it unchanged.
- `Builder.Install` continues past per-app failures and returns an aggregated error at the end.

Verified asset names (latest releases, Linux):

| Tool | amd64 | arm64 | Format |
|---|---|---|---|
| helix | `helix-25.07.1-x86_64-linux.tar.xz` | `aarch64` | tar.xz, needs `runtime/` |
| tuios | `tuios_0.8.5_Linux_x86_64.tar.gz` | `Linux_arm64` | tar.gz |
| oh-my-posh | `posh-linux-amd64` | `posh-linux-arm64` | raw binary |
| lsd | `lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz` | `aarch64` | tar.gz |
| bat | `bat-v0.26.1-x86_64-unknown-linux-gnu.tar.gz` | `aarch64` | tar.gz |

## 2. Profiles and app changes

### Library: profile inheritance

```go
shizuku.WithProfile("desktop", WithLanguages(...), WithPrograms(...), WithAgents(claude...)),
shizuku.WithProfile("work",     shizuku.Extends("desktop"), WithPrograms(acli, k9s, ...)),
shizuku.WithProfile("personal", shizuku.Extends("desktop"), WithPrograms(proton..., notion(MCP on))),
```

`activeProfile()` resolves base → extended chain → selected profile using the existing `mergeNamed`, so later layers still override by name. Cycles and unknown profile names (selected or extended) are errors, not a silent fallback to base.

### `examples/eleonora/main.go`

- **base** (in this order so prerequisites install first): git, terminal, helix, bat, lsd, tuios
- **desktop:** all languages, the claude agent, and all remaining programs
- **work / personal:** unchanged contents, now `Extends("desktop")`

The Mac must have `profile: personal` or `work` set in `~/.config/shizuku/shizuku.yml`; with no profile it gets only the server set.

### Base app changes

| App | Change |
|---|---|
| git | Installs `git` via pkg. gh extension/skill install is skipped with a log line when `gh` is absent. |
| terminal | Installs `zsh`, `ca-certificates`, `xz-utils` (apt-only specs; no-ops on macOS). Replaces antigen with a git clone of `jeffreytse/zsh-vi-mode` into `~/.local/share/shizuku/plugins/` sourced directly. oh-my-posh via pkg (brew tap / raw release). Ensures source lines in `~/.zshenv` and `~/.zshrc` (see §3). |
| helix | pkg spec with apt `hx` + release fallback; release install copies `runtime/` to `~/.config/helix/runtime` (sync never prunes, so this is safe). |
| bat | apt `bat` with `AptBin: batcat` + release fallback. |
| lsd | apt `lsd` + release fallback. |
| tuios | release fallback (not in apt). Config dest becomes `os.UserConfigDir()/tuios/` — matches tuios's `adrg/xdg` resolution on both OSes. |

### Guarded env

Apps whose install can fail per-app must not break the shell. Aliases and editor vars that point at an optional binary are emitted guarded:

```zsh
command -v lsd >/dev/null && alias ls='lsd'
command -v hx  >/dev/null && export EDITOR=hx VISUAL=hx
```

Implemented as an optional `Requires string` field on `app.Alias` and `app.EnvVar`; `GenerateEnvFiles` wraps entries that set it.

## 3. Shell, distribution, upgrade

### Shell wiring

`terminal.Install` idempotently appends (only if the exact line is absent):

- `source ~/.config/shizuku/shizuku.zshenv` → `~/.zshenv`
- `source ~/.config/shizuku/shizuku.sh` → `~/.zshrc`

It never runs `chsh`; it logs a hint when `$SHELL` is not zsh.

### Releases

`.goreleaser.yaml`:

- `project_name: shizuku`
- build `examples/eleonora`, `CGO_ENABLED=0`, targets linux/amd64, linux/arm64, darwin/arm64
- archive `name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"` (no version → stable `releases/latest/download/` URLs)
- checksum `name_template: checksums.txt`

`.github/workflows/release.yml`:

- `on: push: branches: [main]`, `paths-ignore` for `docs/**`, `.claude/**`, `**/*.md`
- `permissions: contents: write`, checkout with `fetch-depth: 0`
- compute next tag: latest `v*` patch + 1, starting at `v0.1.0`
- create and push the tag in the same job, then run `goreleaser release --clean` (a tag pushed with `GITHUB_TOKEN` does not trigger another workflow, so everything happens in one job)

`task build:linux` cross-compiles `out/shizuku-linux-{amd64,arm64}` locally for testing on a box before merge.

### `shizuku upgrade`

- `~/.local/src/shizuku` exists → current behavior (git pull + `task install`).
- otherwise, or with `--from-release` → download `shizuku_<os>_<arch>.tar.gz` from the latest release, verify against `checksums.txt`, write to a temp file beside `os.Executable()`, and rename over it. Reuses the pkg release downloader.

### Server bootstrap (README)

```sh
mkdir -p ~/.local/bin
curl -fsSL https://github.com/eleonorayaya/dotfiles/releases/latest/download/shizuku_linux_$(dpkg --print-architecture).tar.gz \
  | tar -xz -C ~/.local/bin shizuku
~/.local/bin/shizuku install && ~/.local/bin/shizuku sync
```

## Testing

Unit:

- asset matching against the real asset names above × amd64/arm64
- manager selection with OS/apt/uid detection injected
- idempotency check honoring `~/.local/bin`
- `Extends` resolution: override-by-name, multi-level chain, cycle, unknown profile
- guarded alias/env generation
- idempotent rc-file appends
- checksum verification and upgrade mode selection

Integration:

- cross-compile, then run `install` + `sync` in `debian:bookworm` and `ubuntu:24.04` containers as root; assert `hx`, `bat`, `lsd`, `tuios`, `oh-my-posh` resolve and `zsh -i -c exit` succeeds with no errors.
- Mac regression: `shizuku diff -p personal` shows only the expected terminal changes.

## Out of scope

- dnf / pacman / apk
- bash support
- auto-updating release-installed tools (delete the binary and re-run install)
- the `go.mod` module path (`github.com/eleonorayaya/shizuku`) not matching the GitHub repo (`eleonorayaya/dotfiles`)
