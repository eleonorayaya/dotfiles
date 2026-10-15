package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/util"
)

func TestMergeGlobalConfigKeepsOtherKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	if err := os.WriteFile(path, []byte(`{"numStartups": 3, "leftArrowOpensAgents": true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := mergeGlobalConfig(path, map[string]any{"leftArrowOpensAgents": false}); err != nil {
		t.Fatal(err)
	}

	cfg, err := util.ReadJSONMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["leftArrowOpensAgents"] != false {
		t.Errorf("expected leftArrowOpensAgents false, got %v", cfg["leftArrowOpensAgents"])
	}
	if cfg["numStartups"] != float64(3) {
		t.Errorf("expected other keys kept, got %v", cfg["numStartups"])
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestMergeGlobalConfigLeavesUnchangedFileAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	original := []byte(`{"leftArrowOpensAgents":false}`)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := mergeGlobalConfig(path, map[string]any{"leftArrowOpensAgents": false}); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf("expected file untouched, got %s", got)
	}
}
