package tuios

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func keepHosts(generatedPath, existingPath string) error {
	data, err := os.ReadFile(existingPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", existingPath, err)
	}

	var existing map[string]any
	if err := toml.Unmarshal(data, &existing); err != nil {
		return fmt.Errorf("failed to parse %s: %w", existingPath, err)
	}

	hosts, ok := existing["hosts"]
	if !ok {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString("\n")
	enc := toml.NewEncoder(&buf)
	enc.Indent = ""
	if err := enc.Encode(map[string]any{"hosts": hosts}); err != nil {
		return fmt.Errorf("failed to encode hosts: %w", err)
	}

	f, err := os.OpenFile(generatedPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", generatedPath, err)
	}
	defer f.Close()

	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("failed to append hosts: %w", err)
	}

	return nil
}
