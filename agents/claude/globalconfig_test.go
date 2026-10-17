package claude

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eleonorayaya/shizuku/util"
)

func countingRename(t *testing.T) *int {
	t.Helper()
	count := 0
	orig := renameFile
	renameFile = func(oldpath, newpath string) error {
		count++
		return orig(oldpath, newpath)
	}
	t.Cleanup(func() { renameFile = orig })
	return &count
}

func assertPerm0600(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestSyncGlobalConfigMergesValuesKeepingOtherKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	if err := os.WriteFile(path, []byte(`{"numStartups": 3, "leftArrowOpensAgents": true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := syncGlobalConfig(path, map[string]any{"leftArrowOpensAgents": false}, nil); err != nil {
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
	assertPerm0600(t, path)
}

func TestSyncGlobalConfigRemovesMcpServers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	seed := `{"numStartups": 3, "mcpServers": {"homeassistant": {"type": "http"}, "zoo": {"type": "stdio"}, "keep": {"type": "stdio"}}}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := syncGlobalConfig(path, nil, []string{"homeassistant", "zoo", "absent"}); err != nil {
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
	assertPerm0600(t, path)
}

func TestSyncGlobalConfigAppliesBothChangesInOneWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")
	seed := `{"numStartups": 3, "leftArrowOpensAgents": true, "mcpServers": {"zoo": {"type": "stdio"}, "keep": {"type": "stdio"}}}`
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatal(err)
	}
	renames := countingRename(t)

	if err := syncGlobalConfig(path, map[string]any{"leftArrowOpensAgents": false}, []string{"zoo"}); err != nil {
		t.Fatal(err)
	}

	if *renames != 1 {
		t.Errorf("expected 1 write, got %d", *renames)
	}
	cfg, err := util.ReadJSONMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["leftArrowOpensAgents"] != false {
		t.Errorf("expected leftArrowOpensAgents false, got %v", cfg["leftArrowOpensAgents"])
	}
	servers, _ := cfg["mcpServers"].(map[string]any)
	if _, ok := servers["zoo"]; ok {
		t.Errorf("expected zoo removed, got %v", servers)
	}
	if _, ok := servers["keep"]; !ok {
		t.Errorf("expected keep kept, got %v", servers)
	}
	if cfg["numStartups"] != float64(3) {
		t.Errorf("expected other keys kept, got %v", cfg["numStartups"])
	}
	assertPerm0600(t, path)
}

func TestSyncGlobalConfigLeavesUnchangedFileAlone(t *testing.T) {
	cases := map[string]struct {
		seed   string
		values map[string]any
		names  []string
	}{
		"value already set":      {seed: `{"leftArrowOpensAgents":false}`, values: map[string]any{"leftArrowOpensAgents": false}},
		"no server match":        {seed: `{"mcpServers":{"keep":{}}}`, names: []string{"zoo"}},
		"no mcpServers":          {seed: `{"numStartups":3}`, names: []string{"zoo"}},
		"empty names":            {seed: `{"mcpServers":{"zoo":{}}}`, names: nil},
		"nothing to do":          {seed: `{"mcpServers":{"zoo":{}}}`},
		"value set and no match": {seed: `{"a":1,"mcpServers":{"keep":{}}}`, values: map[string]any{"a": float64(1)}, names: []string{"zoo"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "claude.json")
			if err := os.WriteFile(path, []byte(tc.seed), 0o600); err != nil {
				t.Fatal(err)
			}
			renames := countingRename(t)

			if err := syncGlobalConfig(path, tc.values, tc.names); err != nil {
				t.Fatal(err)
			}

			if *renames != 0 {
				t.Errorf("expected no write, got %d", *renames)
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

func TestSyncGlobalConfigMissingFile(t *testing.T) {
	cases := map[string]struct {
		values map[string]any
		names  []string
	}{
		"nothing to do":     {},
		"servers to remove": {names: []string{"zoo"}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "claude.json")

			if err := syncGlobalConfig(path, tc.values, tc.names); err != nil {
				t.Fatal(err)
			}

			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("expected no file created, got %v", err)
			}
		})
	}
}

func TestSyncGlobalConfigMissingFileWithValuesCreatesIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "claude.json")

	if err := syncGlobalConfig(path, map[string]any{"leftArrowOpensAgents": false}, []string{"zoo"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := util.ReadJSONMap(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg["leftArrowOpensAgents"] != false {
		t.Errorf("expected leftArrowOpensAgents false, got %v", cfg)
	}
	if _, ok := cfg["mcpServers"]; ok {
		t.Errorf("expected no mcpServers key added, got %v", cfg)
	}
	assertPerm0600(t, path)
}
