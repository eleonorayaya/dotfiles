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

func TestRemoveGlobalMcpServers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	seed := `{"numStartups": 3, "mcpServers": {"homeassistant": {"type": "http"}, "zoo": {"type": "stdio"}, "keep": {"type": "stdio"}}}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := removeGlobalMcpServers(path, []string{"homeassistant", "zoo", "absent"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := util.ReadJSONMap(path)
	if err != nil {
		t.Fatal(err)
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if _, ok := servers["homeassistant"]; ok {
		t.Errorf("expected homeassistant removed, got %v", servers)
	}
	if _, ok := servers["zoo"]; ok {
		t.Errorf("expected zoo removed, got %v", servers)
	}
	if _, ok := servers["keep"]; !ok {
		t.Errorf("expected keep kept, got %v", servers)
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

func TestRemoveGlobalMcpServersLeavesUnchangedFileAlone(t *testing.T) {
	cases := map[string]struct {
		seed  string
		names []string
	}{
		"no match":      {seed: `{"mcpServers":{"keep":{}}}`, names: []string{"zoo"}},
		"no mcpServers": {seed: `{"numStartups":3}`, names: []string{"zoo"}},
		"empty names":   {seed: `{"mcpServers":{"zoo":{}}}`, names: nil},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "claude.json")
			if err := os.WriteFile(path, []byte(tc.seed), 0o600); err != nil {
				t.Fatal(err)
			}

			if err := removeGlobalMcpServers(path, tc.names); err != nil {
				t.Fatal(err)
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.seed {
				t.Errorf("expected file untouched, got %s", got)
			}
		})
	}
}

func TestRemoveGlobalMcpServersMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")

	if err := removeGlobalMcpServers(path, []string{"zoo"}); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected no file created, got %v", err)
	}
}
