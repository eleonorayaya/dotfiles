package dock

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
)

const hideDelay = "1000"

type App struct{}

func New() *App {
	return &App{}
}

func (a *App) Name() string {
	return "dock"
}

func (a *App) Sync(ctx *app.Context) error {
	current := func(key string) string {
		out, _ := exec.Command("defaults", "read", "com.apple.dock", key).Output()
		return strings.TrimSpace(string(out))
	}
	if current("autohide") == "1" && current("autohide-delay") == hideDelay {
		return nil
	}
	for _, args := range [][]string{
		{"write", "com.apple.dock", "autohide", "-bool", "true"},
		{"write", "com.apple.dock", "autohide-delay", "-float", hideDelay},
	} {
		if out, err := exec.Command("defaults", args...).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to configure dock: %w: %s", err, strings.TrimSpace(string(out)))
		}
	}
	if err := exec.Command("killall", "Dock").Run(); err != nil {
		return fmt.Errorf("failed to restart dock: %w", err)
	}
	return nil
}
