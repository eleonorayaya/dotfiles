package shizukuconfig

import (
	"fmt"
	"os"

	"github.com/eleonorayaya/shizuku/internal/util"
)

const (
	ConfigFilePath = "~/.config/shizuku/shizuku.yml"
	SourceDir      = "~/.local/src/shizuku"
)

func LoadConfig() (*Config, error) {
	configPath, err := util.NormalizeFilePath(ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize config path: %w", err)
	}

	config, err := newConfigFromPath(configPath)
	if err != nil {
		return nil, err
	}

	config.mergeWithDefaults(newConfig())

	return config, nil
}

func InitConfig() (bool, string, error) {
	configPath, err := util.NormalizeFilePath(ConfigFilePath)
	if err != nil {
		return false, "", fmt.Errorf("failed to normalize config path: %w", err)
	}

	if _, err := os.Stat(configPath); err == nil {
		existingConfig, err := loadConfigFromPath(configPath)
		if err != nil {
			return false, "", fmt.Errorf("failed to load existing config: %w", err)
		}

		defaultConfig := newConfig()
		existingConfig.mergeWithDefaults(defaultConfig)

		if err := existingConfig.save(configPath); err != nil {
			return false, "", fmt.Errorf("failed to save merged config: %w", err)
		}

		return false, configPath, nil
	}

	defaultConfig := newConfig()
	if err := defaultConfig.save(configPath); err != nil {
		return false, "", fmt.Errorf("failed to save config: %w", err)
	}

	return true, configPath, nil
}
