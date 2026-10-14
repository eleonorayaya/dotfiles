package sketchybar

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/eleonorayaya/shizuku/programs/sketchybar/bar"
	"github.com/eleonorayaya/shizuku/util"
)

const (
	bottomLabel   = "sh.shizuku.bottombar"
	dockHideDelay = "1000"
)

func syncBottomBar() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to find home dir: %w", err)
	}
	sketchybarPath, err := exec.LookPath("sketchybar")
	if err != nil {
		return fmt.Errorf("sketchybar not installed: %w", err)
	}

	link := filepath.Join(home, ".local", "bin", bar.BottomBinary)
	if err := util.EnsureDirExists(filepath.Dir(link)); err != nil {
		return err
	}
	if current, err := os.Readlink(link); err != nil || current != sketchybarPath {
		os.Remove(link)
		if err := os.Symlink(sketchybarPath, link); err != nil {
			return fmt.Errorf("failed to link %s: %w", link, err)
		}
	}

	plistPath := filepath.Join(home, "Library", "LaunchAgents", bottomLabel+".plist")
	if err := util.EnsureDirExists(filepath.Dir(plistPath)); err != nil {
		return err
	}
	plist, err := bottomPlist(home, link)
	if err != nil {
		return err
	}
	if err := os.WriteFile(plistPath, plist, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", plistPath, err)
	}

	domain := "gui/" + strconv.Itoa(os.Getuid())
	exec.Command("launchctl", "bootout", domain+"/"+bottomLabel).Run()
	if out, err := exec.Command("launchctl", "bootstrap", domain, plistPath).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start bottom bar: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func bottomPlist(home, binary string) ([]byte, error) {
	path := strings.Join([]string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "go", "bin"),
		"/opt/homebrew/bin", "/usr/local/bin", "/usr/bin", "/bin", "/usr/sbin", "/sbin",
	}, ":")
	esc := func(s string) string {
		var b bytes.Buffer
		xml.EscapeText(&b, []byte(s))
		return b.String()
	}
	rc := filepath.Join(home, ".config", "sketchybar", "bottombarrc")
	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
		<string>--config</string>
		<string>%s</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		<key>PATH</key>
		<string>%s</string>
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>ProcessType</key>
	<string>Interactive</string>
</dict>
</plist>
`, bottomLabel, esc(binary), esc(rc), esc(path))), nil
}

func hideDock() {
	current := func(key string) string {
		out, _ := exec.Command("defaults", "read", "com.apple.dock", key).Output()
		return strings.TrimSpace(string(out))
	}
	if current("autohide") == "1" && current("autohide-delay") == dockHideDelay {
		return
	}
	for _, args := range [][]string{
		{"write", "com.apple.dock", "autohide", "-bool", "true"},
		{"write", "com.apple.dock", "autohide-delay", "-float", dockHideDelay},
	} {
		if out, err := exec.Command("defaults", args...).CombinedOutput(); err != nil {
			slog.Warn("failed to configure dock", "error", err, "output", string(out))
			return
		}
	}
	if err := exec.Command("killall", "Dock").Run(); err != nil {
		slog.Warn("failed to restart dock", "error", err)
	}
}
