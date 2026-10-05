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
	envs    []string
}

func (f *fakeRunner) run(env []string, name string, args ...string) ([]byte, error) {
	cmd := strings.Join(append([]string{name}, args...), " ")
	f.calls = append(f.calls, cmd)
	f.envs = append(f.envs, strings.Join(env, " "))
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
	if len(fr.calls) != 2 {
		t.Fatalf("expected 2 calls, got %v", fr.calls)
	}
	if got := fr.calls[1]; got != "brew install aerospace --cask" {
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
		t.Fatalf("got %q, want %q", got, want)
	}
	if !strings.Contains(fr.envs[1], "LC_ALL=C") {
		t.Errorf("apt-cache policy env = %q, want LC_ALL=C", fr.envs[1])
	}
	if !strings.Contains(fr.envs[2], "DEBIAN_FRONTEND=noninteractive") {
		t.Errorf("apt-get install env = %q, want DEBIAN_FRONTEND=noninteractive", fr.envs[2])
	}
}

func TestInstall_AptUpdateFailureFallsBackToRelease(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.fail["apt-get update"] = true
	srv := releaseServer(t, map[string][]byte{"posh-linux-amd64": []byte("POSH")})
	i.APIBase = srv.URL

	err := i.Install(Spec{Apt: "oh-my-posh", Bin: "oh-my-posh", Release: &Release{
		Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`,
		Arch: map[string]string{"amd64": "amd64"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "oh-my-posh"), "POSH")
}

func TestInstall_AptUpdateFailureAptOnlyErrors(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.fail["apt-get update"] = true
	if err := i.Install(Spec{Apt: "zsh", Bin: "zsh"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestInstall_AptUpdateFailureCached(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.fail["apt-get update"] = true
	for _, s := range []Spec{{Apt: "zsh"}, {Apt: "git"}} {
		if err := i.Install(s); err == nil {
			t.Fatalf("expected error for %s", s.Apt)
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

func TestInstall_AptUsesSudoWhenNotRoot(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy zsh"] = "  Candidate: 5.9-4\n"
	if err := i.Install(Spec{Apt: "zsh", Bin: "zsh"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 3 {
		t.Fatalf("expected 3 calls, got %v", fr.calls)
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
