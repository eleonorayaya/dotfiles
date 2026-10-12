package utena

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

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

	if got := gitConfigGetAll(t, "hook.utena.command"); len(got) != 1 || got[0] != "utena hook post-checkout" {
		t.Fatalf("hook.utena.command = %v", got)
	}
	if got := gitConfigGetAll(t, "hook.utena.event"); len(got) != 1 || got[0] != "post-checkout" {
		t.Fatalf("hook.utena.event = %v", got)
	}
}
