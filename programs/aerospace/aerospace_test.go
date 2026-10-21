package aerospace

import (
	"os"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
)

func TestGenerateFloatsUtenaPickerBeforeKittyRule(t *testing.T) {
	ctx := &app.Context{OutDir: t.TempDir(), Styles: styles.New()}

	result, err := New().Generate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(result.FileMap["aerospace.toml"])
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	config := string(data)

	picker := "[[on-window-detected]]\nif.app-id = 'net.kovidgoyal.kitty'\nif.window-title-regex-substring = '^utena-picker$'\nrun = [ 'layout floating' ]\n"
	kitty := "[[on-window-detected]]\nif.app-id = 'net.kovidgoyal.kitty'\nrun = [ 'move-node-to-workspace T --focus-follows-window' ]\n"
	pi, ki := strings.Index(config, picker), strings.Index(config, kitty)
	if pi < 0 {
		t.Fatalf("expected the utena picker rule, got:\n%s", config)
	}
	if ki < 0 || pi > ki {
		t.Error("expected the utena picker rule before the kitty rule, because the first matching rule wins")
	}
}
