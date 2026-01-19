package shizukuconfig

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/util"
)

type Language string

const (
	LanguageRust Language = "rust"
)

var languages []Language = []Language{
	LanguageRust,
}

type LanguageConfig struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:",inline"`
}

func createDefaultLanguageConfig() map[string]LanguageConfig {
	defaultConfig := make(map[string]LanguageConfig)

	for _, lang := range languages {
		defaultConfig[string(lang)] = LanguageConfig{
			Enabled: false,
			Config:  make(map[string]any),
		}
	}

	return defaultConfig
}

func validateLanguageConfig(languageConfig map[string]LanguageConfig) error {
	if languageConfig == nil {
		return nil
	}

	validLangs := make(map[string]bool)
	for _, lang := range languages {
		validLangs[string(lang)] = true
	}

	for langName := range languageConfig {
		if !validLangs[langName] {
			return fmt.Errorf("unsupported language '%s': valid languages are %v", langName, languages)
		}
	}

	return nil
}

func mergeLanguageConfigs(existing, defaults map[string]LanguageConfig) map[string]LanguageConfig {
	result := make(map[string]LanguageConfig)

	for lang, config := range existing {
		result[lang] = config
	}

	for lang, defaultConfig := range defaults {
		if existingLang, exists := result[lang]; exists {
			existingLang.Config = util.MergeStringAnyMap(existingLang.Config, defaultConfig.Config)
			result[lang] = existingLang
		} else {
			result[lang] = defaultConfig
		}
	}

	return result
}
