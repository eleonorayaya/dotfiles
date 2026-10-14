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
		Styles: styles.New(styles.WithTheme(themes.Sylveon), styles.WithWindowOpacity(80)),
	}

	result, err := New(Options{MailApp: "Proton Mail"}).Generate(ctx)
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
	if cfg.MailApp != "Proton Mail" {
		t.Errorf("MailApp = %q", cfg.MailApp)
	}
	if cfg.UtenaURL != defaultUtenaURL || !cfg.Bluetooth {
		t.Errorf("utena and bluetooth should be on by default: %+v", cfg)
	}
	if cfg.WorkspaceIcons["T"] != DefaultWorkspaceIcons["T"] {
		t.Errorf("default workspace icons missing: %v", cfg.WorkspaceIcons)
	}
	if !strings.HasPrefix(cfg.Bar.Color, "0xCC") {
		t.Errorf("bar color should carry 80%% opacity, got %q", cfg.Bar.Color)
	}
}

func TestGenerateDisablesOptionalItems(t *testing.T) {
	ctx := &app.Context{OutDir: t.TempDir(), Styles: styles.New(styles.WithTheme(themes.Sylveon))}
	result, err := New(Options{DisableUtena: true, DisableBT: true}).Generate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := bar.LoadConfig(result.FileMap[configName])
	if err != nil {
		t.Fatal(err)
	}
	if cfg.UtenaURL != "" || cfg.Bluetooth || cfg.MailApp != "" {
		t.Errorf("optional items should be off: %+v", cfg)
	}
}
