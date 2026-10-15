package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpgrade_BrewUpgradesEvenWhenOnPath(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	i.Upgrade = true
	fr.paths["bat"] = "/opt/homebrew/bin/bat"
	if err := i.Install(Spec{Brew: "bat", Bin: "bat"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "brew list bat|brew update|brew upgrade bat"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if !strings.Contains(fr.envs[2], "HOMEBREW_NO_AUTO_UPDATE=1") {
		t.Errorf("brew upgrade env = %q, want HOMEBREW_NO_AUTO_UPDATE=1", fr.envs[2])
	}
}

func TestUpgrade_BrewSkipsWhenNotInstalled(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	i.Upgrade = true
	fr.fail["brew list glow"] = true
	if err := i.Install(Spec{Brew: "glow"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "brew list glow"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_BrewCask(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	i.Upgrade = true
	if err := i.Install(Spec{Brew: "kitty", BrewCask: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "brew list kitty --cask|brew update|brew upgrade kitty --cask"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_BrewUpdateRunsOnce(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	i.Upgrade = true
	for _, name := range []string{"bat", "lsd"} {
		if err := i.Install(Spec{Brew: name}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	want := "brew list bat|brew update|brew upgrade bat|brew list lsd|brew upgrade lsd"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_BrewUpdateFailureCached(t *testing.T) {
	i, fr := testInstaller(t, "darwin")
	i.Upgrade = true
	fr.fail["brew update"] = true
	for _, name := range []string{"bat", "lsd"} {
		if err := i.Install(Spec{Brew: name}); err == nil {
			t.Fatalf("expected error for %s", name)
		}
	}
	want := "brew list bat|brew update|brew list lsd"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_AptOnlyUpgrade(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.paths["git"] = "/usr/bin/git"
	fr.outputs["apt-cache policy git"] = "git:\n  Installed: 1:2.43.0-1\n  Candidate: 1:2.43.1-1\n"
	if err := i.Install(Spec{Brew: "git", Apt: "git", Bin: "git"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "apt-get update|apt-cache policy git|apt-get install --only-upgrade -y git"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_AptSkipsWhenNotInstalled(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	fr.outputs["apt-cache policy lsd"] = "lsd:\n  Installed: (none)\n  Candidate: 0.23.1-1\n"
	if err := i.Install(Spec{Apt: "lsd", Bin: "lsd"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "apt-get update|apt-cache policy lsd"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_AptSkipsUnknownPackage(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	if err := i.Install(Spec{Apt: "oh-my-posh", Bin: "oh-my-posh"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "apt-get update|apt-cache policy oh-my-posh"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_AptSymlinkedBinUsesApt(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	i.IsRoot = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	if err := os.MkdirAll(i.BinDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/usr/bin/batcat", filepath.Join(i.BinDir(), "bat")); err != nil {
		t.Fatal(err)
	}
	fr.outputs["apt-cache policy bat"] = "bat:\n  Installed: 0.24.0-1\n  Candidate: 0.24.0-1\n"
	err := i.Install(Spec{Apt: "bat", AptBin: "batcat", Bin: "bat", Release: &Release{Repo: "sharkdp/bat", Asset: `bat-{arch}`}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "apt-get update|apt-cache policy bat|apt-get install --only-upgrade -y bat"
	if got := strings.Join(fr.calls, "|"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpgrade_ReleaseReinstallsLatest(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	fr.paths["apt-get"] = "/usr/bin/apt-get"
	if err := os.MkdirAll(i.BinDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(i.BinDir(), "oh-my-posh"), []byte("OLD"), 0o755); err != nil {
		t.Fatal(err)
	}
	srv := releaseServer(t, map[string][]byte{"posh-linux-amd64": []byte("NEW")})
	i.APIBase = srv.URL

	err := i.Install(Spec{Apt: "oh-my-posh", Bin: "oh-my-posh", Release: &Release{
		Repo: "JanDeDobbeleer/oh-my-posh", Asset: `posh-linux-{arch}`,
		Arch: map[string]string{"amd64": "amd64"},
	}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertFile(t, filepath.Join(i.BinDir(), "oh-my-posh"), "NEW")
	if len(fr.calls) != 0 {
		t.Errorf("expected no commands, got %v", fr.calls)
	}
}

func TestUpgrade_LinuxSkipsWhenNothingApplies(t *testing.T) {
	i, fr := testInstaller(t, "linux")
	i.Upgrade = true
	if err := i.Install(Spec{Brew: "glow"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fr.calls) != 0 {
		t.Errorf("expected no commands, got %v", fr.calls)
	}
}
