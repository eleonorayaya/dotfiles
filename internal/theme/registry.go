package theme

import "fmt"

type ThemeLoader func() *Theme

var themeRegistry = make(map[string]ThemeLoader)

func Register(name string, loader ThemeLoader) {
	themeRegistry[name] = loader
}

func LoadTheme(themeName string) (*Theme, error) {
	loader, exists := themeRegistry[themeName]
	if !exists {
		return nil, fmt.Errorf("unknown theme '%s'", themeName)
	}

	themeData := loader()

	if err := themeData.Colors.Validate(); err != nil {
		return nil, fmt.Errorf("theme '%s' validation failed: %w", themeName, err)
	}

	return themeData, nil
}
