package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
)

const baseGitConfig = "[user]\nname = Test User\nemail = work@example.com\n"

func testIdentityOptions() Options {
	return Options{
		Identities: []Identity{{
			Name:  "personal",
			Email: "dev@example.com",
			RemotePatterns: []string{
				"https://github.com/me/**",
				"git@github.com:me/**",
			},
		}},
	}
}

func generateWithHome(t *testing.T, opts Options) (string, *app.GenerateResult) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".gitconfig"), []byte(baseGitConfig), 0o644); err != nil {
		t.Fatalf("failed to write base gitconfig: %v", err)
	}

	result, err := New(opts).Generate(&app.Context{OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return home, result
}

func TestGenerateAddsIdentityIncludesAfterUserSection(t *testing.T) {
	_, result := generateWithHome(t, testIdentityOptions())

	data, err := os.ReadFile(result.FileMap[".gitconfig"])
	if err != nil {
		t.Fatalf("failed to read merged gitconfig: %v", err)
	}
	merged := string(data)

	userIdx := strings.Index(merged, "[user]")
	for _, pattern := range []string{"https://github.com/me/**", "git@github.com:me/**"} {
		section := `[includeIf "hasconfig:remote.*.url:` + pattern + `"]`
		idx := strings.Index(merged, section)
		if idx == -1 {
			t.Fatalf("expected %s in merged gitconfig, got:\n%s", section, merged)
		}
		if idx < userIdx {
			t.Errorf("expected %s after [user] so its email wins, got:\n%s", section, merged)
		}
	}
	if !strings.Contains(merged, "~/.config/git/identity-personal.gitconfig") {
		t.Errorf("expected include path for the personal identity, got:\n%s", merged)
	}
}

func TestGenerateWritesIdentityFile(t *testing.T) {
	_, result := generateWithHome(t, testIdentityOptions())

	path, ok := result.FileMap[".config/git/identity-personal.gitconfig"]
	if !ok {
		t.Fatalf("expected identity file in file map, got %v", result.FileMap)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read identity file: %v", err)
	}
	if !strings.Contains(string(data), "email = dev@example.com") {
		t.Errorf("expected identity email in file, got:\n%s", data)
	}
}

func TestGenerateWithoutIdentitiesAddsNoIncludes(t *testing.T) {
	_, result := generateWithHome(t, Options{})

	data, err := os.ReadFile(result.FileMap[".gitconfig"])
	if err != nil {
		t.Fatalf("failed to read merged gitconfig: %v", err)
	}
	if strings.Contains(string(data), "includeIf") {
		t.Errorf("expected no includeIf sections, got:\n%s", data)
	}
}

func TestIdentityAppliesOnlyToMatchingRemotes(t *testing.T) {
	home, result := generateWithHome(t, testIdentityOptions())
	if err := app.SyncAppFiles(result.FileMap, home); err != nil {
		t.Fatalf("failed to sync files: %v", err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(home, ".gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	cases := map[string]string{
		"https://github.com/me/repo.git":   "dev@example.com",
		"git@github.com:me/repo.git":       "dev@example.com",
		"https://github.com/work/repo.git": "work@example.com",
	}
	for remote, want := range cases {
		repo := t.TempDir()
		for _, args := range [][]string{{"init", "-q"}, {"remote", "add", "origin", remote}} {
			if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v failed: %s: %v", args, out, err)
			}
		}
		out, err := exec.Command("git", "-C", repo, "config", "user.email").Output()
		if err != nil {
			t.Fatalf("git config user.email failed for %s: %v", remote, err)
		}
		if got := strings.TrimSpace(string(out)); got != want {
			t.Errorf("remote %s: expected email %q, got %q", remote, want, got)
		}
	}
}
