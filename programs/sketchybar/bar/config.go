package bar

import (
	"encoding/json"
	"fmt"
	"os"
)

type Colors struct {
	Bar             string `json:"bar"`
	BarBorder       string `json:"barBorder"`
	Icon            string `json:"icon"`
	IconHighlight   string `json:"iconHighlight"`
	Label           string `json:"label"`
	LabelHighlight  string `json:"labelHighlight"`
	PopupBackground string `json:"popupBackground"`
	PopupBorder     string `json:"popupBorder"`
}

type Config struct {
	Colors  Colors `json:"colors"`
	MailApp string `json:"mailApp,omitempty"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read bar config: %w", err)
	}
	return decodeJSON[Config](data, "bar config")
}

func decodeJSON[T any](data []byte, what string) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, fmt.Errorf("failed to parse %s: %w", what, err)
	}
	return v, nil
}
