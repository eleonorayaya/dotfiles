package shizukuconfig

import (
	"fmt"

	"github.com/eleonorayaya/shizuku/internal/util"
)

type Language string

const (
	LanguageGo         Language = "go"
	LanguageLua        Language = "lua"
	LanguageRust       Language = "rust"
	LanguageTypescript Language = "typescript"
)

var languages []Language = []Language{
	LanguageGo,
	LanguageLua,
	LanguageRust,
	LanguageTypescript,
}

type LanguageConfig struct {
	Enabled bool           `yaml:"enabled"`
	Config  map[string]any `yaml:",inline"`
}

type LanguageConfigs map[string]LanguageConfig

func createDefaultLanguageConfig() LanguageConfigs {
	defaultConfig := make(LanguageConfigs)

	for _, lang := range languages {
		defaultConfig[string(lang)] = LanguageConfig{
			Enabled: false,
			Config:  make(map[string]any),
		}
	}

	return defaultConfig
}

func (lc LanguageConfigs) validate() error {
	if lc == nil {
		return nil
	}

	validLangs := make(map[string]bool)
	for _, lang := range languages {
		validLangs[string(lang)] = true
	}

	for langName := range lc {
		if !validLangs[langName] {
			return fmt.Errorf("unsupported language '%s': valid languages are %v", langName, languages)
		}
	}

	return nil
}

func (lc LanguageConfigs) merge(defaults LanguageConfigs) {
	for lang, defaultConfig := range defaults {
		if existingLang, exists := lc[lang]; exists {
			existingLang.Config = util.MergeStringAnyMap(existingLang.Config, defaultConfig.Config)
			lc[lang] = existingLang
		} else {
			lc[lang] = defaultConfig
		}
	}
}
