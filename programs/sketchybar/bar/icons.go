package bar

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

const (
	symbolFont    = "SF Pro"
	appFontFamily = "sketchybar-app-font"
	appFont       = appFontFamily + ":Regular:16.0"
	nerdFont      = "Hack Nerd Font"

	iconWifi     = "\U00100647"
	iconOffline  = "\U00100648"
	iconEthernet = "\U00100906"
	iconHeart    = "♥"

	iconSittingCat = "\uEEED"

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

var nerdAppIcons = map[string]string{
	"kitty": iconSittingCat,
}

func nerdIcon(app string) (string, bool) {
	icon, ok := nerdAppIcons[strings.ToLower(app)]
	return icon, ok
}

func appIconProps(app string, size int) []string {
	if icon, ok := nerdIcon(app); ok {
		return []string{P("icon", icon), P("icon.font", fmt.Sprintf("%s:Regular:%d.0", nerdFont, size))}
	}
	return []string{P("icon", appIcon(app)), P("icon.font", fmt.Sprintf("%s:Regular:%d.0", appFontFamily, size))}
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
