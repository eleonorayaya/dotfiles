package data

import (
	"strings"
	"testing"
)

func TestUtenaPickerOpensKittyWithPickerTitle(t *testing.T) {
	want := "$HOME/.local/bin/utena picker --app-id net.kovidgoyal.kitty --title utena-picker -- " +
		"/Applications/kitty.app/Contents/MacOS/kitty --title utena-picker " +
		"-o macos_quit_when_last_window_closed=yes -o remember_window_size=no " +
		"-o initial_window_width=110c -o initial_window_height=32c"

	got := UtenaPicker()

	if got != want {
		t.Errorf("UtenaPicker() = %q, want %q", got, want)
	}
	if strings.Contains(got, "'''") {
		t.Error("picker command must not contain ''' because it is a TOML literal string")
	}
}
