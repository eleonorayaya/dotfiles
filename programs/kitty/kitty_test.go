package kitty

import (
	"os"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func generateConfig(t *testing.T, opts Options) string {
	t.Helper()
	ctx := &app.Context{
		OutDir: t.TempDir(),
		Styles: styles.New(styles.WithTheme(themes.Sylveon)),
	}

	result, err := New(opts).Generate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(result.FileMap["kitty.conf"])
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	return string(data)
}

func TestGenerateOpensURLsWithSystemDefault(t *testing.T) {
	config := generateConfig(t, Options{})
	if !strings.Contains(config, "open_url_with  default\n") {
		t.Error("expected open_url_with to default to the system opener")
	}
}

func TestGenerateOpensURLsWithOverride(t *testing.T) {
	config := generateConfig(t, Options{OpenURLWith: "~/bin/opener --flag"})
	if !strings.Contains(config, "open_url_with  ~/bin/opener --flag\n") {
		t.Error("expected open_url_with to use the override")
	}
	if strings.Contains(config, "&#") || strings.Contains(config, "&quot;") {
		t.Error("expected config to be free of HTML escapes")
	}
}
