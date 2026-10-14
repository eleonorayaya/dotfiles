package tuios

import (
	"os"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func TestGenerateConfigUsesThemeColors(t *testing.T) {
	ctx := &app.Context{
		OutDir: t.TempDir(),
		Styles: styles.New(styles.WithTheme(themes.Sylveon)),
	}

	result, err := New().Generate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(result.FileMap["config.toml"])
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	config := string(data)

	c := ctx.Styles.Theme.Colors
	for _, want := range []string{
		`border_focused_color = "` + c.Primary + `"`,
		`bg = "` + c.Selection + `"`,
		`cursor_bg = "` + c.Cursor + `"`,
		`-d "{\"session_name\":\"$TUIOS_SESSION_ID\"}"`,
	} {
		if !strings.Contains(config, want) {
			t.Errorf("expected config to contain %s", want)
		}
	}
	if strings.Contains(config, "&#") || strings.Contains(config, "&quot;") {
		t.Error("expected config to be free of HTML escapes")
	}
}
