package typescript

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/eleonorayaya/shizuku/util"
	"gopkg.in/yaml.v3"
)

type yarnrcScope struct {
	NpmRegistryServer string `yaml:"npmRegistryServer"`
}

type yarnrc struct {
	NpmRegistryServer string                 `yaml:"npmRegistryServer"`
	NpmScopes         map[string]yarnrcScope `yaml:"npmScopes"`
	NpmRegistries     map[string]any         `yaml:"npmRegistries"`
}

func registryHosts(npmrcPath, yarnrcPath string) ([]string, error) {
	npmHosts, err := npmrcHosts(npmrcPath)
	if err != nil {
		return nil, err
	}
	yarnHosts, err := yarnrcHosts(yarnrcPath)
	if err != nil {
		return nil, err
	}

	hosts := append(npmHosts, yarnHosts...)
	slices.Sort(hosts)
	return slices.Compact(hosts), nil
}

func readOptional(path string) ([]byte, error) {
	normalized, err := util.NormalizeFilePath(path)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize path: %w", err)
	}

	data, err := os.ReadFile(normalized)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return data, nil
}

func npmrcHosts(path string) ([]string, error) {
	data, err := readOptional(path)
	if err != nil || data == nil {
		return nil, err
	}

	var hosts []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		switch {
		case strings.HasPrefix(key, "//"):
			hosts = appendHost(hosts, key)
		case key == "registry" || strings.HasSuffix(key, ":registry"):
			hosts = appendHost(hosts, strings.TrimSpace(value))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", path, err)
	}
	return hosts, nil
}

func yarnrcHosts(path string) ([]string, error) {
	data, err := readOptional(path)
	if err != nil || data == nil {
		return nil, err
	}

	var cfg yarnrc
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}

	hosts := appendHost(nil, cfg.NpmRegistryServer)
	for _, scope := range cfg.NpmScopes {
		hosts = appendHost(hosts, scope.NpmRegistryServer)
	}
	for registry := range cfg.NpmRegistries {
		hosts = appendHost(hosts, registry)
	}
	return hosts, nil
}

func appendHost(hosts []string, raw string) []string {
	raw = strings.Trim(raw, `"'`)
	if raw == "" {
		return hosts
	}
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return hosts
	}
	return append(hosts, u.Hostname())
}
