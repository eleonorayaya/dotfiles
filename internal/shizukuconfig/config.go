package shizukuconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eleonorayaya/shizuku/internal/util"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Styles    Styles          `yaml:"styles"`
	Languages LanguageConfigs `yaml:"languages"`
	Apps      map[string]any  `yaml:"apps"`
}

func newConfig() *Config {
	return &Config{
		Styles:    createDefaultStyles(),
		Languages: createDefaultLanguageConfig(),
	}
}

func newConfigFromPath(configPath string) (*Config, error) {
	c, err := loadConfigFromPath(configPath)
	if err != nil {
		return nil, err
	}

	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return c, nil
}

func loadConfigFromPath(configPath string) (*Config, error) {
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

	return &c, nil
}

func (c *Config) validate() error {
	if err := c.Styles.validate(); err != nil {
		return fmt.Errorf("invalid styles config: %w", err)
	}

	if err := c.Languages.validate(); err != nil {
		return fmt.Errorf("invalid language config: %w", err)
	}

	return nil
}

func (c *Config) GetAppConfig(appName string, configKey string) (any, bool) {
	if c.Apps == nil {
		return nil, false
	}

	appConfig, ok := c.Apps[appName]
	if !ok {
		return nil, false
	}

	appMap, ok := appConfig.(map[string]any)
	if !ok {
		return nil, false
	}

	value, ok := appMap[configKey]
	if !ok {
		return nil, false
	}

	return value, true
}

func (c *Config) GetAppConfigBool(appName string, configKey string, defaultValue bool) bool {
	value, ok := c.GetAppConfig(appName, configKey)
	if !ok {
		return defaultValue
	}

	boolValue, ok := value.(bool)
	if !ok {
		return defaultValue
	}

	return boolValue
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

func (c *Config) mergeWithDefaults(defaultConfig *Config) {
	c.Styles.merge(defaultConfig.Styles)

	if c.Languages == nil {
		c.Languages = defaultConfig.Languages
	} else {
		c.Languages.merge(defaultConfig.Languages)
	}

	if c.Apps == nil {
		c.Apps = defaultConfig.Apps
	} else if defaultConfig.Apps != nil {
		c.Apps = mergeAppConfigs(c.Apps, defaultConfig.Apps)
	}
}

func mergeAppConfigs(existing, defaults map[string]any) map[string]any {
	return util.MergeStringAnyMap(existing, defaults)
}
