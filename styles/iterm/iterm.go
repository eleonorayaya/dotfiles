package iterm

import (
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
)

type RGB struct {
	R, G, B float64
}

func (c RGB) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", channel(c.R), channel(c.G), channel(c.B))
}

func channel(v float64) int {
	return int(math.Round(math.Max(0, math.Min(1, v)) * 255))
}

type Scheme map[string]RGB

func Parse(r io.Reader) (Scheme, error) {
	dec := xml.NewDecoder(r)
	scheme := Scheme{}
	depth := 0
	var key, component string
	var color RGB
	inColor := false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read plist: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "dict":
				depth++
				if depth == 2 {
					inColor = true
					color = RGB{}
				}
			case "key", "real", "integer", "string":
				var text string
				if err := dec.DecodeElement(&text, &t); err != nil {
					return nil, fmt.Errorf("failed to read plist %s: %w", t.Name.Local, err)
				}
				text = strings.TrimSpace(text)
				switch {
				case t.Name.Local == "key" && depth == 1:
					key = text
				case t.Name.Local == "key" && inColor:
					component = text
				case inColor && (t.Name.Local == "real" || t.Name.Local == "integer"):
					v, err := strconv.ParseFloat(text, 64)
					if err != nil {
						return nil, fmt.Errorf("invalid %s %s: %w", key, component, err)
					}
					switch component {
					case "Red Component":
						color.R = v
					case "Green Component":
						color.G = v
					case "Blue Component":
						color.B = v
					}
				}
			}
		case xml.EndElement:
			if t.Name.Local == "dict" {
				if depth == 2 && inColor {
					scheme[key] = color
					inColor = false
				}
				depth--
			}
		}
	}

	for i := range 16 {
		if _, ok := scheme[ansiKey(i)]; !ok {
			return nil, fmt.Errorf("scheme is missing %q", ansiKey(i))
		}
	}
	for _, k := range []string{"Background Color", "Foreground Color"} {
		if _, ok := scheme[k]; !ok {
			return nil, fmt.Errorf("scheme is missing %q", k)
		}
	}
	return scheme, nil
}

func ansiKey(i int) string {
	return fmt.Sprintf("Ansi %d Color", i)
}

func (s Scheme) Ansi(i int) RGB {
	return s[ansiKey(i)]
}

func (s Scheme) Get(key string, fallback RGB) RGB {
	if c, ok := s[key]; ok {
		return c
	}
	return fallback
}

func Mix(a, b RGB, t float64) RGB {
	return RGB{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
	}
}

func luminance(c RGB) float64 {
	lin := func(v float64) float64 {
		v = math.Max(0, math.Min(1, v))
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

func Contrast(a, b RGB) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func mostReadable(on RGB, candidates ...RGB) RGB {
	best := candidates[0]
	for _, c := range candidates[1:] {
		if Contrast(c, on) > Contrast(best, on) {
			best = c
		}
	}
	return best
}
