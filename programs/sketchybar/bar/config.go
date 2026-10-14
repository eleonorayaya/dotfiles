package bar

import (
	"encoding/json"
	"fmt"
	"os"
)

type BarStyle struct {
	Height       int    `json:"height"`
	Margin       int    `json:"margin"`
	YOffset      int    `json:"yOffset"`
	CornerRadius int    `json:"cornerRadius"`
	BorderWidth  int    `json:"borderWidth"`
	BlurRadius   int    `json:"blurRadius"`
	Color        string `json:"color"`
	BorderColor  string `json:"borderColor"`
}

type Colors struct {
	Icon            string `json:"icon"`
	IconHighlight   string `json:"iconHighlight"`
	Label           string `json:"label"`
	LabelHighlight  string `json:"labelHighlight"`
	PopupBackground string `json:"popupBackground"`
	PopupBorder     string `json:"popupBorder"`
}

type Fonts struct {
	Text string `json:"text"`
	Icon string `json:"icon"`
}

type Config struct {
	Bar            BarStyle          `json:"bar"`
	Colors         Colors            `json:"colors"`
	Fonts          Fonts             `json:"fonts"`
	WorkspaceIcons map[string]string `json:"workspaceIcons"`
	MailApp        string            `json:"mailApp,omitempty"`
	UtenaURL       string            `json:"utenaURL,omitempty"`
	Bluetooth      bool              `json:"bluetooth"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read bar config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse bar config: %w", err)
	}
	return cfg, nil
}
