package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateEnv_GuardedAlias(t *testing.T) {
	files, err := GenerateEnvFiles([]*EnvSetup{{Aliases: []Alias{
		{Name: "ls", Command: "lsd", Requires: "lsd"},
		{Name: "c", Command: "clear"},
	}}}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(files["shizuku.sh"])
	if err != nil {
		t.Fatal(err)
	}
	sh := string(data)
	if !strings.Contains(sh, "command -v lsd >/dev/null 2>&1 && alias ls='lsd'\n") {
		t.Errorf("missing guarded alias:\n%s", sh)
	}
	if !strings.Contains(sh, "\nalias c='clear'\n") {
		t.Errorf("missing plain alias:\n%s", sh)
	}
}

func TestGenerateEnv_AliasWithSingleQuotes(t *testing.T) {
	curltime := `curl -o /dev/null -s -w 'Total: %{time_total}s\n'`
	say := `echo 'hi'`
	files, err := GenerateEnvFiles([]*EnvSetup{{Aliases: []Alias{
		{Name: "curltime", Command: curltime},
		{Name: "say", Command: say, Requires: "echo"},
	}}}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(files["shizuku.sh"])
	if err != nil {
		t.Fatal(err)
	}
	sh := string(data)
	if !strings.Contains(sh, `alias curltime='curl -o /dev/null -s -w '\''Total: %{time_total}s\n'\'''`+"\n") {
		t.Errorf("plain alias not escaped:\n%s", sh)
	}
	if !strings.Contains(sh, `command -v echo >/dev/null 2>&1 && alias say='echo '\''hi'\'''`+"\n") {
		t.Errorf("guarded alias not escaped:\n%s", sh)
	}

	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}
	if out, err := exec.Command(zsh, "-n", files["shizuku.sh"]).CombinedOutput(); err != nil {
		t.Fatalf("zsh -n failed: %v\n%s", err, out)
	}
	for name, want := range map[string]string{"curltime": curltime, "say": say} {
		out, err := exec.Command(zsh, "-f", "-c", "source "+files["shizuku.sh"]+"; alias "+name+" >/dev/null && print -r -- \"$aliases["+name+"]\"").CombinedOutput()
		if err != nil {
			t.Fatalf("zsh source failed for %s: %v\n%s", name, err, out)
		}
		if string(out) != want+"\n" {
			t.Errorf("alias %s = %q, want %q", name, out, want)
		}
	}
}

const testSecretsLine = `[ -f "$HOME/.config/shizuku/secrets.env" ] && { set -a; . "$HOME/.config/shizuku/secrets.env"; set +a; }`

func TestGenerateEnv_ZshenvScripts(t *testing.T) {
	files, err := GenerateEnvFiles([]*EnvSetup{{
		Variables:     []EnvVar{{Key: "EDITOR", Value: "hx"}},
		ZshenvScripts: []string{testSecretsLine},
	}}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	zshenv, err := os.ReadFile(files["shizuku.zshenv"])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(zshenv), testSecretsLine+"\n") {
		t.Errorf("zshenv missing script:\n%s", zshenv)
	}
	if strings.Index(string(zshenv), testSecretsLine) < strings.Index(string(zshenv), "export EDITOR=") {
		t.Errorf("expected scripts after variables:\n%s", zshenv)
	}
	sh, err := os.ReadFile(files["shizuku.sh"])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(sh), testSecretsLine) {
		t.Errorf("shizuku.sh should not contain zshenv script:\n%s", sh)
	}
}

func TestGenerateEnv_ZshenvScriptsLoadSecretsInZsh(t *testing.T) {
	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh not available")
	}
	files, err := GenerateEnvFiles([]*EnvSetup{{ZshenvScripts: []string{testSecretsLine}}}, t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := map[string]struct {
		secrets string
		want    string
	}{
		"present": {secrets: "X=1\n", want: "1\n"},
		"missing": {want: "\n"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			if tc.secrets != "" {
				dir := filepath.Join(home, ".config", "shizuku")
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "secrets.env"), []byte(tc.secrets), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command(zsh, "-f", "-c", "source "+files["shizuku.zshenv"]+"; zsh -f -c 'echo $X'")
			cmd.Env = []string{"HOME=" + home, "PATH=" + os.Getenv("PATH")}
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("zsh failed: %v\n%s", err, out)
			}
			if string(out) != tc.want {
				t.Errorf("echo $X = %q, want %q", out, tc.want)
			}
		})
	}
}
