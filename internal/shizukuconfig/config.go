package shizukuconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/internal/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Languages map[string]LanguageConfig `yaml:"languages"`
}

func newConfig() *Config {
	return &Config{
		Languages: createDefaultLanguageConfig(),
	}
}

func newConfigFromPath(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s: please create a shizuku.yml configuration file", configPath)
		}

		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid language configuration: %w", err)
	}

	return &c, nil
}

func (c *Config) validate() error {
	if err := validateLanguageConfig(c.Languages); err != nil {
		return fmt.Errorf("invalid language config: %w", err)
	}

	return nil
}

func (c *Config) save(configPath string) error {
	yamlData, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config to YAML: %w", err)
	}

	configDir := filepath.Dir(configPath)
	if err := util.EnsureDirExists(configDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
