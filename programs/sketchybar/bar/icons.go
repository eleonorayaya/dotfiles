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

	iconMail     = "\U00100355"
	iconWifi     = "\U00100647"
	iconOffline  = "\U00100648"
	iconEthernet = "\U00100906"
	iconCPU      = "\U00100AE5"
	iconMemory   = "\U00100AE6"
	iconSession  = "\U00100A8F"

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

func appIcon(app string) string {
	if icon, ok := appIcons()[strings.ToLower(app)]; ok {
		return icon
	}
	return appIconDefault
}
