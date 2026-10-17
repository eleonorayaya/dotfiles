package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/eleonorayaya/shizuku/util"
)

const globalConfigPath = "~/.claude.json"

var renameFile = os.Rename

func syncGlobalConfig(path string, values map[string]any, removeServers []string) error {
	if len(values) == 0 && len(removeServers) == 0 {
		return nil
	}

	cfg, err := util.ReadJSONMap(path)
	if err != nil {
		return fmt.Errorf("failed to read global config: %w", err)
	}

	changed := false
	for key, value := range values {
		if !reflect.DeepEqual(cfg[key], value) {
			cfg[key] = value
			changed = true
		}
	}

	servers, _ := cfg["mcpServers"].(map[string]any)
	for _, name := range removeServers {
		if _, ok := servers[name]; ok {
			delete(servers, name)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	return writeGlobalConfig(path, cfg)
}

func writeGlobalConfig(path string, cfg map[string]any) error {
	normalized, err := util.NormalizeFilePath(path)
	if err != nil {
		return fmt.Errorf("failed to normalize global config path: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	tmp := normalized + ".shizuku-tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}
	if err := renameFile(tmp, normalized); err != nil {
		return fmt.Errorf("failed to replace global config: %w", err)
	}
	return nil
}
