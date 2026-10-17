package utena

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
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

func TestGenerate_ThemeHasNoEmptyColors(t *testing.T) {
	ctx := &app.Context{
		OutDir: t.TempDir(),
		Styles: styles.New(styles.WithTheme(themes.MonadeDark)),
	}

	result, err := New(Options{}).Generate(ctx)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	raw, err := os.ReadFile(result.FileMap["theme.json"])
	if err != nil {
		t.Fatalf("read theme.json: %v", err)
	}

	var theme map[string]string
	if err := json.Unmarshal(raw, &theme); err != nil {
		t.Fatalf("parse theme.json: %v", err)
	}

	for key, value := range theme {
		if value == "" {
			t.Errorf("theme.json key %q is empty", key)
		}
	}
}

func TestGenerate_DaemonBrowser(t *testing.T) {
	cases := map[string]struct {
		app  *App
		want string
	}{
		"default": {app: New(Options{}), want: ""},
		"firefox": {app: New(Options{Browser: "firefox"}), want: "firefox"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := &app.Context{
				OutDir: t.TempDir(),
				Styles: styles.New(styles.WithTheme(themes.MonadeDark)),
			}
			result, err := tc.app.Generate(ctx)
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			raw, err := os.ReadFile(result.FileMap["daemon.json"])
			if err != nil {
				t.Fatalf("read daemon.json: %v", err)
			}
			var daemon struct {
				Browser string `json:"browser"`
			}
			if err := json.Unmarshal(raw, &daemon); err != nil {
				t.Fatalf("parse daemon.json: %v", err)
			}
			if daemon.Browser != tc.want {
				t.Errorf("browser = %q, want %q", daemon.Browser, tc.want)
			}
		})
	}
}

func TestGenerate_DaemonBundles(t *testing.T) {
	cases := map[string]struct {
		opts Options
		want string
	}{
		"none":    {opts: Options{}, want: `{}`},
		"browser": {opts: Options{Browser: "firefox"}, want: `{"browser":"firefox"}`},
		"bundles": {
			opts: Options{Browser: "firefox", Bundles: map[string]app.Bundle{
				"go":   {Plugins: []string{"gopls-lsp@claude-plugins-official", "go@go-skills"}},
				"docs": {Plugins: []string{"notion-workspace-plugin@notion-plugin-marketplace"}},
			}},
			want: `{"browser":"firefox","bundles":{` +
				`"docs":{"plugins":["notion-workspace-plugin@notion-plugin-marketplace"]},` +
				`"go":{"plugins":["gopls-lsp@claude-plugins-official","go@go-skills"]}}}`,
		},
		"mcp servers": {
			opts: Options{
				Bundles: map[string]app.Bundle{
					"homelab": {McpServers: []string{"homeassistant"}},
					"mixed":   {Plugins: []string{"p@m"}, McpServers: []string{"zoo"}},
					"empty":   {},
				},
				McpServers: map[string]map[string]any{
					"homeassistant": {"type": "http", "url": "https://example.test/mcp"},
					"zoo":           {"type": "stdio", "command": "uvx", "args": []any{"zoo-mcp"}},
				},
			},
			want: `{"bundles":{` +
				`"empty":{},` +
				`"homelab":{"mcp_servers":["homeassistant"]},` +
				`"mixed":{"mcp_servers":["zoo"],"plugins":["p@m"]}},` +
				`"mcp_servers":{` +
				`"homeassistant":{"type":"http","url":"https://example.test/mcp"},` +
				`"zoo":{"args":["zoo-mcp"],"command":"uvx","type":"stdio"}}}`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := &app.Context{
				OutDir: t.TempDir(),
				Styles: styles.New(styles.WithTheme(themes.MonadeDark)),
			}
			result, err := New(tc.opts).Generate(ctx)
			if err != nil {
				t.Fatalf("Generate: %v", err)
			}
			raw, err := os.ReadFile(result.FileMap["daemon.json"])
			if err != nil {
				t.Fatalf("read daemon.json: %v", err)
			}
			var got, want any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatalf("parse daemon.json: %v\n%s", err, raw)
			}
			if err := json.Unmarshal([]byte(tc.want), &want); err != nil {
				t.Fatal(err)
			}
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Errorf("daemon.json = %s, want %s", gotJSON, wantJSON)
			}
		})
	}
}

func TestAgentConfigProvidesShellInit(t *testing.T) {
	cfg := New(Options{}).AgentConfig()
	if len(cfg.ShellInitScripts) != 1 || cfg.ShellInitScripts[0] != `eval "$(utena shell-init)"` {
		t.Errorf("ShellInitScripts = %v", cfg.ShellInitScripts)
	}
}
