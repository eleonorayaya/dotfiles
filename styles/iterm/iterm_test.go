package iterm

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/styles"
)

var update = flag.Bool("update", false, "rewrite golden files")

func load(t *testing.T, name string) Scheme {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", name+".itermcolors"))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()
	s, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return s
}

func convert(t *testing.T, name string) styles.Theme {
	t.Helper()
	theme, err := Convert(load(t, name), Options{Name: Slug(name, "-")})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	return theme
}

func TestSourceMatchesGolden(t *testing.T) {
	for _, name := range []string{"blue_berry_pie", "solarized_light"} {
		t.Run(name, func(t *testing.T) {
			got, err := Source(convert(t, name))
			if err != nil {
				t.Fatalf("Source: %v", err)
			}
			golden := filepath.Join("testdata", name+".go.golden")
			if *update {
				if err := os.WriteFile(golden, got, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden (run with -update to create): %v", err)
			}
			if string(got) != string(want) {
				t.Errorf("generated source differs from %s; rerun with -update if intended\n%s", golden, got)
			}
		})
	}
}

func TestConvertDetectsThemeType(t *testing.T) {
	if got := convert(t, "blue_berry_pie").Type; got != "dark" {
		t.Errorf("Blue Berry Pie type = %q, want dark", got)
	}
	if got := convert(t, "solarized_light").Type; got != "light" {
		t.Errorf("Solarized Light type = %q, want light", got)
	}
}

func TestConvertKeepsTextTiersOrdered(t *testing.T) {
	for _, name := range []string{"blue_berry_pie", "solarized_light"} {
		c := convert(t, name).Colors
		bg := parseHex(c.Surface)
		ratio := func(h string) float64 { return Contrast(parseHex(h), bg) }

		if !(ratio(c.TextOnSurfaceEmphasis) > ratio(c.TextOnSurface) &&
			ratio(c.TextOnSurface) > ratio(c.TextOnSurfaceVariant) &&
			ratio(c.TextOnSurfaceVariant) > ratio(c.TextOnSurfaceMuted)) {
			t.Errorf("%s: text tiers out of order: emphasis %.2f, text %.2f, variant %.2f, muted %.2f",
				name, ratio(c.TextOnSurfaceEmphasis), ratio(c.TextOnSurface), ratio(c.TextOnSurfaceVariant), ratio(c.TextOnSurfaceMuted))
		}
		if ratio(c.Comment) >= ratio(c.TextOnSurface) {
			t.Errorf("%s: comment %.2f is louder than text %.2f", name, ratio(c.Comment), ratio(c.TextOnSurface))
		}
	}
}

func TestConvertDefaultsPrimaryToMagenta(t *testing.T) {
	theme := convert(t, "blue_berry_pie")
	if theme.Colors.Primary != theme.Colors.Palette.Magenta {
		t.Errorf("Primary = %s, want Palette.Magenta %s", theme.Colors.Primary, theme.Colors.Palette.Magenta)
	}
}

func TestConvertRejectsUnknownHue(t *testing.T) {
	if _, err := Convert(load(t, "blue_berry_pie"), Options{Primary: "orange"}); err == nil {
		t.Error("expected error for unknown hue")
	}
}

func TestParseRejectsIncompleteScheme(t *testing.T) {
	_, err := Parse(strings.NewReader(`<plist><dict><key>Background Color</key><dict></dict></dict></plist>`))
	if err == nil {
		t.Error("expected error for scheme missing ANSI colors")
	}
}

func TestVarName(t *testing.T) {
	cases := map[string]string{
		"blue-berry-pie":         "BlueBerryPie",
		"3024 Night":             "Theme3024Night",
		"Builtin Solarized+Dark": "BuiltinSolarizedDark",
		"iTerm2 Smoooooth":       "ITerm2Smoooooth",
	}
	for in, want := range cases {
		if got := VarName(in); got != want {
			t.Errorf("VarName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRegisterIsIdempotent(t *testing.T) {
	registry := []byte("package themes\n\nimport \"github.com/eleonorayaya/shizuku/styles\"\n\nvar All = []styles.Theme{\n\tMonade,\n}\n")
	once, err := Register(registry, "BlueBerryPie")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	twice, err := Register(once, "BlueBerryPie")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if string(once) != string(twice) || strings.Count(string(twice), "BlueBerryPie") != 1 {
		t.Errorf("Register not idempotent:\n%s", twice)
	}
}
