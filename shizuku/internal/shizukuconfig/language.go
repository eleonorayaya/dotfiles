package shizukuconfig

import "fmt"

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
