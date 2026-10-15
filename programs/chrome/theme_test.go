package chrome

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func TestThemeManifestMapsColors(t *testing.T) {
	theme := themes.Sylveon.WithDefaults()
	manifest, err := themeManifest(theme, themeImages{patternImage: []byte("img")})
	if err != nil {
		t.Fatal(err)
	}

	if manifest["manifest_version"] != 3 {
		t.Errorf("expected manifest v3, got %v", manifest["manifest_version"])
	}
	body := manifest["theme"].(map[string]any)
	colors := body["colors"].(map[string][3]int)
	if got, want := colors["toolbar"], [3]int{0xfb, 0xee, 0xf6}; got != want {
		t.Errorf("toolbar = %v, want %v", got, want)
	}
	if got, want := colors["toolbar_button_icon"], [3]int{0xa8, 0x08, 0x48}; got != want {
		t.Errorf("toolbar_button_icon = %v, want %v", got, want)
	}
	if ntp := body["images"].(map[string]any)["theme_ntp_background"]; ntp != patternImage {
		t.Errorf("unexpected new tab image %v", ntp)
	}
	if body["properties"].(map[string]string)["ntp_background_repeat"] != "repeat" {
		t.Error("expected the pattern to tile")
	}
}

func TestThemeManifestVersionTracksContent(t *testing.T) {
	theme := themes.Sylveon.WithDefaults()
	a, _ := themeManifest(theme, themeImages{patternImage: []byte("one")})
	b, _ := themeManifest(theme, themeImages{patternImage: []byte("one")})
	c, _ := themeManifest(theme, themeImages{patternImage: []byte("two")})

	if a["version"] != b["version"] {
		t.Error("expected identical input to keep the version")
	}
	if a["version"] == c["version"] {
		t.Error("expected a changed image to bump the version")
	}
}

func TestThemeManifestSkipsEmptyColors(t *testing.T) {
	manifest, err := themeManifest(styles.Theme{Name: "x", Colors: styles.ThemeColors{Surface: "#101010"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	body := manifest["theme"].(map[string]any)
	if len(body["colors"].(map[string][3]int)) != 2 {
		t.Errorf("expected only toolbar and ntp_background, got %v", body["colors"])
	}
	if _, ok := body["images"]; ok {
		t.Error("expected no images without a pattern")
	}
}

func TestDotTileDrawsCenteredDot(t *testing.T) {
	data, err := dotTile("#fbeef6", "#eebccc", 3)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dx(); got != dotSpacing*3 {
		t.Errorf("tile width = %d, want %d", got, dotSpacing*3)
	}
	r, g, b, _ := img.At(dotSpacing+dotSpacing/2, 2*dotSpacing+dotSpacing/2).RGBA()
	if r>>8 != 0xee || g>>8 != 0xbc || b>>8 != 0xcc {
		t.Errorf("center = %x%x%x, want the dot color", r>>8, g>>8, b>>8)
	}
	r, g, b, _ = img.At(0, 0).RGBA()
	if r>>8 != 0xfb || g>>8 != 0xee || b>>8 != 0xf6 {
		t.Errorf("corner = %x%x%x, want the background", r>>8, g>>8, b>>8)
	}
}

func TestPatternImagesShippedThemes(t *testing.T) {
	for _, theme := range themes.All {
		images, err := patternImages(theme.WithDefaults().Colors)
		if err != nil {
			t.Errorf("%s: %v", theme.Name, err)
			continue
		}
		if _, err := themeManifest(theme.WithDefaults(), images); err != nil {
			t.Errorf("%s: %v", theme.Name, err)
		}
	}
}
