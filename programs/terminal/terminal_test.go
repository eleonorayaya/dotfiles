package terminal

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/pkg"
)

func stubInstall(t *testing.T, install func(pkg.Spec) error, clone func(repo, dest string) error) {
	t.Helper()
	origInstall, origClone := installPkg, cloneRepo
	installPkg, cloneRepo = install, clone
	t.Cleanup(func() { installPkg, cloneRepo = origInstall, origClone })
}

func writePlugin(dest string) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dest, zshViModePlugin), []byte("x"), 0o644)
}

func TestInstall_ContinuesPastFailures(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	attempted := 0
	stubInstall(t,
		func(s pkg.Spec) error { attempted++; return errors.New("boom") },
		func(repo, dest string) error { return errors.New("clone failed") },
	)

	err := New().Install(nil)
	if err == nil {
		t.Fatal("expected joined error")
	}
	if strings.Contains(err.Error(), "zsh-vi-mode") {
		t.Errorf("zsh-vi-mode failure should not be an error: %v", err)
	}
	want := 0
	for _, p := range packages {
		if p.skipIfExists == "" {
			want++
		} else if _, statErr := os.Stat(p.skipIfExists); statErr != nil {
			want++
		}
	}
	if attempted != want {
		t.Errorf("attempted %d packages, want %d", attempted, want)
	}

	zshenv, _ := os.ReadFile(filepath.Join(home, ".zshenv"))
	if string(zshenv) != "[[ -r ~/.config/shizuku/shizuku.zshenv ]] && source ~/.config/shizuku/shizuku.zshenv\n" {
		t.Errorf("zshenv got %q", zshenv)
	}
	zshrc, _ := os.ReadFile(filepath.Join(home, ".zshrc"))
	if string(zshrc) != "[[ -r ~/.config/shizuku/shizuku.sh ]] && source ~/.config/shizuku/shizuku.sh\n" {
		t.Errorf("zshrc got %q", zshrc)
	}
}

func TestInstallPackages_SkipIfExists(t *testing.T) {
	present := filepath.Join(t.TempDir(), "ca-certificates.crt")
	if err := os.WriteFile(present, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var installed []string
	stubInstall(t, func(s pkg.Spec) error { installed = append(installed, s.Apt); return nil }, nil)

	err := installPackages([]terminalPackage{
		{spec: pkg.Spec{Apt: "ca-certificates"}, skipIfExists: present},
		{spec: pkg.Spec{Apt: "zsh"}, skipIfExists: filepath.Join(t.TempDir(), "missing")},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(installed) != 1 || installed[0] != "zsh" {
		t.Errorf("installed %v, want [zsh]", installed)
	}
}

func TestEnsureZshViMode_SkipsWhenPluginPresent(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, _ := zshViModeDir()
	if err := writePlugin(path); err != nil {
		t.Fatal(err)
	}
	stubInstall(t, nil, func(repo, dest string) error { t.Error("clone should not run"); return nil })

	if err := ensureZshViMode(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureZshViMode_ReplacesIncompleteDir(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, _ := zshViModeDir()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "stale"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	stubInstall(t, nil, func(repo, dest string) error {
		if _, err := os.Stat(dest); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("dest should be removed before clone, stat err: %v", err)
		}
		return writePlugin(dest)
	})

	if err := ensureZshViMode(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(path, zshViModePlugin)); err != nil {
		t.Errorf("plugin missing: %v", err)
	}
}

func TestEnsureZshViMode_CleansUpFailedClone(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path, _ := zshViModeDir()
	stubInstall(t, nil, func(repo, dest string) error {
		if err := os.MkdirAll(dest, 0o755); err != nil {
			return err
		}
		return errors.New("network down")
	})

	if err := ensureZshViMode(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("partial clone left behind, stat err: %v", err)
	}
}

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
