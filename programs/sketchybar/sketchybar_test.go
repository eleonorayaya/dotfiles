package sketchybar

import (
	"os"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/programs/sketchybar/bar"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func TestGenerateWritesBarConfig(t *testing.T) {
	ctx := &app.Context{
		OutDir: t.TempDir(),
		Styles: styles.New(styles.WithTheme(themes.Sylveon), styles.WithWindowOpacity(80), styles.WithGaps(styles.Gaps{OuterLeft: 64}), styles.WithGapOverride("built-in retina", styles.Gaps{OuterLeft: 8})),
	}

	result, err := New(Options{Badges: []string{"Proton Mail", "Discord"}}).Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rc, err := os.ReadFile(result.FileMap["sketchybarrc"])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rc), `exec shizuku bar run --config "$HOME/.config/sketchybar/bar.json"`) {
		t.Fatalf("sketchybarrc should exec the daemon:\n%s", rc)
	}

	cfg, err := bar.LoadConfig(result.FileMap[configName])
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Badges) != 2 || cfg.Badges[1] != "Discord" {
		t.Errorf("Badges = %v", cfg.Badges)
	}
	if cfg.Margin != 64 || cfg.MarginOverride == nil || cfg.MarginOverride.Margin != 8 {
		t.Errorf("margins should follow window gaps: %d %+v", cfg.Margin, cfg.MarginOverride)
	}
	if !strings.HasPrefix(cfg.Colors.Island, "0xCC") {
		t.Errorf("bar color should carry 80%% opacity, got %q", cfg.Colors.Island)
	}
}
