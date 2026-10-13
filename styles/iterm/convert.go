package iterm

import (
	"fmt"
	"math"
	"strings"

	"github.com/eleonorayaya/shizuku/styles"
)

var Hues = map[string]int{
	"red":     1,
	"green":   2,
	"yellow":  3,
	"blue":    4,
	"magenta": 5,
	"cyan":    6,
}

type Options struct {
	Name      string
	Primary   string
	Secondary string
	Tertiary  string
}

func (o Options) withDefaults() Options {
	if o.Primary == "" {
		o.Primary = "magenta"
	}
	if o.Secondary == "" {
		o.Secondary = "blue"
	}
	if o.Tertiary == "" {
		o.Tertiary = "cyan"
	}
	return o
}

func hue(s Scheme, name string) (RGB, RGB, error) {
	i, ok := Hues[strings.ToLower(name)]
	if !ok {
		return RGB{}, RGB{}, fmt.Errorf("unknown hue %q", name)
	}
	return s.Ansi(i), s.Ansi(i + 8), nil
}

func variant(base, bright, toward RGB) RGB {
	bh, bs, _ := hsl(base)
	vh, vs, _ := hsl(bright)
	d := math.Abs(bh - vh)
	if d > 180 {
		d = 360 - d
	}
	if bs > 0.15 && vs > 0.15 && d < 30 && bright != base {
		return bright
	}
	return Mix(base, toward, 0.25)
}

func hsl(c RGB) (h, s, l float64) {
	maxC := math.Max(c.R, math.Max(c.G, c.B))
	minC := math.Min(c.R, math.Min(c.G, c.B))
	l = (maxC + minC) / 2
	d := maxC - minC
	if d == 0 {
		return 0, 0, l
	}
	s = d / (1 - math.Abs(2*l-1))
	switch maxC {
	case c.R:
		h = math.Mod((c.G-c.B)/d, 6)
	case c.G:
		h = (c.B-c.R)/d + 2
	default:
		h = (c.R-c.G)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return h, s, l
}

func Convert(s Scheme, opts Options) (styles.Theme, error) {
	opts = opts.withDefaults()

	bg := s["Background Color"]
	fg := s["Foreground Color"]
	dark := luminance(bg) < luminance(fg)
	extreme := RGB{}
	themeType := "light"
	if dark {
		extreme = RGB{1, 1, 1}
		themeType = "dark"
	}

	emphasis := Mix(fg, extreme, 0.5)
	muted := Mix(fg, bg, 0.45)
	onFill := func(fill RGB) string { return mostReadable(fill, bg, emphasis).Hex() }

	primary, primaryBright, err := hue(s, opts.Primary)
	if err != nil {
		return styles.Theme{}, fmt.Errorf("invalid primary: %w", err)
	}
	secondary, _, err := hue(s, opts.Secondary)
	if err != nil {
		return styles.Theme{}, fmt.Errorf("invalid secondary: %w", err)
	}
	tertiary, tertiaryBright, err := hue(s, opts.Tertiary)
	if err != nil {
		return styles.Theme{}, fmt.Errorf("invalid tertiary: %w", err)
	}
	primaryVariant := variant(primary, primaryBright, emphasis)
	tertiaryVariant := variant(tertiary, tertiaryBright, emphasis)

	comment := s.Ansi(8)
	if c := Contrast(comment, bg); c < 3 || c >= Contrast(fg, bg) {
		comment = Mix(fg, bg, 0.3)
	}

	link := s.Get("Link Color", s.Ansi(4))

	return styles.Theme{
		Name: opts.Name,
		Type: themeType,
		Colors: styles.ThemeColors{
			Palette: styles.Palette{
				Black:         s.Ansi(0).Hex(),
				Red:           s.Ansi(1).Hex(),
				Green:         s.Ansi(2).Hex(),
				Yellow:        s.Ansi(3).Hex(),
				Blue:          s.Ansi(4).Hex(),
				Magenta:       s.Ansi(5).Hex(),
				Cyan:          s.Ansi(6).Hex(),
				White:         s.Ansi(7).Hex(),
				BrightBlack:   s.Ansi(8).Hex(),
				BrightRed:     s.Ansi(9).Hex(),
				BrightGreen:   s.Ansi(10).Hex(),
				BrightYellow:  s.Ansi(11).Hex(),
				BrightBlue:    s.Ansi(12).Hex(),
				BrightMagenta: s.Ansi(13).Hex(),
				BrightCyan:    s.Ansi(14).Hex(),
				BrightWhite:   s.Ansi(15).Hex(),
			},
			Surface:               bg.Hex(),
			SurfaceVariant:        Mix(bg, fg, 0.06).Hex(),
			SurfaceHighlight:      Mix(bg, fg, 0.10).Hex(),
			SurfaceBorder:         Mix(bg, fg, 0.20).Hex(),
			TextOnSurface:         fg.Hex(),
			TextOnSurfaceVariant:  Mix(fg, bg, 0.20).Hex(),
			TextOnSurfaceMuted:    muted.Hex(),
			TextOnSurfaceEmphasis: emphasis.Hex(),
			Primary:               primary.Hex(),
			PrimaryVariant:        primaryVariant.Hex(),
			TextOnPrimary:         onFill(primary),
			Secondary:             secondary.Hex(),
			TextOnSecondary:       onFill(secondary),
			Tertiary:              tertiary.Hex(),
			TertiaryVariant:       tertiaryVariant.Hex(),
			TextOnTertiary:        onFill(tertiary),
			Error:                 s.Ansi(1).Hex(),
			TextOnError:           onFill(s.Ansi(1)),
			Warning:               s.Ansi(3).Hex(),
			TextOnWarning:         onFill(s.Ansi(3)),
			Success:               s.Ansi(2).Hex(),
			TextOnSuccess:         onFill(s.Ansi(2)),
			Info:                  s.Ansi(4).Hex(),
			TextOnInfo:            onFill(s.Ansi(4)),
			Selection:             s.Get("Selection Color", Mix(bg, fg, 0.25)).Hex(),
			SelectionForeground:   s.Get("Selected Text Color", fg).Hex(),
			Cursor:                s.Get("Cursor Color", fg).Hex(),
			CursorText:            s.Get("Cursor Text Color", bg).Hex(),
			Link:                  link.Hex(),
			LinkHover:             Mix(link, emphasis, 0.3).Hex(),
			Comment:               comment.Hex(),
		},
	}, nil
}

type Warning struct {
	Pair     string
	Contrast float64
	Min      float64
}

func Check(t styles.Theme) []Warning {
	c := t.WithDefaults().Colors
	p := c.Palette
	pairs := []struct {
		name   string
		fg, bg string
		min    float64
	}{
		{"text on surface", c.TextOnSurface, c.Surface, 4.5},
		{"emphasis on surface", c.TextOnSurfaceEmphasis, c.Surface, 4.5},
		{"variant text on surface", c.TextOnSurfaceVariant, c.Surface, 3},
		{"muted text on surface", c.TextOnSurfaceMuted, c.Surface, 2},
		{"comment on surface", c.Comment, c.Surface, 3},
		{"selection text", c.SelectionForeground, c.Selection, 3},
		{"cursor text", c.CursorText, c.Cursor, 3},
		{"text on primary", c.TextOnPrimary, c.Primary, 3},
		{"text on secondary", c.TextOnSecondary, c.Secondary, 3},
		{"text on tertiary", c.TextOnTertiary, c.Tertiary, 3},
		{"text on error", c.TextOnError, c.Error, 3},
		{"text on warning", c.TextOnWarning, c.Warning, 3},
		{"primary on surface", c.Primary, c.Surface, 3},
		{"link on surface", c.Link, c.Surface, 3},
		{"red on surface", p.Red, c.Surface, 3},
		{"green on surface", p.Green, c.Surface, 3},
		{"yellow on surface", p.Yellow, c.Surface, 3},
		{"blue on surface", p.Blue, c.Surface, 3},
		{"magenta on surface", p.Magenta, c.Surface, 3},
		{"cyan on surface", p.Cyan, c.Surface, 3},
	}

	var warnings []Warning
	for _, pair := range pairs {
		ratio := Contrast(parseHex(pair.fg), parseHex(pair.bg))
		if ratio < pair.min {
			warnings = append(warnings, Warning{Pair: pair.name, Contrast: ratio, Min: pair.min})
		}
	}
	return warnings
}

func parseHex(h string) RGB {
	var r, g, b int
	if _, err := fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return RGB{}
	}
	return RGB{float64(r) / 255, float64(g) / 255, float64(b) / 255}
}
