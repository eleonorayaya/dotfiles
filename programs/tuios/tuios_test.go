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

	result, err := New(Options{}).Generate(ctx)
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

	data, err := os.ReadFile(result.FileMap["config.toml"])
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	return string(data)
}

func TestGenerateConfigKeepsAfterAttachHookLocally(t *testing.T) {
	config := generateConfig(t, Options{})

	if !strings.Contains(config, "right = [\"notifications\", \"clock\"]\n\n[hooks]\nafter-attach = '''curl") {
		t.Error("expected the after-attach hook to follow the dock section")
	}
	if !strings.HasSuffix(config, "/mux/hooks/client-session-changed'''\n") {
		t.Errorf("unexpected config tail %q", config[len(config)-80:])
	}
}

func TestGenerateRemoteConfig(t *testing.T) {
	config := generateConfig(t, Options{Remote: true})

	if !strings.Contains(config, "[daemon]\nresume_agents = \"auto\"\n") {
		t.Error(`expected remote config to keep resume_agents = "auto"`)
	}
	for _, unwanted := range []string{"[hooks]", "localhost:3333", "&#", "&quot;"} {
		if strings.Contains(config, unwanted) {
			t.Errorf("expected remote config to not contain %s", unwanted)
		}
	}
	if !strings.HasSuffix(config, "right = [\"notifications\", \"clock\"]\n") {
		t.Errorf("unexpected remote config tail %q", config[len(config)-80:])
	}
}

func TestRemoteInstallIsNoop(t *testing.T) {
	if err := New(Options{Remote: true}).Install(&app.Context{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

const pickerCommand = "$HOME/.local/bin/utena picker --app-id com.example.term --title utena-picker -- /opt/term --title utena-picker -o a=b+c"

const pickerBinding = `[[keybindings.command]]
key = "prefix+p"
type = "shell"
command = '''` + pickerCommand + `'''
description = "utena sessions"
`

const popupBinding = `[[keybindings.command]]
key = "prefix+p"
type = "popup"
command = "UTENA_POPUP=1 $HOME/.local/bin/utena"
description = "utena sessions"
`

func TestGenerateConfigWithPickerUsesShellBinding(t *testing.T) {
	config := generateConfig(t, Options{Picker: pickerCommand})

	if !strings.Contains(config, "start_in_terminal_mode = true\n\n"+pickerBinding+"\n[[keybindings.command]]\nkey = \"prefix+n\"") {
		t.Errorf("expected the picker shell binding, got:\n%s", config)
	}
	for _, unwanted := range []string{`type = "popup"`, "UTENA_POPUP", "&#", "&quot;"} {
		if strings.Contains(config, unwanted) {
			t.Errorf("expected picker config to not contain %s", unwanted)
		}
	}
}

func TestGenerateConfigWithoutPickerKeepsPopup(t *testing.T) {
	config := generateConfig(t, Options{})

	if !strings.Contains(config, "start_in_terminal_mode = true\n\n"+popupBinding+"\n[[keybindings.command]]\nkey = \"prefix+n\"") {
		t.Errorf("expected the popup binding, got:\n%s", config)
	}
}

func TestGenerateRemoteConfigHasNoSessionPicker(t *testing.T) {
	for _, opts := range []Options{{Remote: true}, {Remote: true, Picker: pickerCommand}} {
		config := generateConfig(t, opts)

		if strings.Contains(config, "prefix+p") {
			t.Errorf("expected remote config to not bind prefix+p, got:\n%s", config)
		}
		if !strings.Contains(config, "start_in_terminal_mode = true\n\n[[keybindings.command]]\nkey = \"prefix+n\"") {
			t.Errorf("unexpected remote keybindings:\n%s", config)
		}
	}
}
