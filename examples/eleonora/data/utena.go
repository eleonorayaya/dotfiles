package data

import (
	"strings"

	"github.com/eleonorayaya/shizuku/programs/utena"
)

func Utena(browser string) utena.Options {
	return utena.Options{
		Browser:    browser,
		Bundles:    Bundles(),
		McpServers: McpServers(),
	}
}

func CoderUtena() utena.Options {
	opts := Utena("")
	opts.Follower = true
	return opts
}

const (
	pickerTitle = "utena-picker"
	kittyAppID  = "net.kovidgoyal.kitty"
	kittyBin    = "/Applications/kitty.app/Contents/MacOS/kitty"
)

func UtenaPicker() string {
	return strings.Join([]string{
		"$HOME/.local/bin/utena", "picker",
		"--app-id", kittyAppID,
		"--title", pickerTitle,
		"--",
		kittyBin,
		"--title", pickerTitle,
		"-o", "macos_quit_when_last_window_closed=yes",
		"-o", "remember_window_size=no",
		"-o", "initial_window_width=110c",
		"-o", "initial_window_height=32c",
	}, " ")
}
