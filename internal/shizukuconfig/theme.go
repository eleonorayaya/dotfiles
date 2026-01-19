package shizukuconfig

import (
	"github.com/eleonorayaya/shizuku/internal/theme"
	_ "github.com/eleonorayaya/shizuku/internal/theme/monade"
)

type Theme = theme.Theme
type ThemeColors = theme.ThemeColors

func loadTheme(themeName string) (*Theme, error) {
	return theme.LoadTheme(themeName)
}
