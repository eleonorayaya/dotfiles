package shizukuconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Styles struct {
	ThemeName     string `yaml:"theme"`
	WindowOpacity int    `yaml:"windowOpacity"`
	Theme         *Theme `yaml:"-"`
}

func (s *Styles) UnmarshalYAML(node *yaml.Node) error {
	type alias Styles
	aux := (*alias)(s)
	if err := node.Decode(aux); err != nil {
		return err
	}

	if s.ThemeName != "" {
		theme, err := loadThemeFromRegistry(s.ThemeName)
		if err != nil {
			return fmt.Errorf("failed to load theme: %w", err)
		}
		s.Theme = theme
	}

	return nil
}

func createDefaultStyles() Styles {
	theme, _ := loadThemeFromRegistry("monade")
	return Styles{
		ThemeName:     "monade",
		WindowOpacity: 100,
		Theme:         theme,
	}
}

func (s *Styles) validate() error {
	if s.ThemeName == "" {
		return fmt.Errorf("theme name cannot be empty")
	}

	if s.WindowOpacity < 0 || s.WindowOpacity > 100 {
		return fmt.Errorf("windowOpacity must be between 0 and 100, got %d", s.WindowOpacity)
	}

	if s.Theme == nil {
		return fmt.Errorf("theme not loaded")
	}

	return nil
}

func (s *Styles) merge(defaults Styles) {
	if s.ThemeName == "" {
		s.ThemeName = defaults.ThemeName
		s.Theme = defaults.Theme
	}

	if s.WindowOpacity == 0 {
		s.WindowOpacity = defaults.WindowOpacity
	}
}
