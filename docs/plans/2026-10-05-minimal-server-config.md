# Minimal Server Config Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the shizuku base profile a minimal, server-safe set (git, terminal/zsh, helix, bat, lsd, tuios) installable on Debian/Ubuntu via apt or GitHub release binaries, move desktop apps into an extendable `desktop` profile, and ship the binary through auto-cut GitHub releases with a release-aware `shizuku upgrade`.

**Architecture:** A new `pkg/` package owns installation: one `Spec` per tool lists brew/apt/release sources, and an `Installer` picks the first method that applies on the current OS. Profiles gain single-parent `Extends` so `work`/`personal` → `desktop` → base. Env generation gains `Requires` guards so a failed optional install never breaks `ls`/`cat`/`$EDITOR`.

**Tech Stack:** Go 1.25 stdlib (`net/http`, `regexp`, `os.CopyFS`, `crypto/sha256`), system `tar`, Cobra, GoReleaser v2, GitHub Actions.

**Design doc:** `docs/plans/2026-10-05-minimal-server-config-design.md`

---

## Conventions for every task

- All paths are relative to the repo root (`dotfiles/`). App packages live at top-level `languages/`, `programs/`, `agents/` — **not** `apps/...` as CLAUDE.md claims.
- **Never** run `go test`/`go build`/`go run` directly. Use Task (see the `task` skill):
  - Tests: `task test --force -- -run <Regex>` (the task runs `go test <args> ./...`). `--force` avoids Task's up-to-date skip.
  - Full suite: `task test --force`
  - Format: `task lint`
  - Build: `task build`
- No code comments (repo style). Wrap errors with `fmt.Errorf("context: %w", err)`.
- Every commit message ends with:
  ```
  Co-Authored-By: Claude Opus 5.5 (1M context) <noreply@anthropic.com>
  ```
- Deviation from the design doc: `Release.Bin` is dropped; `Spec.Bin` names both the binary inside the archive and the installed name (one field instead of two that must agree).

---

### Task 1: Taskfile — include new sources, add `build:linux`

**Files:**
- Modify: `Taskfile.yml`

**Step 1: Edit `Taskfile.yml`**

Replace the `vars` block and add a `build:linux` task after `build`, and make `test` watch root-level Go files:

```yaml
vars:
  GO_SOURCES: '{agents,programs,languages,examples,app,config,util,pkg}/**/*.go'
  MAIN: examples/eleonora/main.go
```

```yaml
  build:linux:
    sources:
      - "{{.GO_SOURCES}}"
      - "*.go"
    generates:
      - out/shizuku-linux-amd64
      - out/shizuku-linux-arm64
    cmds:
      - GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o out/shizuku-linux-amd64 {{.MAIN}}
      - GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o out/shizuku-linux-arm64 {{.MAIN}}
```

In `build`, `lint` and `test`, replace `- shizuku.go` (where present) / add `- "*.go"` under `sources` so root files (`shizuku.go`, `cli.go`, root tests) are tracked.

**Step 2: Verify**

Run: `task build:linux && file out/shizuku-linux-amd64 out/shizuku-linux-arm64`
Expected: both reported as `ELF 64-bit LSB executable`, x86-64 and ARM aarch64 respectively.

**Step 3: Commit**

```bash
git add Taskfile.yml
git commit -m "build: add build:linux task and track pkg/ and root sources"
```

---

### Task 2: Profile inheritance (`Extends`)

**Files:**
- Modify: `shizuku.go` (Profile struct ~L22, options ~L55-90, `activeProfile` ~L117, `Sync`/`Diff`/`Install`/`List`)
- Modify: `cli.go` (`listCmd` ~L104)
- Create: `shizuku_test.go`

**Step 1: Write the failing tests** — `shizuku_test.go`

```go
package shizuku

import (
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
)

type fakeProgram struct {
	name string
	tag  string
}

func (f fakeProgram) Name() string { return f.name }

func programNames(p Profile) []string {
	out := []string{}
	for _, prog := range p.Programs {
		out = append(out, prog.Name())
	}
	return out
}

func newTestBuilder(profile string) *Builder {
	return New(
		WithProfileName(profile),
		WithPrograms(fakeProgram{name: "git"}, fakeProgram{name: "notion", tag: "base"}),
		WithProfile("desktop",
			WithPrograms(fakeProgram{name: "kitty"}),
		),
		WithProfile("work",
			Extends("desktop"),
			WithPrograms(fakeProgram{name: "k9s"}, fakeProgram{name: "notion", tag: "work"}),
		),
		WithProfile("loop-a", Extends("loop-b")),
		WithProfile("loop-b", Extends("loop-a")),
		WithProfile("orphan", Extends("missing")),
	)
}

func TestActiveProfile_BaseWhenUnset(t *testing.T) {
	p, err := newTestBuilder("").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_SingleLayer(t *testing.T) {
	p, err := newTestBuilder("desktop").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion,kitty" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_ExtendsChain(t *testing.T) {
	p, err := newTestBuilder("work").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := strings.Join(programNames(p), ","); got != "git,notion,kitty,k9s" {
		t.Errorf("got %q", got)
	}
}

func TestActiveProfile_LaterLayerOverridesByName(t *testing.T) {
	p, err := newTestBuilder("work").activeProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, prog := range p.Programs {
		if prog.Name() == "notion" && prog.(fakeProgram).tag != "work" {
			t.Errorf("notion not overridden, tag=%q", prog.(fakeProgram).tag)
		}
	}
}

func TestActiveProfile_UnknownProfile(t *testing.T) {
	if _, err := newTestBuilder("nope").activeProfile(); err == nil {
		t.Fatal("expected error for unknown profile")
	}
}

func TestActiveProfile_UnknownParent(t *testing.T) {
	if _, err := newTestBuilder("orphan").activeProfile(); err == nil {
		t.Fatal("expected error for unknown parent profile")
	}
}

func TestActiveProfile_Cycle(t *testing.T) {
	_, err := newTestBuilder("loop-a").activeProfile()
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

var _ app.Program = fakeProgram{}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestActiveProfile`
Expected: build failure — `undefined: Extends`, and `activeProfile` returns one value.

**Step 3: Implement**

In `shizuku.go`:

1. Add the field to `Profile`:
   ```go
   type Profile struct {
   	Extends   string
   	Languages []app.Language
   	Programs  []app.Program
   	Agents    []app.Agent
   }
   ```
2. Add the option after `WithAgents`:
   ```go
   func Extends(name string) Option {
   	return func(b *Builder) {
   		b.target.Extends = name
   	}
   }
   ```
3. Replace `activeProfile`:
   ```go
   func (b *Builder) activeProfile() (Profile, error) {
   	if b.opts.Profile == "" {
   		return b.base, nil
   	}

   	chain := []*Profile{}
   	seen := map[string]bool{}
   	for name := b.opts.Profile; name != ""; {
   		if seen[name] {
   			return Profile{}, fmt.Errorf("profile cycle detected at %q", name)
   		}
   		seen[name] = true
   		p, ok := b.profiles[name]
   		if !ok {
   			return Profile{}, fmt.Errorf("unknown profile %q", name)
   		}
   		chain = append(chain, p)
   		name = p.Extends
   	}

   	out := b.base
   	for i := len(chain) - 1; i >= 0; i-- {
   		out = Profile{
   			Languages: mergeNamed(out.Languages, chain[i].Languages),
   			Programs:  mergeNamed(out.Programs, chain[i].Programs),
   			Agents:    mergeNamed(out.Agents, chain[i].Agents),
   		}
   	}
   	return out, nil
   }
   ```
4. In `Sync`, `Diff`, `Install`, replace `profile := b.activeProfile()` with:
   ```go
   profile, err := b.activeProfile()
   if err != nil {
   	return err            // Diff: return nil, err
   }
   ```
   (`err` is already declared in those functions from `resolveOutDir`, so use `=` / `:=` as the compiler requires.)
5. Change `List`:
   ```go
   func (b *Builder) List() ([]AppStatus, error) {
   	profile, err := b.activeProfile()
   	if err != nil {
   		return nil, err
   	}
   	...
   	return statuses, nil
   }
   ```

In `cli.go` `listCmd.RunE`:
```go
statuses, err := b.List()
if err != nil {
	return err
}
```

**Step 4: Run tests**

Run: `task test --force -- -run TestActiveProfile`
Expected: PASS (7 tests).

**Step 5: Commit**

```bash
git add shizuku.go cli.go shizuku_test.go
git commit -m "feat: let profiles extend other profiles"
```

---

### Task 3: `Install` continues past failures

**Files:**
- Modify: `shizuku.go` (`Install`, ~L360)
- Modify: `shizuku_test.go`

**Step 1: Write the failing test** — append to `shizuku_test.go`:

```go
type failingInstaller struct {
	name  string
	calls *[]string
}

func (f failingInstaller) Name() string { return f.name }

func (f failingInstaller) Install(ctx *app.Context) error {
	*f.calls = append(*f.calls, f.name)
	return fmt.Errorf("%s broke", f.name)
}

func TestInstall_ContinuesAndJoinsErrors(t *testing.T) {
	calls := []string{}
	b := New(
		WithOutDir(t.TempDir()),
		WithPrograms(failingInstaller{name: "a", calls: &calls}, failingInstaller{name: "b", calls: &calls}),
	)
	err := b.Install(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Join(calls, ",") != "a,b" {
		t.Errorf("expected both installers to run, got %v", calls)
	}
	if !strings.Contains(err.Error(), "a broke") || !strings.Contains(err.Error(), "b broke") {
		t.Errorf("expected both errors, got %v", err)
	}
}
```
Add `"context"` and `"fmt"` to the test imports.

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestInstall_ContinuesAndJoinsErrors`
Expected: FAIL — `expected both installers to run, got [a]`.

**Step 3: Implement** — in `Builder.Install`, replace the three loops' early returns:

```go
	var errs []error
	for _, l := range profile.Languages {
		if err := install(l); err != nil {
			slog.Error("install failed", "appName", l.Name(), "err", err)
			errs = append(errs, err)
		}
	}
	for _, p := range profile.Programs {
		if err := install(p); err != nil {
			slog.Error("install failed", "appName", p.Name(), "err", err)
			errs = append(errs, err)
		}
	}
	for _, a := range profile.Agents {
		if err := install(a); err != nil {
			slog.Error("install failed", "appName", a.Name(), "err", err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
```
Add `"errors"` to imports.

**Step 4: Run tests**

Run: `task test --force -- -run 'TestInstall_|TestActiveProfile'`
Expected: PASS.

**Step 5: Commit**

```bash
git add shizuku.go shizuku_test.go
git commit -m "feat: keep installing remaining apps when one fails"
```

---

### Task 4: `pkg` — Spec, Release, asset matching

**Files:**
- Create: `pkg/pkg.go`
- Create: `pkg/pkg_test.go`

**Step 1: Write the failing test** — `pkg/pkg_test.go`

Asset names are the real ones from the latest releases (verified 2026-10-04).

```go
package pkg

import "testing"

func TestMatchAsset(t *testing.T) {
	cases := []struct {
		name    string
		release Release
		assets  []string
		goarch  string
		want    string
	}{
		{
			name:    "helix amd64",
			release: Release{Repo: "helix-editor/helix", Asset: `helix-.*-{arch}-linux\.tar\.xz`},
			assets:  []string{"helix-25.07.1-aarch64-linux.tar.xz", "helix-25.07.1-x86_64-linux.tar.xz", "helix-25.07.1-x86_64-macos.tar.xz"},
			goarch:  "amd64",
			want:    "helix-25.07.1-x86_64-linux.tar.xz",
		},
		{
			name:    "helix arm64",
			release: Release{Repo: "helix-editor/helix", Asset: `helix-.*-{arch}-linux\.tar\.xz`},
			assets:  []string{"helix-25.07.1-aarch64-linux.tar.xz", "helix-25.07.1-x86_64-linux.tar.xz"},
			goarch:  "arm64",
			want:    "helix-25.07.1-aarch64-linux.tar.xz",
		},
		{
			name:    "tuios arm64 skips ghostty and web variants",
			release: Release{Repo: "Gaurav-Gosain/tuios", Asset: `tuios_[0-9.]+_Linux_{arch}\.tar\.gz`, Arch: map[string]string{"amd64": "x86_64", "arm64": "arm64"}},
			assets:  []string{"tuios-ghostty_0.8.5_Linux_arm64.tar.gz", "tuios-web_0.8.5_Linux_arm64.tar.gz", "tuios_0.8.5_Linux_arm64.tar.gz", "tuios_0.8.5_Linux_x86_64.tar.gz"},
			goarch:  "arm64",
			want:    "tuios_0.8.5_Linux_arm64.tar.gz",
		},
		{
			name:    "oh-my-posh raw amd64",
			release: Release{Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`, Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"}},
			assets:  []string{"posh-linux-amd64", "posh-linux-arm", "posh-linux-arm64"},
			goarch:  "amd64",
			want:    "posh-linux-amd64",
		},
		{
			name:    "oh-my-posh raw arm64 does not match arm",
			release: Release{Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`, Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"}},
			assets:  []string{"posh-linux-arm", "posh-linux-arm64"},
			goarch:  "arm64",
			want:    "posh-linux-arm64",
		},
		{
			name:    "lsd prefers gnu over musl",
			release: Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
			assets:  []string{"lsd-v1.2.0-x86_64-unknown-linux-musl.tar.gz", "lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz"},
			goarch:  "amd64",
			want:    "lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz",
		},
		{
			name:    "bat skips deb",
			release: Release{Repo: "sharkdp/bat", Asset: `bat-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
			assets:  []string{"bat-musl_0.26.1_musl-linux-amd64.deb", "bat-v0.26.1-aarch64-unknown-linux-gnu.tar.gz"},
			goarch:  "arm64",
			want:    "bat-v0.26.1-aarch64-unknown-linux-gnu.tar.gz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.release.matchAsset(tc.goarch, tc.assets)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMatchAsset_NoMatch(t *testing.T) {
	r := Release{Repo: "x/y", Asset: `thing-{arch}`}
	if _, err := r.matchAsset("amd64", []string{"thing-aarch64"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestMatchAsset_UnsupportedArch(t *testing.T) {
	r := Release{Repo: "x/y", Asset: `thing-{arch}`}
	if _, err := r.matchAsset("riscv64", []string{"thing-riscv64"}); err == nil {
		t.Fatal("expected error")
	}
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestMatchAsset`
Expected: build failure — `undefined: Release`.

**Step 3: Implement** — `pkg/pkg.go`

```go
package pkg

import (
	"fmt"
	"regexp"
	"strings"
)

type Spec struct {
	Brew     string
	BrewCask bool
	Apt      string
	AptBin   string
	Bin      string
	Release  *Release
}

type Release struct {
	Repo  string
	Asset string
	Arch  map[string]string
	Dirs  map[string]string
}

var defaultArch = map[string]string{
	"amd64": "x86_64",
	"arm64": "aarch64",
}

func (s Spec) name() string {
	for _, n := range []string{s.Bin, s.Apt, s.Brew} {
		if n != "" {
			return n
		}
	}
	if s.Release != nil {
		return s.Release.Repo
	}
	return "unnamed package"
}

func (r *Release) matchAsset(goarch string, names []string) (string, error) {
	token, ok := r.Arch[goarch]
	if !ok {
		token, ok = defaultArch[goarch]
	}
	if !ok {
		return "", fmt.Errorf("unsupported architecture %s for %s", goarch, r.Repo)
	}

	re, err := regexp.Compile("^" + strings.ReplaceAll(r.Asset, "{arch}", regexp.QuoteMeta(token)) + "$")
	if err != nil {
		return "", fmt.Errorf("invalid asset pattern for %s: %w", r.Repo, err)
	}

	for _, n := range names {
		if re.MatchString(n) {
			return n, nil
		}
	}
	return "", fmt.Errorf("no release asset in %s matches %s", r.Repo, re)
}
```

**Step 4: Run tests**

Run: `task test --force -- -run TestMatchAsset`
Expected: PASS (all subtests).

**Step 5: Commit**

```bash
git add pkg/
git commit -m "feat(pkg): add install spec and release asset matching"
```

---

### Task 5: `pkg` — Installer core: idempotency, brew, apt

**Files:**
- Create: `pkg/installer.go`
- Create: `pkg/installer_test.go`

**Step 1: Write the failing tests** — `pkg/installer_test.go`

```go
package pkg

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	paths   map[string]string
	outputs map[string]string
	fail    map[string]bool
	calls   []string
}

func (f *fakeRunner) run(env []string, name string, args ...string) ([]byte, error) {
	cmd := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, cmd)
	if f.fail[cmd] {
		return nil, errors.New("command failed")
	}
	return []byte(f.outputs[cmd]), nil
}

func (f *fakeRunner) lookPath(name string) (string, error) {
	if p, ok := f.paths[name]; ok {
		return p, nil
	}
	return "", exec.ErrNotFound
}

func testInstaller(t *testing.T, goos string) (*Installer, *fakeRunner) {
	t.Helper()
	fr := &fakeRunner{paths: map[string]string{}, outputs: map[string]string{}, fail: map[string]bool{}}
	return &Installer{
		GOOS:     goos,
		GOARCH:   "amd64",
		Home:     t.TempDir(),
		HTTP:     http.DefaultClient,
		lookPath: fr.lookPath,
		run:      fr.run,
	}, fr
}

func TestInstall_SkipsWhenOnPath(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	fr.paths["bat"] = "/opt/homebrew/bin/bat"
	if err := i.Install(Spec{Brew: "bat", Bin: "bat"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no commands, got %v", fr.calls)
	}
}

func TestInstall_SkipsWhenInLocalBin(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	if err := os.MkdirAll(i.BinDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(i.BinDir(), "hx"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := i.Install(Spec{Apt: "hx", Bin: "hx"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no commands, got %v", fr.calls)
	}
}

func TestInstall_DarwinUsesBrew(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	fr.fail["brew list bat"] = true
	if err := i.Install(Spec{Brew: "bat", Apt: "bat", Bin: "bat"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "brew list bat|brew install bat"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInstall_DarwinCask(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	fr.fail["brew list aerospace --cask"] = true
	if err := i.Install(Spec{Brew: "aerospace", BrewCask: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := fr.calls[len(fr.calls)-1]; got != "brew install aerospace --cask" {
		t.Errorf("got %q", got)
	}
}

func TestInstall_DarwinIgnoresAptOnlySpec(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	if err := i.Install(Spec{Apt: "xz-utils", Bin: "xz"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no commands, got %v", fr.calls)
	}
}

func TestInstall_AptAsRoot(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy lsd"] = "lsd:\n  Installed: (none)\n  Candidate: 0.23.1-1\n"
	if err := i.Install(Spec{Brew: "lsd", Apt: "lsd", Bin: "lsd"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "apt-get update|apt-cache policy lsd|apt-get install -y lsd"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInstall_AptUsesSudoWhenNotRoot(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy zsh"] = "  Candidate: 5.9-4\n"
	if err := i.Install(Spec{Apt: "zsh", Bin: "zsh"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.calls[0] != "sudo DEBIAN_FRONTEND=noninteractive apt-get update" {
		t.Errorf("got %q", fr.calls[0])
	}
	if fr.calls[2] != "sudo DEBIAN_FRONTEND=noninteractive apt-get install -y zsh" {
		t.Errorf("got %q", fr.calls[2])
	}
}

func TestInstall_AptUpdateRunsOnce(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy zsh"] = "  Candidate: 5.9-4\n"
	fr.outputs["apt-cache policy git"] = "  Candidate: 2.39\n"
	for _, s := range []Spec{{Apt: "zsh"}, {Apt: "git"}} {
		if err := i.Install(s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	count := 0
	for _, c := range fr.calls {
		if c == "apt-get update" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 apt-get update, got %d (%v)", count, fr.calls)
	}
}

func TestInstall_AptBinSymlink(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.paths["batcat"] = "/usr/bin/batcat"
	fr.outputs["apt-cache policy bat"] = "  Candidate: 0.24.0\n"
	if err := i.Install(Spec{Apt: "bat", AptBin: "batcat", Bin: "bat"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	target, err := os.Readlink(filepath.Join(i.BinDir(), "bat"))
	if err != nil {
		t.Fatalf("expected symlink: %v", err)
	}
	if target != "/usr/bin/batcat" {
		t.Errorf("got %q", target)
	}
}

func TestInstall_LinuxBrewOnlyErrors(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	if err := i.Install(Spec{Brew: "aerospace", BrewCask: true}); err == nil {
		t.Fatal("expected error")
	}
}

func TestInstall_AptNoCandidateNoReleaseErrors(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy hx"] = "  Candidate: (none)\n"
	if err := i.Install(Spec{Apt: "hx", Bin: "hx"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCandidate(t *testing.T) {
	cases := map[string]bool{
		"hx:\n  Installed: (none)\n  Candidate: 25.01-1\n": true,
		"hx:\n  Installed: (none)\n  Candidate: (none)\n":  false,
		"": false,
	}
	for in, want := range cases {
		if got := parseCandidate(in); got != want {
			t.Errorf("parseCandidate(%q) = %v, want %v", in, got, want)
		}
	}
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run 'TestInstall_|TestParseCandidate'`
Expected: build failure — `undefined: Installer`.

**Step 3: Implement** — `pkg/installer.go`

```go
package pkg

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type runFunc func(env []string, name string, args ...string) ([]byte, error)

type Installer struct {
	GOOS     string
	GOARCH   string
	Home     string
	IsRoot   bool
	APIBase  string
	Token    string
	HTTP     *http.Client
	lookPath func(string) (string, error)
	run      runFunc
	aptReady bool
}

func NewInstaller() (*Installer, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}
	return &Installer{
		GOOS:     runtime.GOOS,
		GOARCH:   runtime.GOARCH,
		Home:     home,
		IsRoot:   os.Geteuid() == 0,
		APIBase:  "https://api.github.com",
		Token:    os.Getenv("GITHUB_TOKEN"),
		HTTP:     http.DefaultClient,
		lookPath: exec.LookPath,
		run:      runCommand,
	}, nil
}

var defaultInstaller = sync.OnceValues(NewInstaller)

func Install(s Spec) error {
	i, err := defaultInstaller()
	if err != nil {
		return err
	}
	return i.Install(s)
}

func runCommand(env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return out, nil
}

func (i *Installer) BinDir() string {
	return filepath.Join(i.Home, ".local", "bin")
}

func (i *Installer) installed(bin string) bool {
	if bin == "" {
		return false
	}
	if _, err := i.lookPath(bin); err == nil {
		return true
	}
	_, err := os.Stat(filepath.Join(i.BinDir(), bin))
	return err == nil
}

func (i *Installer) hasApt() bool {
	_, err := i.lookPath("apt-get")
	return err == nil
}

func (i *Installer) Install(s Spec) error {
	if i.installed(s.Bin) {
		slog.Debug("already installed, skipping", "bin", s.Bin)
		return nil
	}

	switch i.GOOS {
	case "darwin":
		if s.Brew == "" {
			return nil
		}
		return i.brewInstall(s.Brew, s.BrewCask)
	case "linux":
		if s.Apt != "" && i.hasApt() {
			ok, err := i.aptHasCandidate(s.Apt)
			if err != nil {
				return err
			}
			if ok {
				return i.aptInstall(s)
			}
		}
		if s.Release != nil {
			return i.releaseInstall(s)
		}
		return fmt.Errorf("no linux install method available for %s", s.name())
	default:
		return fmt.Errorf("unsupported OS %s for %s", i.GOOS, s.name())
	}
}

func (i *Installer) brewInstall(name string, cask bool) error {
	args := []string{"list", name}
	if cask {
		args = append(args, "--cask")
	}
	if _, err := i.run(nil, "brew", args...); err == nil {
		slog.Debug("brew package already installed, skipping", "package", name)
		return nil
	}

	args[0] = "install"
	if _, err := i.run(nil, "brew", args...); err != nil {
		return fmt.Errorf("brew install %s failed: %w", name, err)
	}
	return nil
}

func (i *Installer) privileged(name string, args ...string) ([]byte, error) {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	if i.IsRoot {
		return i.run(env, name, args...)
	}
	return i.run(env, "sudo", append([]string{"DEBIAN_FRONTEND=noninteractive", name}, args...)...)
}

func (i *Installer) aptUpdate() error {
	if i.aptReady {
		return nil
	}
	if _, err := i.privileged("apt-get", "update"); err != nil {
		return fmt.Errorf("apt-get update failed: %w", err)
	}
	i.aptReady = true
	return nil
}

func (i *Installer) aptHasCandidate(name string) (bool, error) {
	if err := i.aptUpdate(); err != nil {
		return false, err
	}
	out, err := i.run(nil, "apt-cache", "policy", name)
	if err != nil {
		return false, fmt.Errorf("apt-cache policy %s failed: %w", name, err)
	}
	return parseCandidate(string(out)), nil
}

func parseCandidate(out string) bool {
	for _, line := range strings.Split(out, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "Candidate:"); ok {
			v = strings.TrimSpace(v)
			return v != "" && v != "(none)"
		}
	}
	return false
}

func (i *Installer) aptInstall(s Spec) error {
	if _, err := i.privileged("apt-get", "install", "-y", s.Apt); err != nil {
		return fmt.Errorf("apt-get install %s failed: %w", s.Apt, err)
	}
	if s.AptBin != "" && s.Bin != "" && s.AptBin != s.Bin {
		return i.linkBin(s.AptBin, s.Bin)
	}
	return nil
}

func (i *Installer) linkBin(from, to string) error {
	target, err := i.lookPath(from)
	if err != nil {
		return fmt.Errorf("%s not found after install: %w", from, err)
	}
	if err := os.MkdirAll(i.BinDir(), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", i.BinDir(), err)
	}
	link := filepath.Join(i.BinDir(), to)
	if err := os.Remove(link); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing %s: %w", link, err)
	}
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("failed to link %s to %s: %w", link, target, err)
	}
	return nil
}
```

`releaseInstall` does not exist yet; add a temporary stub at the bottom of `installer.go` so the package compiles (Task 6 replaces it):

```go
func (i *Installer) releaseInstall(s Spec) error {
	return fmt.Errorf("release install not implemented for %s", s.name())
}
```

**Step 4: Run tests**

Run: `task test --force -- -run 'TestInstall_|TestParseCandidate|TestMatchAsset'`
Expected: PASS. (The root package's `TestInstall_ContinuesAndJoinsErrors` also matches and should still pass.)

**Step 5: Commit**

```bash
git add pkg/
git commit -m "feat(pkg): installer with brew and apt support"
```

---

### Task 6: `pkg` — GitHub release installs

**Files:**
- Create: `pkg/release.go`
- Create: `pkg/release_test.go`
- Modify: `pkg/installer.go` (delete the `releaseInstall` stub)

**Step 1: Write the failing tests** — `pkg/release_test.go`

```go
package pkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func releaseServer(t *testing.T, assets map[string][]byte) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			list := []map[string]string{}
			for name := range assets {
				list = append(list, map[string]string{"name": name, "browser_download_url": srv.URL + "/dl/" + name})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v1.0.0", "assets": list})
			return
		}
		body, ok := assets[strings.TrimPrefix(r.URL.Path, "/dl/")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestReleaseInstall_Archive(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"lsd-v1.2.0-x86_64-unknown-linux-gnu.tar.gz": makeTarGz(t, map[string]string{
			"lsd-v1.2.0-x86_64-unknown-linux-gnu/lsd":       "LSD",
			"lsd-v1.2.0-x86_64-unknown-linux-gnu/README.md": "readme",
		}),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "lsd", Release: &Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFile(t, filepath.Join(i.BinDir(), "lsd"), "LSD")
	info, _ := os.Stat(filepath.Join(i.BinDir(), "lsd"))
	if info.Mode().Perm()&0o100 == 0 {
		t.Errorf("binary not executable: %v", info.Mode())
	}
}

func TestReleaseInstall_RawBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"posh-linux-amd64": []byte("POSH"),
		"posh-linux-arm64": []byte("WRONG"),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "oh-my-posh", Release: &Release{
		Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`,
		Arch: map[string]string{"amd64": "amd64", "arm64": "arm64"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "oh-my-posh"), "POSH")
}

func TestReleaseInstall_CopiesDirs(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := releaseServer(t, map[string][]byte{
		"helix-25.07.1-x86_64-linux.tar.gz": makeTarGz(t, map[string]string{
			"helix-25.07.1-x86_64-linux/hx":                          "HX",
			"helix-25.07.1-x86_64-linux/runtime/themes/x.toml":       "theme",
			"helix-25.07.1-x86_64-linux/runtime/queries/go/hl.scm":   "query",
		}),
	})
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "hx", Release: &Release{
		Repo:  "helix-editor/helix",
		Asset: `helix-.*-{arch}-linux\.tar\.gz`,
		Dirs:  map[string]string{"runtime": "~/.config/helix/runtime"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "hx"), "HX")
	assertFile(t, filepath.Join(i.Home, ".config/helix/runtime/queries/go/hl.scm"), "query")
}

func TestReleaseInstall_APIError(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer srv.Close()
	i.APIBase = srv.URL

	err := i.Install(Spec{Bin: "lsd", Release: &Release{Repo: "lsd-rs/lsd", Asset: `lsd`}})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 error, got %v", err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Errorf("%s = %q, want %q", path, got, want)
	}
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestReleaseInstall`
Expected: FAIL — `release install not implemented`.

**Step 3: Implement** — delete the stub in `installer.go`, then create `pkg/release.go`:

```go
package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func (r *ghRelease) asset(name string) (ghAsset, bool) {
	for _, a := range r.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return ghAsset{}, false
}

func (i *Installer) latestRelease(repo string) (*ghRelease, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/releases/latest", i.APIBase, repo), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build release request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if i.Token != "" {
		req.Header.Set("Authorization", "Bearer "+i.Token)
	}

	resp, err := i.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release for %s: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release for %s: status %d", repo, resp.StatusCode)
	}

	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to decode release for %s: %w", repo, err)
	}
	return &rel, nil
}

func (i *Installer) download(url, dest string) error {
	resp, err := i.HTTP.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", url, resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	return nil
}

func (i *Installer) releaseInstall(s Spec) error {
	r := s.Release
	if s.Bin == "" {
		return fmt.Errorf("release install for %s requires Bin", r.Repo)
	}

	rel, err := i.latestRelease(r.Repo)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(rel.Assets))
	for _, a := range rel.Assets {
		names = append(names, a.Name)
	}
	name, err := r.matchAsset(i.GOARCH, names)
	if err != nil {
		return err
	}
	asset, _ := rel.asset(name)

	tmp, err := os.MkdirTemp("", "shizuku-pkg-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	downloaded := filepath.Join(tmp, name)
	if err := i.download(asset.URL, downloaded); err != nil {
		return err
	}

	binPath := downloaded
	root := tmp
	if isArchive(name) {
		root = filepath.Join(tmp, "extract")
		if err := extract(downloaded, root); err != nil {
			return err
		}
		binPath, err = findInTree(root, s.Bin, false)
		if err != nil {
			return err
		}
	}

	if err := installFile(binPath, filepath.Join(i.BinDir(), s.Bin)); err != nil {
		return err
	}

	for src, dst := range r.Dirs {
		dir, err := findInTree(root, src, true)
		if err != nil {
			return err
		}
		if err := replaceDir(dir, i.expandHome(dst)); err != nil {
			return err
		}
	}

	slog.Info("installed from github release", "repo", r.Repo, "tag", rel.TagName, "bin", s.Bin)
	return nil
}

func (i *Installer) expandHome(p string) string {
	if rest, ok := strings.CutPrefix(p, "~/"); ok {
		return filepath.Join(i.Home, rest)
	}
	return p
}

func isArchive(name string) bool {
	for _, ext := range []string{".tar.gz", ".tgz", ".tar.xz"} {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func extract(archive, dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	out, err := exec.Command("tar", "-xf", archive, "-C", dest).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to extract %s: %w: %s", archive, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func findInTree(root, name string, wantDir bool) (string, error) {
	var found string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Name() != name {
			return nil
		}
		if wantDir && d.IsDir() || !wantDir && d.Type().IsRegular() {
			found = p
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to search %s: %w", root, err)
	}
	if found == "" {
		return "", fmt.Errorf("%s not found in release archive", name)
	}
	return found, nil
}

func installFile(src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(dest), err)
	}

	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", src, err)
	}
	defer in.Close()

	tmp := dest + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", tmp, err)
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return fmt.Errorf("failed to write %s: %w", tmp, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		return fmt.Errorf("failed to move %s into place: %w", dest, err)
	}
	return nil
}

func replaceDir(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove %s: %w", dest, err)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("failed to create %s: %w", filepath.Dir(dest), err)
	}
	if err := os.CopyFS(dest, os.DirFS(src)); err != nil {
		return fmt.Errorf("failed to copy %s to %s: %w", src, dest, err)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `task test --force -- -run 'TestReleaseInstall|TestInstall_|TestMatchAsset'`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/
git commit -m "feat(pkg): install binaries from github releases"
```

---

### Task 7: Route `util.InstallBrewPackage` through `pkg`

**Files:**
- Modify: `util/homebrew.go`

**Step 1: Edit** — replace `InstallBrewPackage` and delete `BrewPackageExists` (no other callers; verify with `grep -rn BrewPackageExists --include='*.go' .`):

```go
func InstallBrewPackage(name string, isCask bool) error {
	return pkg.Install(pkg.Spec{Brew: name, BrewCask: isCask})
}
```

Add import `"github.com/eleonorayaya/shizuku/pkg"`; drop `"log/slog"` if now unused. Keep `GetBrewAppPrefix`, `AddTap`, `runBrewCommand` untouched (`pkg` must not import `util`, so no cycle).

**Step 2: Verify**

Run: `task build && task test --force`
Expected: build succeeds; all tests PASS.

**Step 3: Commit**

```bash
git add util/homebrew.go
git commit -m "refactor: route brew installs through pkg"
```

---

### Task 8: Guarded aliases and env vars (`Requires`)

**Files:**
- Modify: `app/env.go`
- Create: `app/env_test.go`

**Step 1: Write the failing test** — `app/env_test.go`

```go
package app

import (
	"os"
	"strings"
	"testing"
)

func generateEnv(t *testing.T, setups ...*EnvSetup) (zshenv, sh string) {
	t.Helper()
	files, err := GenerateEnvFiles(setups, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a, err := os.ReadFile(files["shizuku.zshenv"])
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(files["shizuku.sh"])
	if err != nil {
		t.Fatal(err)
	}
	return string(a), string(b)
}

func TestGenerateEnv_GuardedAlias(t *testing.T) {
	_, sh := generateEnv(t, &EnvSetup{Aliases: []Alias{
		{Name: "ls", Command: "lsd", Requires: "lsd"},
		{Name: "c", Command: "clear"},
	}})
	if !strings.Contains(sh, "command -v lsd >/dev/null 2>&1 && alias ls='lsd'\n") {
		t.Errorf("missing guarded alias:\n%s", sh)
	}
	if !strings.Contains(sh, "\nalias c='clear'\n") {
		t.Errorf("missing plain alias:\n%s", sh)
	}
}

func TestGenerateEnv_GuardedVarAfterPath(t *testing.T) {
	zshenv, _ := generateEnv(t, &EnvSetup{
		Variables: []EnvVar{
			{Key: "EDITOR", Value: "hx", Requires: "hx"},
			{Key: "GOPATH", Value: "/go"},
		},
		PathDirs: []PathDir{{Path: "$HOME/.local/bin", Priority: 5}},
	})
	guarded := strings.Index(zshenv, `command -v hx >/dev/null 2>&1 && export EDITOR="hx"`)
	path := strings.Index(zshenv, `export PATH=`)
	plain := strings.Index(zshenv, `export GOPATH="/go"`)
	if guarded < 0 || path < 0 || plain < 0 {
		t.Fatalf("missing lines:\n%s", zshenv)
	}
	if !(plain < path && path < guarded) {
		t.Errorf("want plain vars < PATH < guarded vars, got %d %d %d:\n%s", plain, path, guarded, zshenv)
	}
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestGenerateEnv`
Expected: build failure — `unknown field Requires`.

**Step 3: Implement** in `app/env.go`:

1. Add `Requires string` to both `EnvVar` and `Alias`.
2. Change the accumulators: `vars := make(map[string]EnvVar)` (store `vars[v.Key] = v`) and `aliases := make(map[string]Alias)` (store `aliases[a.Name] = a`).
3. Replace the zshenv variables + PATH sections with:

```go
	sortedVarKeys := make([]string, 0, len(vars))
	for k := range vars {
		sortedVarKeys = append(sortedVarKeys, k)
	}
	sort.Strings(sortedVarKeys)

	plainVars := []EnvVar{}
	guardedVars := []EnvVar{}
	for _, k := range sortedVarKeys {
		if vars[k].Requires == "" {
			plainVars = append(plainVars, vars[k])
		} else {
			guardedVars = append(guardedVars, vars[k])
		}
	}

	if len(plainVars) > 0 {
		writeSectionHeader(&zshenv, "Environment Variables")
		for _, v := range plainVars {
			zshenv.WriteString(fmt.Sprintf("export %s=\"%s\"\n", v.Key, v.Value))
		}
		zshenv.WriteString("\n")
	}

	if len(pathDirs) > 0 {
		writeSectionHeader(&zshenv, "PATH Configuration")
		pathParts := []string{}
		for _, pd := range pathDirs {
			pathParts = append(pathParts, pd.Path)
		}
		pathParts = append(pathParts, "$PATH")
		zshenv.WriteString(fmt.Sprintf("export PATH=\"%s\"\n", strings.Join(pathParts, ":")))
		zshenv.WriteString("\n")
	}

	if len(guardedVars) > 0 {
		writeSectionHeader(&zshenv, "Conditional Environment Variables")
		for _, v := range guardedVars {
			zshenv.WriteString(fmt.Sprintf("command -v %s >/dev/null 2>&1 && export %s=\"%s\"\n", v.Requires, v.Key, v.Value))
		}
		zshenv.WriteString("\n")
	}
```

4. In the aliases section write each entry as:

```go
		for _, k := range sortedAliasKeys {
			a := aliases[k]
			if a.Requires != "" {
				sh.WriteString(fmt.Sprintf("command -v %s >/dev/null 2>&1 && alias %s='%s'\n", a.Requires, a.Name, a.Command))
			} else {
				sh.WriteString(fmt.Sprintf("alias %s='%s'\n", a.Name, a.Command))
			}
		}
```

5. Add the helper and use it for the existing section headers too (they're all the same three lines):

```go
func writeSectionHeader(b *strings.Builder, title string) {
	b.WriteString("# ============================================================\n")
	b.WriteString("# " + title + "\n")
	b.WriteString("# ============================================================\n")
}
```

**Step 4: Run tests**

Run: `task test --force`
Expected: PASS (including existing `agents/claude` and `config` tests).

**Step 5: Commit**

```bash
git add app/
git commit -m "feat(env): guard aliases and env vars on binary presence"
```

---

### Task 9: Convert bat, lsd, helix, tuios to `pkg`

**Files:**
- Modify: `programs/bat/bat.go`
- Modify: `programs/lsd/lsd.go`
- Modify: `programs/helix/helix.go`
- Modify: `programs/tuios/tuios.go`

No new unit tests: the specs are data exercised by `pkg` tests (asset regexes are covered verbatim in Task 4); behavior is verified end-to-end in Task 14.

**Step 1: bat** — replace imports of `util` with `pkg`, and:

```go
var spec = pkg.Spec{
	Brew:   "bat",
	Apt:    "bat",
	AptBin: "batcat",
	Bin:    "bat",
	Release: &pkg.Release{
		Repo:  "sharkdp/bat",
		Asset: `bat-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`,
	},
}

func (a *App) Install(ctx *app.Context) error {
	if err := pkg.Install(spec); err != nil {
		return fmt.Errorf("failed to install bat: %w", err)
	}
	return nil
}
```
Alias becomes `{Name: "cat", Command: "bat", Requires: "bat"}`.

**Step 2: lsd** — same shape:

```go
var spec = pkg.Spec{
	Brew: "lsd",
	Apt:  "lsd",
	Bin:  "lsd",
	Release: &pkg.Release{
		Repo:  "lsd-rs/lsd",
		Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`,
	},
}
```
Alias becomes `{Name: "ls", Command: "lsd", Requires: "lsd"}`.

**Step 3: helix**

```go
var spec = pkg.Spec{
	Brew: "helix",
	Apt:  "hx",
	Bin:  "hx",
	Release: &pkg.Release{
		Repo:  "helix-editor/helix",
		Asset: `helix-.*-{arch}-linux\.tar\.xz`,
		Dirs:  map[string]string{"runtime": "~/.config/helix/runtime"},
	},
}
```
Env vars become `{Key: "EDITOR", Value: "hx", Requires: "hx"}` and `{Key: "VISUAL", Value: "hx", Requires: "hx"}`.

**Step 4: tuios**

```go
var spec = pkg.Spec{
	Brew: "tuios",
	Bin:  "tuios",
	Release: &pkg.Release{
		Repo:  "Gaurav-Gosain/tuios",
		Asset: `tuios_[0-9.]+_Linux_{arch}\.tar\.gz`,
		Arch:  map[string]string{"amd64": "x86_64", "arm64": "arm64"},
	},
}
```
In `Generate`, replace the hard-coded macOS dir:

```go
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user config dir: %w", err)
	}

	return &app.GenerateResult{
		FileMap: fileMap,
		DestDir: filepath.Join(configDir, "tuios"),
	}, nil
```
(`os.UserConfigDir` = `~/Library/Application Support` on darwin, `$XDG_CONFIG_HOME` or `~/.config` on Linux — identical to tuios's `adrg/xdg` lookup. `app.SyncAppFile` uses `path.Join`, so no trailing slash needed.)

**Step 5: Verify**

Run: `task lint && task build && task test --force`
Expected: all succeed.

Run: `task run -- diff --profile personal`
Expected: no changes reported for `tuios` (same destination on macOS).

**Step 6: Commit**

```bash
git add programs/bat programs/lsd programs/helix programs/tuios
git commit -m "feat: install bat, lsd, helix, tuios via apt or github releases on linux"
```

---

### Task 10: git — install git, tolerate missing gh

**Files:**
- Modify: `programs/git/git.go` (`Install`, ~L42)

**Step 1: Edit `Install`**

```go
func (a *App) Install(ctx *app.Context) error {
	if err := pkg.Install(pkg.Spec{Brew: "git", Apt: "git", Bin: "git"}); err != nil {
		return fmt.Errorf("failed to install git: %w", err)
	}

	if !util.BinaryExists("gh") {
		slog.Info("gh not found in PATH, skipping gh extensions")
		return nil
	}

	... (existing extension + skill loop unchanged)
}
```
Add imports `"log/slog"` and `"github.com/eleonorayaya/shizuku/pkg"`.

**Step 2: Verify**

Run: `task build && task test --force`
Expected: success.

**Step 3: Commit**

```bash
git add programs/git
git commit -m "feat(git): install git via pkg and skip gh extensions when gh is absent"
```

---

### Task 11: terminal — drop antigen, portable zsh setup, rc wiring

**Files:**
- Modify: `programs/terminal/terminal.go`
- Create: `programs/terminal/terminal_test.go`

**Step 1: Write the failing test** — `programs/terminal/terminal_test.go`

```go
package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureLine_CreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "source ~/.config/shizuku/shizuku.sh\n" {
		t.Errorf("got %q", got)
	}
}

func TestEnsureLine_AppendsWithNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	if err := os.WriteFile(path, []byte("export FOO=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "export FOO=1\nsource ~/.config/shizuku/shizuku.sh\n" {
		t.Errorf("got %q", got)
	}
}

func TestEnsureLine_RespectsExistingVariant(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")
	existing := `source "$HOME/.config/shizuku/shizuku.sh"` + "\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ensureLine(path, "source ~/.config/shizuku/shizuku.sh", "shizuku/shizuku.sh"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != existing {
		t.Errorf("file changed: %q", got)
	}
}

func TestEnsureLine_Idempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshenv")
	for range 3 {
		if err := ensureLine(path, "source ~/.config/shizuku/shizuku.zshenv", "shizuku/shizuku.zshenv"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	got, _ := os.ReadFile(path)
	if strings.Count(string(got), "shizuku.zshenv") != 1 {
		t.Errorf("got %q", got)
	}
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run TestEnsureLine`
Expected: build failure — `undefined: ensureLine`.

**Step 3: Implement** — rewrite the top of `programs/terminal/terminal.go` (keep `Name`, `Generate`, `Sync`, `colormapFunction` as-is):

```go
package terminal

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/pkg"
	"github.com/eleonorayaya/shizuku/util"
)

//go:embed all:contents
var contents embed.FS

const zshViModeRepo = "https://github.com/jeffreytse/zsh-vi-mode"

const zshViModePath = "~/.local/share/shizuku/plugins/zsh-vi-mode"

const zshViModeInit = `if [[ -f ~/.local/share/shizuku/plugins/zsh-vi-mode/zsh-vi-mode.plugin.zsh ]]; then
  source ~/.local/share/shizuku/plugins/zsh-vi-mode/zsh-vi-mode.plugin.zsh
fi`

const ohmyposhInit = `if command -v oh-my-posh >/dev/null 2>&1; then
  eval "$(oh-my-posh init zsh --config ~/.config/ohmyposh/ohmyposh.json)"
fi`

type rcLine struct {
	file   string
	line   string
	marker string
}

var rcLines = []rcLine{
	{file: "~/.zshenv", line: "source ~/.config/shizuku/shizuku.zshenv", marker: "shizuku/shizuku.zshenv"},
	{file: "~/.zshrc", line: "source ~/.config/shizuku/shizuku.sh", marker: "shizuku/shizuku.sh"},
}

var packages = []pkg.Spec{
	{Apt: "zsh", Bin: "zsh"},
	{Apt: "ca-certificates"},
	{Apt: "xz-utils", Bin: "xz"},
	{
		Brew: "jandedobbeleer/oh-my-posh/oh-my-posh",
		Bin:  "oh-my-posh",
		Release: &pkg.Release{
			Repo:  "JanDeDobbeleer/oh-my-posh",
			Asset: `posh-linux-{arch}`,
			Arch:  map[string]string{"amd64": "amd64", "arm64": "arm64"},
		},
	},
}

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "terminal"
}

func (a *App) Install(ctx *app.Context) error {
	for _, spec := range packages {
		if err := pkg.Install(spec); err != nil {
			return fmt.Errorf("failed to install terminal package: %w", err)
		}
	}

	if err := ensureZshViMode(); err != nil {
		return fmt.Errorf("failed to install zsh-vi-mode: %w", err)
	}

	for _, rc := range rcLines {
		path, err := util.NormalizeFilePath(rc.file)
		if err != nil {
			return fmt.Errorf("failed to resolve %s: %w", rc.file, err)
		}
		if err := ensureLine(path, rc.line, rc.marker); err != nil {
			return fmt.Errorf("failed to update %s: %w", rc.file, err)
		}
	}

	if shell := os.Getenv("SHELL"); !strings.HasSuffix(shell, "zsh") {
		slog.Warn("login shell is not zsh; switch with: chsh -s \"$(command -v zsh)\"", "shell", shell)
	}

	return nil
}

func ensureZshViMode() error {
	path, err := util.NormalizeFilePath(zshViModePath)
	if err != nil {
		return fmt.Errorf("failed to resolve zsh-vi-mode path: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		slog.Debug("zsh-vi-mode already installed, skipping")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("failed to create plugin dir: %w", err)
	}

	output, err := exec.Command("git", "clone", "--depth", "1", zshViModeRepo, path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone zsh-vi-mode: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func ensureLine(path, line, marker string) error {
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	if bytes.Contains(data, []byte(marker)) {
		return nil
	}

	prefix := ""
	if len(data) > 0 && !bytes.HasSuffix(data, []byte("\n")) {
		prefix = "\n"
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", path, err)
	}
	defer f.Close()

	if _, err := f.WriteString(prefix + line + "\n"); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}
```

In `Env()`, change `InitScripts: []string{antigenInit, ohmyposhInit}` to `InitScripts: []string{zshViModeInit, ohmyposhInit}`. Remove the `antigenInit` const.

**Step 4: Run tests**

Run: `task test --force -- -run TestEnsureLine && task build`
Expected: PASS, build OK.

**Step 5: Commit**

```bash
git add programs/terminal
git commit -m "feat(terminal): replace antigen with portable zsh-vi-mode, wire rc files, install zsh on linux"
```

---

### Task 12: Restructure `examples/eleonora/main.go` profiles

**Files:**
- Modify: `examples/eleonora/main.go` (`main`, ~L72-140)

**Step 1: Edit** — the builder call becomes:

```go
	cmd := shizuku.New(
		shizuku.WithStyles(styles.New(
			styles.WithTheme(themes.MonadeDark),
			styles.WithWindowOpacity(65),
			styles.WithGaps(desktopGaps),
			styles.WithGapOverride("built-in retina", laptopGaps),
		)),
		shizuku.WithPrograms(
			git.New(),
			terminal.New(),
			helix.New(),
			bat.New(),
			lsd.New(),
			tuios.New(),
		),
		shizuku.WithProfile("desktop",
			shizuku.WithLanguages(
				golang.New(),
				lua.New(),
				python.New(),
				rust.New(),
				typescript.New(),
			),
			shizuku.WithPrograms(
				aerospace.New(),
				desktoppr.New(),
				fastfetch.New(),
				glow.New(),
				jankyborders.New(),
				kitty.New(),
				notion.New(notion.Options{DisableClaudeMCP: true}),
				nvim.New(),
				rtk.New(),
				sfsymbols.New(),
				sketchybar.New(),
				terminalbrowser.New(),
				terraform.New(),
				tmux.New(),
				utena.New(),
			),
			shizuku.WithAgents(
				claude.New(data.ClaudeOptions()),
			),
		),
		shizuku.WithProfile("work",
			shizuku.Extends("desktop"),
			shizuku.WithLanguages(
				ruby.New(),
			),
			shizuku.WithPrograms(
				acli.New(),
				buildkite.New(),
				datadog.New(),
				k9s.New(),
				mise.New(),
			),
		),
		shizuku.WithProfile("personal",
			shizuku.Extends("desktop"),
			shizuku.WithLanguages(
				swift.New(),
				zig.New(),
			),
			shizuku.WithPrograms(
				protonpass.New(),
				protonvpn.New(),
				notion.New(notion.Options{DisableClaudeMCP: false}),
			),
		),
	).Command()
```

Check the base program set covers exactly: git, terminal, helix, bat, lsd, tuios — and every previously-registered program appears once across base + desktop.

**Step 2: Verify**

Run: `task build`
Expected: success.

Run: `task run -- list --profile personal`
Expected: git, terminal, helix, bat, lsd, tuios first; then all desktop programs; personal additions (protonpass, protonvpn); languages incl. swift/zig; agent claude.

Run: `task run -- list --profile desktop`
Expected: same, without the personal additions. (Base-only resolution is covered by `TestActiveProfile_BaseWhenUnset`.)

**Step 3: Commit**

```bash
git add examples/eleonora/main.go
git commit -m "feat: make base profile minimal and move desktop apps into desktop profile"
```

---

### Task 13: Self-update from releases

**Files:**
- Create: `pkg/selfupdate.go`
- Create: `pkg/selfupdate_test.go`
- Modify: `examples/eleonora/main.go` (`upgradeCmd`, ~L150)

**Step 1: Write the failing tests** — `pkg/selfupdate_test.go`

```go
package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestVerifyChecksum(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "shizuku_linux_amd64.tar.gz")
	if err := os.WriteFile(file, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	sums := filepath.Join(dir, "checksums.txt")

	write := func(content string) {
		if err := os.WriteFile(sums, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write(sha([]byte("other")) + "  shizuku_darwin_arm64.tar.gz\n" + sha([]byte("payload")) + "  shizuku_linux_amd64.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err != nil {
		t.Errorf("expected match: %v", err)
	}

	write(sha([]byte("tampered")) + "  shizuku_linux_amd64.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err == nil {
		t.Error("expected mismatch error")
	}

	write(sha([]byte("payload")) + "  something_else.tar.gz\n")
	if err := verifyChecksum(sums, "shizuku_linux_amd64.tar.gz", file); err == nil {
		t.Error("expected missing-entry error")
	}
}

func TestUpdateBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	archive := makeTarGz(t, map[string]string{"shizuku": "NEW", "README.md": "x"})
	srv := releaseServer(t, map[string][]byte{
		"shizuku_linux_amd64.tar.gz": archive,
		"checksums.txt":              []byte(sha(archive) + "  shizuku_linux_amd64.tar.gz\n"),
	})
	i.APIBase = srv.URL

	dest := filepath.Join(t.TempDir(), "shizuku")
	if err := os.WriteFile(dest, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := i.updateBinary("eleonorayaya/dotfiles", "shizuku", dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, dest, "NEW")
}

func TestUpdateBinary_BadChecksumLeavesBinary(t *testing.T) {
	i, _ := testInstaller(t, "linux")
	archive := makeTarGz(t, map[string]string{"shizuku": "NEW"})
	srv := releaseServer(t, map[string][]byte{
		"shizuku_linux_amd64.tar.gz": archive,
		"checksums.txt":              []byte(sha([]byte("nope")) + "  shizuku_linux_amd64.tar.gz\n"),
	})
	i.APIBase = srv.URL

	dest := filepath.Join(t.TempDir(), "shizuku")
	if err := os.WriteFile(dest, []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := i.updateBinary("eleonorayaya/dotfiles", "shizuku", dest); err == nil {
		t.Fatal("expected checksum error")
	}
	assertFile(t, dest, "OLD")
}
```

**Step 2: Run to verify failure**

Run: `task test --force -- -run 'TestVerifyChecksum|TestUpdateBinary'`
Expected: build failure — `undefined: verifyChecksum`.

**Step 3: Implement** — `pkg/selfupdate.go`

```go
package pkg

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func SelfUpdate(repo, project string) error {
	i, err := defaultInstaller()
	if err != nil {
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate running binary: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("failed to resolve running binary: %w", err)
	}

	return i.updateBinary(repo, project, exe)
}

func (i *Installer) updateBinary(repo, project, dest string) error {
	rel, err := i.latestRelease(repo)
	if err != nil {
		return err
	}

	archiveName := fmt.Sprintf("%s_%s_%s.tar.gz", project, i.GOOS, i.GOARCH)
	archive, ok := rel.asset(archiveName)
	if !ok {
		return fmt.Errorf("release %s has no asset %s", rel.TagName, archiveName)
	}
	sums, ok := rel.asset("checksums.txt")
	if !ok {
		return fmt.Errorf("release %s has no checksums.txt", rel.TagName)
	}

	tmp, err := os.MkdirTemp("", "shizuku-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	archivePath := filepath.Join(tmp, archiveName)
	sumsPath := filepath.Join(tmp, "checksums.txt")
	if err := i.download(archive.URL, archivePath); err != nil {
		return err
	}
	if err := i.download(sums.URL, sumsPath); err != nil {
		return err
	}
	if err := verifyChecksum(sumsPath, archiveName, archivePath); err != nil {
		return err
	}

	root := filepath.Join(tmp, "extract")
	if err := extract(archivePath, root); err != nil {
		return err
	}
	bin, err := findInTree(root, project, false)
	if err != nil {
		return err
	}
	if err := installFile(bin, dest); err != nil {
		return err
	}

	slog.Info("updated from github release", "tag", rel.TagName, "path", dest)
	return nil
}

func verifyChecksum(sumsPath, name, filePath string) error {
	data, err := os.ReadFile(sumsPath)
	if err != nil {
		return fmt.Errorf("failed to read checksums: %w", err)
	}

	want := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == name {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no checksum listed for %s", name)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("failed to hash %s: %w", filePath, err)
	}

	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		return fmt.Errorf("checksum mismatch for %s: got %s, want %s", name, got, want)
	}
	return nil
}
```

**Step 4: Run tests**

Run: `task test --force -- -run 'TestVerifyChecksum|TestUpdateBinary'`
Expected: PASS.

**Step 5: Wire into `upgradeCmd`** in `examples/eleonora/main.go`:

Add `const releaseRepo = "eleonorayaya/dotfiles"` next to `sourceDir`, a `var fromRelease bool` beside `branch`, and at the top of `RunE`:

```go
			repoDir, err := util.NormalizeFilePath(sourceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve source directory: %w", err)
			}

			if _, err := os.Stat(repoDir); fromRelease || os.IsNotExist(err) {
				slog.Info("upgrading from latest github release", "repo", releaseRepo)
				return pkg.SelfUpdate(releaseRepo, "shizuku")
			}
```
Delete the old `os.IsNotExist` error branch (it's now the release path). Register the flag:

```go
	cmd.Flags().BoolVar(&fromRelease, "from-release", false, "Download the latest release binary instead of building from source")
```
Import `"github.com/eleonorayaya/shizuku/pkg"`.

(The design doc listed a unit test for mode selection; it's a single `if` in a `main` package with no seam, so it's verified manually in Task 14 instead.)

**Step 6: Verify**

Run: `task build && task test --force && task run -- upgrade --help`
Expected: help lists `--from-release`.

**Step 7: Commit**

```bash
git add pkg/selfupdate.go pkg/selfupdate_test.go examples/eleonora/main.go
git commit -m "feat: upgrade from github releases when no source checkout exists"
```

---

### Task 14: Release pipeline

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/release.yml`

**Step 1: `.goreleaser.yaml`**

```yaml
version: 2

project_name: shizuku

builds:
  - id: shizuku
    main: ./examples/eleonora
    binary: shizuku
    env:
      - CGO_ENABLED=0
    goos: [linux, darwin]
    goarch: [amd64, arm64]
    ignore:
      - goos: darwin
        goarch: amd64

archives:
  - formats: [tar.gz]
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"

checksum:
  name_template: checksums.txt

changelog:
  use: github
```

**Step 2: `.github/workflows/release.yml`**

```yaml
name: release

on:
  push:
    branches: [main]
    paths-ignore:
      - "docs/**"
      - ".claude/**"
      - "**/*.md"

permissions:
  contents: write

concurrency:
  group: release
  cancel-in-progress: false

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Test
        run: go test ./...

      - name: Tag next patch version
        run: |
          last=$(git tag --list 'v*' --sort=-v:refname | head -n1)
          if [ -z "$last" ]; then
            next=v0.1.0
          else
            IFS=. read -r major minor patch <<< "${last#v}"
            next="v${major}.${minor}.$((patch + 1))"
          fi
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git tag "$next"
          git push origin "$next"

      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

**Step 3: Validate locally** (if goreleaser is installed; otherwise `brew install goreleaser` first — ask before installing)

Run: `goreleaser check && goreleaser release --snapshot --clean --skip=publish && ls dist/`
Expected: `checksums.txt`, `shizuku_linux_amd64.tar.gz`, `shizuku_linux_arm64.tar.gz`, `shizuku_darwin_arm64.tar.gz`.

Add `dist/` to `.gitignore` if not already ignored.

**Step 4: Commit**

```bash
git add .goreleaser.yaml .github/workflows/release.yml .gitignore
git commit -m "ci: cut a github release on every merge to main"
```

---

### Task 15: Docs

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`

**Step 1: README** — append:

````markdown
## Servers

The base profile is a minimal, server-safe set: git, zsh (vi-mode + oh-my-posh), helix, bat, lsd, tuios. On Debian/Ubuntu it installs via apt, falling back to GitHub release binaries in `~/.local/bin`.

```sh
mkdir -p ~/.local/bin
curl -fsSL https://github.com/eleonorayaya/dotfiles/releases/latest/download/shizuku_linux_$(dpkg --print-architecture).tar.gz \
  | tar -xz -C ~/.local/bin shizuku
~/.local/bin/shizuku install && ~/.local/bin/shizuku sync
```

Update later with `shizuku upgrade` (downloads the latest release when there is no source checkout).

On a desktop, set a profile: `shizuku config set profile personal` (or `work`). Both extend the `desktop` profile.
````

**Step 2: CLAUDE.md** — add after "Adding a New App":

````markdown
## Installing Packages

Use `pkg.Install(pkg.Spec{...})` in an app's `Install()`. A spec lists every source; the first that applies wins (brew on macOS; apt, then GitHub release on Linux):

```go
pkg.Spec{
    Brew: "lsd", Apt: "lsd", Bin: "lsd",
    Release: &pkg.Release{Repo: "lsd-rs/lsd", Asset: `lsd-v[0-9.]+-{arch}-unknown-linux-gnu\.tar\.gz`},
}
```

`util.InstallBrewPackage` remains for macOS-only apps. Aliases/env vars that depend on an optional binary should set `Requires` so the shell stays usable if the install failed.

## Profiles

The base profile is server-safe. `desktop` holds macOS apps, languages and agents; `work` and `personal` use `shizuku.Extends("desktop")`. Only add apps to the base if they work on Debian/Ubuntu.
````

Also fix the stale paths in CLAUDE.md: `apps/languages/` → `languages/`, `apps/programs/` → `programs/`, `apps/agents/` → `agents/`.

**Step 3: Commit**

```bash
git add README.md CLAUDE.md
git commit -m "docs: document server bootstrap, pkg installs and profile layering"
```

---

### Task 16: End-to-end verification

**Step 1: Full suite + build**

Run: `task lint && task test --force && task build && task build:linux`
Expected: all pass.

**Step 2: Linux containers** (requires a running Docker daemon, e.g. `colima start`; if unavailable, report that and skip)

Mac is arm64, so use the arm64 binary:

```bash
for image in debian:bookworm ubuntu:24.04; do
  docker run --rm -v "$PWD/out/shizuku-linux-arm64:/usr/local/bin/shizuku:ro" -w /root "$image" bash -c '
    set -e
    shizuku install
    shizuku sync
    zsh -i -c "command -v hx bat lsd tuios oh-my-posh && ls ~/.config/helix/runtime >/dev/null && echo OK"
  '
done
```
Expected per image: five paths printed, then `OK`; no errors from `zsh -i`. On Ubuntu 24.04, `bat` resolves to `~/.local/bin/bat` (symlink to `batcat`).

Re-run `shizuku install` in the same container (chain it twice in the `bash -c`) and confirm no re-downloads (`installed from github release` should not appear the second time).

**Step 3: Mac regression**

Run: `task run -- diff --profile personal`
Expected: only `shizuku (env)` changes — the antigen init replaced by the zsh-vi-mode block, oh-my-posh init wrapped in a guard, guarded `cat`/`ls` aliases and `EDITOR`/`VISUAL`. Nothing else.

Then: `task run -- install --profile personal` — expect zsh-vi-mode cloned to `~/.local/share/shizuku/plugins/zsh-vi-mode`, no duplicate lines added to an existing `~/.zshrc`/`~/.zshenv` that already sources shizuku.

**Step 4: Confirm config** — ensure `~/.config/shizuku/shizuku.yml` has `profile: personal` (or `work`); otherwise the Mac gets only the base set.

**Step 5:** Use superpowers:finishing-a-development-branch.
