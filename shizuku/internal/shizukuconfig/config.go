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

const (
	ConfigFilePath = "~/.config/shizuku/shizuku.yml"
)

func LoadConfig() (*Config, error) {
	return LoadConfigFromPath(ConfigFilePath)
}

func LoadConfigFromPath(configFilePath string) (*Config, error) {
	configPath, err := util.NormalizeFilePath(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize config path: %w", err)
	}

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

func NormalizeConfigPath() (string, error) {
	return util.NormalizeFilePath(ConfigFilePath)
}

func ConfigDir() (string, error) {
	configPath, err := NormalizeConfigPath()
	if err != nil {
		return "", err
	}

	return filepath.Dir(configPath), nil
}
