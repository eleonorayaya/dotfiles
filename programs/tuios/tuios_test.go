package tuios

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func isolateConfigDir(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("failed to resolve user config dir: %v", err)
	}
	dir := filepath.Join(configDir, "tuios")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}
	return dir
}

func TestGenerateConfigUsesThemeColors(t *testing.T) {
	isolateConfigDir(t)
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
	isolateConfigDir(t)
	return generateConfigInPlace(t, opts)
}

func generateConfigInPlace(t *testing.T, opts Options) string {
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
	if !strings.HasPrefix(config, "[appearance]\npreferred_shell = '/usr/bin/zsh'\nhide_window_buttons = true\n") {
		t.Errorf("expected remote config to prefer zsh, got:\n%s", config[:120])
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

func TestGenerateLocalConfigLeavesShellToSHELL(t *testing.T) {
	if config := generateConfig(t, Options{}); strings.Contains(config, "preferred_shell") {
		t.Error("expected the local config to not set preferred_shell")
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

func writeExistingConfig(t *testing.T, contents string) {
	t.Helper()
	dir := isolateConfigDir(t)
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}
}

func TestGenerateConfigKeepsExistingHosts(t *testing.T) {
	writeExistingConfig(t, `[appearance]
theme = "stale"

[hosts.build]
address = "build.example"
ssh_options = ["-A"]

[hosts."*"]
connect_timeout = "10s"
`)

	config := generateConfigInPlace(t, Options{})

	for _, want := range []string{
		"[hosts.build]\naddress = \"build.example\"\nssh_options = [\"-A\"]\n",
		"[hosts.\"*\"]\nconnect_timeout = \"10s\"\n",
		"/mux/hooks/client-session-changed'''\n\n[hosts]\n",
	} {
		if !strings.Contains(config, want) {
			t.Errorf("expected config to contain %q, got:\n%s", want, config)
		}
	}
	if strings.Contains(config, `theme = "stale"`) {
		t.Error("expected only the hosts table to be kept")
	}
}

func TestGenerateConfigWithoutExistingHostsAddsNone(t *testing.T) {
	writeExistingConfig(t, "[appearance]\ntheme = \"stale\"\n")

	if config := generateConfigInPlace(t, Options{}); strings.Contains(config, "[hosts") {
		t.Errorf("expected no hosts table, got:\n%s", config)
	}
}

func TestGenerateConfigRejectsUnparsableExistingConfig(t *testing.T) {
	writeExistingConfig(t, "[hosts\n")

	ctx := &app.Context{OutDir: t.TempDir(), Styles: styles.New(styles.WithTheme(themes.Sylveon))}
	if _, err := New(Options{}).Generate(ctx); err == nil {
		t.Error("expected an unparsable config to stop the sync rather than drop its hosts")
	}
}
