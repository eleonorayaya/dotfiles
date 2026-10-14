package bar

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

const (
	symbolFont = "SF Pro"
	appFont    = "sketchybar-app-font:Regular:16.0"

	iconWifi     = "\U00100647"
	iconOffline  = "\U00100648"
	iconEthernet = "\U00100906"
	iconHeart    = "♥"

	appIconDefault   = ":default:"
	appIconBluetooth = ":bluetooth:"
)

//go:embed app_icons.json
var appIconsJSON []byte

var appIcons = sync.OnceValue(func() map[string]string {
	var entries []struct {
		IconName string   `json:"iconName"`
		AppNames []string `json:"appNames"`
	}
	if err := json.Unmarshal(appIconsJSON, &entries); err != nil {
		panic("invalid app_icons.json: " + err.Error())
	}
	icons := map[string]string{}
	for _, e := range entries {
		for _, name := range e.AppNames {
			icons[strings.ToLower(name)] = e.IconName
		}
	}
	return icons
})

var appIconOverrides = map[string]string{
	"vesktop": ":discord:",
}

func appIcon(app string) string {
	if icon, ok := appIconOverrides[strings.ToLower(app)]; ok {
		return icon
	}
	if icon, ok := appIcons()[strings.ToLower(app)]; ok {
		return icon
	}
	return appIconDefault
}
