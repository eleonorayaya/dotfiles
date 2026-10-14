package vesktop

import (
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func TestDiscordThemeMapsColors(t *testing.T) {
	theme := themes.Sylveon.WithDefaults()
	css := discordTheme(theme, 1)

	for _, want := range []string{
		"@name Shizuku",
		".visual-refresh.theme-light",
		"--brand-500: " + theme.Colors.Primary + " !important;",
		"--text-default: " + theme.Colors.TextOnSurface + " !important;",
		"--background-base-low: " + theme.Colors.Surface + " !important;",
		"--brand-50a: " + theme.Colors.Primary + "80 !important;",
		"tooltipGrey_",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("expected css to contain %q", want)
		}
	}
}

func TestDiscordThemeTranslucent(t *testing.T) {
	theme := themes.Sylveon.WithDefaults()

	opaque := discordTheme(theme, 1)
	if strings.Contains(opaque, "background: transparent") {
		t.Error("expected opaque theme to keep the page background")
	}

	css := discordTheme(theme, 0.8)
	if !strings.Contains(css, "background: transparent !important;") {
		t.Error("expected translucent theme to clear the page background")
	}
	if !strings.Contains(css, "--app-frame-background: #fefcfdcc !important;") {
		t.Error("expected the light frame to be a whitened surface at window opacity")
	}
	if !strings.Contains(css, "--chat-background: transparent !important;") {
		t.Error("expected panels over the frame to be transparent")
	}
	if !strings.Contains(css, "--modal-background: "+theme.Colors.Surface+" !important;") {
		t.Error("expected modals to stay opaque")
	}
}

func TestDiscordThemeDarkSkipsLightFlips(t *testing.T) {
	css := discordTheme(styles.Theme{Type: "dark", Colors: styles.ThemeColors{Surface: "#101010", Primary: "#ff0000"}}, 1)

	if strings.Contains(css, "tooltipGrey_") || strings.Contains(css, "--white:") {
		t.Error("expected light-only overrides to be omitted for dark themes")
	}
	if strings.Contains(css, ": !important") {
		t.Error("expected empty colors to be omitted")
	}
}

func TestDiscordThemeShippedThemesHaveValidColors(t *testing.T) {
	for _, theme := range themes.All {
		for _, line := range strings.Split(discordTheme(theme.WithDefaults(), 0.8), "\n") {
			if !strings.HasPrefix(line, "  --") {
				continue
			}
			value := strings.TrimSuffix(strings.SplitN(line, ": ", 2)[1], " !important;")
			if value != "transparent" && (value[0] != '#' || (len(value) != 7 && len(value) != 9)) {
				t.Errorf("%s: invalid color in %q", theme.Name, line)
			}
		}
	}
}
