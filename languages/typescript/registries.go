package typescript

import (
	"fmt"
	"net/url"
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

	return append(npmHosts, yarnHosts...), nil
}

func npmrcHosts(path string) ([]string, error) {
	data, err := util.ReadFileIfExists(path)
	if err != nil || data == nil {
		return nil, err
	}

	var hosts []string
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
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
	return hosts, nil
}

func yarnrcHosts(path string) ([]string, error) {
	data, err := util.ReadFileIfExists(path)
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
