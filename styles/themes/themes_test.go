package themes

import (
	"reflect"
	"testing"

	"github.com/eleonorayaya/shizuku/styles"
)

func emptyFields(prefix string, v reflect.Value) []string {
	var empty []string
	for i := range v.NumField() {
		name := prefix + v.Type().Field(i).Name
		switch f := v.Field(i); f.Kind() {
		case reflect.Struct:
			empty = append(empty, emptyFields(name+".", f)...)
		case reflect.String:
			if f.String() == "" {
				empty = append(empty, name)
			}
		}
	}
	return empty
}

func TestShippedThemesSetEveryColor(t *testing.T) {
	for _, theme := range []styles.Theme{Monade, MonadeDark} {
		t.Run(theme.Name, func(t *testing.T) {
			for _, field := range emptyFields("", reflect.ValueOf(theme.Colors)) {
				t.Errorf("%s is empty", field)
			}
		})
	}
}

func TestWithDefaultsFillsRolesFromPalette(t *testing.T) {
	theme := MonadeDark
	theme.Colors.String = ""
	theme.Colors.Accent = ""

	got := theme.WithDefaults().Colors
	if got.String != theme.Colors.Palette.Green {
		t.Errorf("String = %q, want Palette.Green %q", got.String, theme.Colors.Palette.Green)
	}
	if got.Accent != theme.Colors.Primary {
		t.Errorf("Accent = %q, want Primary %q", got.Accent, theme.Colors.Primary)
	}
	if got.Keyword != MonadeDark.Colors.Keyword {
		t.Errorf("Keyword override was replaced: got %q", got.Keyword)
	}
}
