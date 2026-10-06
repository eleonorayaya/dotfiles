package app

import (
	"os"
	"os/exec"
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
