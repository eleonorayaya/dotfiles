package util

import (
	"path/filepath"
	"testing"
)

func TestCloneOrPull(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, "gitconfig"))

	origin := filepath.Join(home, "origin")
	gitIn := func(dir string, args ...string) {
		t.Helper()
		if err := Run(dir, "git", args...); err != nil {
			t.Fatal(err)
		}
	}
	gitIn("", "init", "--initial-branch", "dev", origin)
	gitIn(origin, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "--allow-empty", "-m", "one")

	path, err := CloneOrPull(origin, "dev", "~/src/repo")
	if err != nil {
		t.Fatalf("clone: %v", err)
	}
	if want := filepath.Join(home, "src", "repo"); path != want {
		t.Fatalf("path = %s, want %s", path, want)
	}

	writeFile(t, filepath.Join(origin, "new.txt"), "x", 0o644)
	gitIn(origin, "add", "new.txt")
	gitIn(origin, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-m", "two")

	if _, err := CloneOrPull(origin, "dev", "~/src/repo"); err != nil {
		t.Fatalf("pull: %v", err)
	}
	if got := readFile(t, filepath.Join(path, "new.txt")); got != "x" {
		t.Fatalf("pulled file = %q", got)
	}
}
