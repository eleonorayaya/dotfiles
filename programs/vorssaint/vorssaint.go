package vorssaint

import (
	"fmt"
	"os/exec"
	"slices"
	"strconv"
	"strings"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/util"
)

const (
	domain               = "com.vorssaint.utils"
	appName              = "Vorssaint"
	onboardingFeatureSet = 4
)

var allFeatures = []string{
	"switcher", "dockPreview", "dockClick", "windowMaximizer", "windowLayout", "autoQuit", "spacesOrder",
	"scrollInverter", "scrollHorizontal", "focusFollowsMouse", "smoothScroll", "linearScroll", "mouseAcceleration",
	"mouseNavigation", "mouseButtonShortcuts", "middleClick", "mouseClickDebounce", "keyboardDebounce",
	"textSnippets", "superKey", "quitWindowProtection",
	"clipboardHistory", "pastePlain", "finderCutPaste", "finderRename", "shelf", "urlCleaner", "diskImageInstaller",
	"mixer", "soundOutputSwitcher", "audioPriority", "micMute", "musicBlock",
	"keepAwake", "brightness", "extraBrightness", "bluetoothSleep",
	"quickLauncher", "quickToggles", "colorPicker", "screenOCR", "cleaningMode", "mediaTools",
	"cleaner", "uninstaller", "homebrew", "appUpdates", "screenshot", "cameraPreview", "radialMenu", "scratchpad",
	"commandBar", "screenRecorder", "wallpaper", "killProcess", "portManager",
	"notch", "notchCalendar", "notchNotifications", "notchGestures", "notchTimer", "notchAccessories", "notchLyrics",
	"notchQueue", "notchLiveEqualizer", "notchDownloads", "notchAgents", "notchWatch", "notchMascot",
	"monitorCPU", "monitorGPU", "monitorMemory", "monitorNetwork", "monitorDisk", "monitorPower",
	"connectedDevices", "fanControl",
}

type Options struct {
	Features           []string
	CommandBarShortcut string
}

type App struct {
	opts Options
}

func New(opts Options) *App {
	return &App{opts: opts}
}

func (a *App) Name() string {
	return "vorssaint"
}

func (a *App) Install(ctx *app.Context) error {
	if err := util.InstallBrewPackage("vorssaint", true); err != nil {
		return fmt.Errorf("failed to install vorssaint: %w", err)
	}

	return nil
}

type setting struct {
	key   string
	kind  string
	value string
}

func (s setting) writeArgs() []string {
	value := s.value
	if s.kind == "-bool" {
		value = strconv.FormatBool(value == "1")
	}
	return []string{"write", domain, s.key, s.kind, value}
}

func boolSetting(key string, v bool) setting {
	if v {
		return setting{key: key, kind: "-bool", value: "1"}
	}
	return setting{key: key, kind: "-bool", value: "0"}
}

func (a *App) settings() ([]setting, error) {
	for _, f := range a.opts.Features {
		if !slices.Contains(allFeatures, f) {
			return nil, fmt.Errorf("unknown vorssaint feature %q", f)
		}
	}

	settings := []setting{
		boolSetting("hasOnboarded", true),
		{key: "featuresOnboardingVersion", kind: "-int", value: strconv.Itoa(onboardingFeatureSet)},
		boolSetting("launchAtLoginWanted", true),
	}
	for _, f := range allFeatures {
		settings = append(settings, boolSetting("featureAvailable."+f, slices.Contains(a.opts.Features, f)))
	}
	if a.opts.CommandBarShortcut != "" {
		settings = append(settings,
			boolSetting("commandBarShortcutEnabled", true),
			setting{key: "commandBarShortcut", kind: "-string", value: a.opts.CommandBarShortcut},
		)
	}

	return settings, nil
}

func (a *App) Sync(ctx *app.Context) error {
	settings, err := a.settings()
	if err != nil {
		return err
	}

	var stale []setting
	for _, s := range settings {
		out, err := exec.Command("defaults", "read", domain, s.key).Output()
		if err != nil || strings.TrimSpace(string(out)) != s.value {
			stale = append(stale, s)
		}
	}
	if len(stale) == 0 {
		return nil
	}

	_ = exec.Command("osascript", "-e", fmt.Sprintf("quit app %q", appName)).Run()

	for _, s := range stale {
		if out, err := exec.Command("defaults", s.writeArgs()...).CombinedOutput(); err != nil {
			return fmt.Errorf("failed to write vorssaint setting %s: %w: %s", s.key, err, strings.TrimSpace(string(out)))
		}
	}

	if out, err := exec.Command("open", "-a", appName).CombinedOutput(); err != nil {
		return fmt.Errorf("failed to launch vorssaint: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return nil
}
