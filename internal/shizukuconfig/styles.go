package shizukuconfig

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Gaps struct {
	InnerHorizontal int `yaml:"innerHorizontal"`
	InnerVertical   int `yaml:"innerVertical"`
	OuterLeft       int `yaml:"outerLeft"`
	OuterBottom     int `yaml:"outerBottom"`
	OuterTop        int `yaml:"outerTop"`
	OuterRight      int `yaml:"outerRight"`
}

type Styles struct {
	ThemeName     string `yaml:"theme"`
	WindowOpacity int    `yaml:"windowOpacity"`
	Gaps          Gaps   `yaml:"gaps"`
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

var (
	defaultThemeName     = "monade"
	defaultWindowOpacity = 85
	defaultGapsDesktop   = Gaps{
		InnerHorizontal: 16,
		InnerVertical:   16,
		OuterLeft:       64,
		OuterBottom:     128,
		OuterTop:        64,
		OuterRight:      64,
	}
	defaultGapsLaptop = Gaps{
		InnerHorizontal: 8,
		InnerVertical:   8,
		OuterLeft:       8,
		OuterBottom:     8,
		OuterTop:        8,
		OuterRight:      8,
	}
)

func createDefaultStyles() Styles {
	theme, _ := loadThemeFromRegistry(defaultThemeName)
	return Styles{
		ThemeName:     defaultThemeName,
		WindowOpacity: defaultWindowOpacity,
		Gaps:          defaultGapsDesktop,
		Theme:         theme,
	}
}

func DefaultGapsForLaptop(laptop bool) Gaps {
	if laptop {
		return defaultGapsLaptop
	}
	return defaultGapsDesktop
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

	if s.Gaps == (Gaps{}) {
		s.Gaps = defaults.Gaps
	}
}
