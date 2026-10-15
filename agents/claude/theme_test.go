package claude

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/eleonorayaya/shizuku/app"
	"github.com/eleonorayaya/shizuku/styles"
	"github.com/eleonorayaya/shizuku/styles/themes"
)

func TestClaudeThemeMapsColors(t *testing.T) {
	theme := styles.Theme{Name: "test", Type: "light", Colors: styles.ThemeColors{
		Primary: "#a80848",
		Surface: "#ffffff",
		Added:   "#000000",
	}}

	ct := claudeTheme(theme)

	if ct["base"] != "light" {
		t.Errorf("expected light base, got %v", ct["base"])
	}
	if ct["name"] != "test" {
		t.Errorf("expected theme name, got %v", ct["name"])
	}
	overrides := ct["overrides"].(map[string]any)
	if overrides["claude"] != "#a80848" {
		t.Errorf("expected claude from primary, got %v", overrides["claude"])
	}
	if overrides["diffAdded"] != "#b3b3b3" {
		t.Errorf("expected diffAdded blended from surface, got %v", overrides["diffAdded"])
	}
	if _, ok := overrides["permission"]; ok {
		t.Error("expected empty colors to be omitted")
	}
}

func TestClaudeThemeDefaultsToDarkBase(t *testing.T) {
	if base := claudeTheme(styles.Theme{Type: "dark"})["base"]; base != "dark" {
		t.Errorf("expected dark base, got %v", base)
	}
}

func TestClaudeThemeShippedThemesHaveValidColors(t *testing.T) {
	for _, theme := range themes.All {
		overrides := claudeTheme(theme.WithDefaults())["overrides"].(map[string]any)
		for token, color := range overrides {
			if s, _ := color.(string); len(s) != 7 || s[0] != '#' {
				t.Errorf("%s: token %s has invalid color %v", theme.Name, token, color)
			}
		}
		if len(overrides) < 40 {
			t.Errorf("%s: expected most tokens set, got %d", theme.Name, len(overrides))
		}
	}
}

func TestMergeSettingsSelectsCustomTheme(t *testing.T) {
	a := New(testOptions())

	result, err := a.mergeSettings(t.TempDir(), styles.Theme{Name: "sylveon"}, app.AgentContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(result)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("failed to parse output: %v", err)
	}
	if settings["theme"] != "custom:shizuku" {
		t.Errorf("expected custom theme selected, got %v", settings["theme"])
	}
}

func TestClaudeThemeWarningUsesTertiaryVariant(t *testing.T) {
	overrides := claudeTheme(styles.Theme{Colors: styles.ThemeColors{TertiaryVariant: "#8a4cc0", Warning: "#a85a14"}})["overrides"].(map[string]any)
	if overrides["warning"] != "#8a4cc0" {
		t.Errorf("expected warning from tertiary variant, got %v", overrides["warning"])
	}

	fallback := claudeTheme(styles.Theme{Colors: styles.ThemeColors{Warning: "#a85a14"}})["overrides"].(map[string]any)
	if fallback["warning"] != "#a85a14" {
		t.Errorf("expected warning to fall back to the warning role, got %v", fallback["warning"])
	}
}
