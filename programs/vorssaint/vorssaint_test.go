package vorssaint

import (
	"testing"
)

func TestSettingsInstallsOnlyListedFeatures(t *testing.T) {
	a := New(Options{Features: []string{"commandBar"}, CommandBarShortcut: "command:49"})

	settings, err := a.settings()
	if err != nil {
		t.Fatalf("settings: %v", err)
	}

	got := map[string]string{}
	for _, s := range settings {
		got[s.key] = s.value
	}

	for _, f := range allFeatures {
		want := "0"
		if f == "commandBar" {
			want = "1"
		}
		if got["featureAvailable."+f] != want {
			t.Errorf("featureAvailable.%s = %q, want %q", f, got["featureAvailable."+f], want)
		}
	}
	if got["hasOnboarded"] != "1" {
		t.Errorf("hasOnboarded = %q, want 1", got["hasOnboarded"])
	}
	if got["commandBarShortcut"] != "command:49" || got["commandBarShortcutEnabled"] != "1" {
		t.Errorf("command bar shortcut = %q enabled %q", got["commandBarShortcut"], got["commandBarShortcutEnabled"])
	}
}

func TestWriteArgsUseDefaultsBoolSyntax(t *testing.T) {
	for _, tc := range []struct {
		setting setting
		want    string
	}{
		{boolSetting("hasOnboarded", true), "true"},
		{boolSetting("featureAvailable.mixer", false), "false"},
		{setting{key: "featuresOnboardingVersion", kind: "-int", value: "4"}, "4"},
	} {
		args := tc.setting.writeArgs()
		if got := args[len(args)-1]; got != tc.want {
			t.Errorf("%s writes %q, want %q", tc.setting.key, got, tc.want)
		}
	}
}

func TestSettingsRejectsUnknownFeature(t *testing.T) {
	if _, err := New(Options{Features: []string{"commandbar"}}).settings(); err == nil {
		t.Fatal("expected error for unknown feature")
	}
}
