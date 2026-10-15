package chrome

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/util"
)

const (
	dotSpacing   = 22
	dotRadius    = 1.3
	dotSamples   = 4
	patternImage = "images/dots.png"
	patternHiDPI = "images/dots@2x.png"
)

func themeColors(c styles.ThemeColors) map[string]string {
	return map[string]string{
		"frame":                        c.SurfaceHighlight,
		"frame_inactive":               c.SurfaceVariant,
		"toolbar":                      c.Surface,
		"toolbar_text":                 c.TextOnSurface,
		"toolbar_button_icon":          c.Primary,
		"tab_text":                     c.TextOnSurface,
		"tab_background_text":          c.TextOnSurfaceVariant,
		"tab_background_text_inactive": c.TextOnSurfaceMuted,
		"bookmark_text":                c.TextOnSurface,
		"omnibox_background":           c.SurfaceVariant,
		"omnibox_text":                 c.TextOnSurface,
		"ntp_background":               c.Surface,
		"ntp_text":                     c.TextOnSurface,
		"ntp_link":                     c.Link,
	}
}

type themeImages map[string][]byte

func patternImages(c styles.ThemeColors) (themeImages, error) {
	if c.Surface == "" || c.SurfaceBorder == "" {
		return nil, nil
	}
	images := themeImages{}
	for path, scale := range map[string]int{patternImage: 1, patternHiDPI: 2} {
		data, err := dotTile(c.Surface, c.SurfaceBorder, scale)
		if err != nil {
			return nil, err
		}
		images[path] = data
	}
	return images, nil
}

func dotTile(background, dot string, scale int) ([]byte, error) {
	bg, err := util.HexToRGB(background)
	if err != nil {
		return nil, fmt.Errorf("failed to convert background: %w", err)
	}
	fg, err := util.HexToRGB(dot)
	if err != nil {
		return nil, fmt.Errorf("failed to convert dot: %w", err)
	}

	size := dotSpacing * scale
	center := float64(size) / 2
	radius := dotRadius * float64(scale)
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			covered := 0
			for sy := range dotSamples {
				for sx := range dotSamples {
					px := float64(x) + (float64(sx)+0.5)/dotSamples - center
					py := float64(y) + (float64(sy)+0.5)/dotSamples - center
					if math.Hypot(px, py) <= radius {
						covered++
					}
				}
			}
			px := util.MixRGB(bg, fg, float64(covered)/(dotSamples*dotSamples))
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(px[0]), G: uint8(px[1]), B: uint8(px[2]), A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode dot tile: %w", err)
	}
	return buf.Bytes(), nil
}

func themeManifest(theme styles.Theme, images themeImages) (map[string]any, error) {
	colors := map[string][3]int{}
	for key, hex := range themeColors(theme.Colors) {
		if hex == "" {
			continue
		}
		rgb, err := util.HexToRGB(hex)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s: %w", key, err)
		}
		colors[key] = rgb
	}

	body := map[string]any{"colors": colors}
	if len(images) > 0 {
		body["images"] = map[string]any{
			"theme_ntp_background": map[string]string{
				"100_percent": patternImage,
				"200_percent": patternHiDPI,
			},
		}
		body["properties"] = map[string]string{
			"ntp_background_alignment": "top",
			"ntp_background_repeat":    "repeat",
		}
	}

	version, err := contentVersion(body, images)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"manifest_version": 3,
		"name":             "Shizuku " + theme.Name,
		"version":          version,
		"theme":            body,
	}, nil
}

func contentVersion(body map[string]any, images themeImages) (string, error) {
	data, err := json.Marshal(map[string]any{"theme": body, "images": images})
	if err != nil {
		return "", fmt.Errorf("failed to marshal theme: %w", err)
	}
	h := fnv.New32a()
	h.Write(data)
	sum := h.Sum32()
	return fmt.Sprintf("1.%d.%d", sum>>16, sum&0xffff), nil
}
