package utena

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const hookCommand = `utena hook post-checkout "$@" >/dev/null 2>&1 || :`

func gitConfigGetAll(t *testing.T, key string) []string {
	t.Helper()
	out, err := exec.Command("git", "config", "--global", "--get-all", key).Output()
	if err != nil {
		t.Fatalf("git config --get-all %s: %v", key, err)
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n")
}

func TestConfigureGitHook_IsIdempotent(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))

	for range 2 {
		if err := configureGitHook(); err != nil {
			t.Fatalf("configureGitHook: %v", err)
		}
	}

	if got := gitConfigGetAll(t, "hook.utena.command"); len(got) != 1 || got[0] != hookCommand {
		t.Fatalf("hook.utena.command = %v", got)
	}
	if got := gitConfigGetAll(t, "hook.utena.event"); len(got) != 1 || got[0] != "post-checkout" {
		t.Fatalf("hook.utena.event = %v", got)
	}
}

func TestConfigureGitHook_CommandHasNoConfigCommentChars(t *testing.T) {
	if strings.ContainsAny(hookCommand, ";#") {
		t.Fatalf("hook command contains ; or #: %s", hookCommand)
	}
}

func TestConfigureGitHook_CheckoutSucceedsWithoutBinary(t *testing.T) {
	cfg := filepath.Join(t.TempDir(), "gitconfig")
	t.Setenv("GIT_CONFIG_GLOBAL", cfg)

	if err := configureGitHook(); err != nil {
		t.Fatalf("configureGitHook: %v", err)
	}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	if err := os.Symlink(gitPath, filepath.Join(binDir, "git")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":/usr/bin:/bin")
	if _, err := exec.LookPath("utena"); err == nil {
		t.Skip("utena binary on PATH")
	}

	repo := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
	run("init", "-q")
	run("-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "--allow-empty", "-m", "init")
	run("checkout", "-q", "-b", "x")
}
