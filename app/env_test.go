package app

import (
	"os"
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
